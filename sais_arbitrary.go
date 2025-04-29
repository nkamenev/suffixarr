// Copyright (c) 2025 Nikita Kamenev
// Licensed under the MIT License. See LICENSE file in the project root for details.
package suffixarr

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/bits"
	"slices"
)

// bucket represents a bucket for sorting characters in the SA-IS algorithm.
type bucket struct {
	start, end, size int32
}

// linearCount estimates the number of unique characters using probabilistic linear counting.
// It employs a hash-based approach to map characters to bit positions in a temporary array,
// then calculates the number of unset bits to estimate the alphabet size using a logarithmic formula.
// This method is efficient for large texts with potentially sparse character sets.
func linearCount(text, tmp []int32) uint64 {
	n := len(text)
	totalBits := uint64(n * 32)

	var buf [4]byte
	h := fnv.New64a()

	// Convert each character to a 32-bit integer and hash it using FNV to a bit position.
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(buf[:], uint32(text[i]))
		h.Reset()
		h.Write(buf[:])
		x := h.Sum64()
		// Map hash to a bit index within the temporary array.
		bitIndex := x % totalBits
		slot := bitIndex / 32
		bit := uint32(bitIndex % 32)
		tmp[slot] |= int32(1 << bit) // Set the bit to mark presence.
	}

	// Count unset bits to estimate unique characters.
	zeroBits := 0
	for i := 0; i < n; i++ {
		val := uint32(tmp[i])
		zeroBits += bits.OnesCount32(^val)
		tmp[i] = 0 // Clear the array for reuse.
	}

	if zeroBits == 0 {
		return totalBits // All bits are set, assume maximum estimate.
	}
	// Apply linear counting formula: -N * ln(Z/N), where Z is the number of zero bits.
	estimate := -float64(totalBits) * math.Log(float64(zeroBits)/float64(totalBits))
	return uint64(estimate + 0.5)
}

// makeBucketsMap creates a map of buckets for character sorting in the SA-IS algorithm.
// It estimates the alphabet size, counts character frequencies, sorts unique characters
// to define bucket order, and assigns start and end positions for each bucket based on
// cumulative frequencies.
func makeBucketsMap(sa, text []int32) (map[int32]bucket, int32) {
	// Estimate alphabet size using linear counting, adding a 10% margin for potential errors.
	lc := int(linearCount(text, sa))
	sz := lc + int(float32(lc)*0.1)
	bucketsMap := make(map[int32]bucket, sz)
	var alphaSize int32
	// Scan text to collect unique characters and count their frequencies.
	for i := 0; i < len(text); i++ {
		curr := text[i]
		bkt, exists := bucketsMap[curr]
		if !exists {
			sa[alphaSize] = curr
			alphaSize++
		}
		bkt.size++
		bucketsMap[curr] = bkt
	}
	// Sort unique characters to establish a consistent bucket order.
	alphabet := sa[:alphaSize]
	slices.Sort(alphabet)
	var (
		offset, n int32
		curr      bucket
	)
	// Assign start and end positions for each bucket based on cumulative frequencies.
	for i := 0; i < len(alphabet); i++ {
		n, alphabet[i] = alphabet[i], 0
		curr = bucketsMap[n]
		curr.start = offset
		offset += curr.size
		curr.end = offset - 1
		bucketsMap[n] = curr
	}
	return bucketsMap, alphaSize
}

// induceSort_arb constructs the suffix array for arbitrary alphabets using induced sorting.
// It maps characters to buckets, inserts LMS suffixes, induces L- and S-type suffixes for a summary array,
// recursively processes the summary string if LMS substrings are not unique, and completes the suffix array
// with final L- and S-type inductions. This approach efficiently handles large or sparse character sets.
func induceSort_arb(text, sa, data []int32, numLMS int32) []int32 {
	// Build bucket map for character sorting.
	bucketsMap, alphaSize := makeBucketsMap(sa, text)
	var summary []int32

	// Insert LMS suffixes into their bucket ends.
	insertLMS_arb(text, sa, bucketsMap)
	if numLMS > 1 {
		// Induce L-type suffixes to prepare summary array.
		induceSubL_arb(text, sa, bucketsMap)
		// Induce S-type suffixes to complete summary array.
		induceSubS_arb(text, sa, bucketsMap)
		// Extract LMS substring indices for summary string.
		summary = sa[len(sa)-int(numLMS):]
		maxName := summarise(text, sa, summary, numLMS)

		summarySA := sa[:numLMS]
		if maxName < numLMS {
			// Recursively construct suffix array for summary string if LMS substrings repeat.
			_sais(summary, summarySA, data, alphaSize)
			// Map summary indices back to original text positions.
			unmap(text, sa, summarySA, summary)
		} else {
			// Use summary directly if all LMS substrings are unique.
			copy(summarySA, summary)
			clear(sa[numLMS:])
		}
		// Expand LMS suffixes to their final bucket positions.
		expand_arb(text, sa, summarySA, bucketsMap)
	}
	// Complete suffix array with final L-type induction.
	induceL_arb(text, sa, bucketsMap)
	// Complete suffix array with final S-type induction.
	induceS_arb(text, sa, bucketsMap)
	return sa
}

// bucketStart_arb updates start positions of buckets for L-type sorting.
func bucketStart_arb(buckets map[int32]bucket) {
	// Reset start to beginning of each bucket.
	for ch, b := range buckets {
		b.start = b.end - b.size + 1
		buckets[ch] = b
	}
}

// bucketEnd_arb updates end positions of buckets for S-type sorting.
func bucketEnd_arb(buckets map[int32]bucket) {
	// Reset end to end of each bucket.
	for ch, b := range buckets {
		b.end = b.start + b.size - 1
		buckets[ch] = b
	}
}

// expand_arb places LMS suffixes into final positions using bucket sorting for arbitrary alphabets.
// It uses the summary suffix array to insert LMS suffixes at the ends of their character buckets,
// updating bucket positions to prepare for final induction steps.
func expand_arb(text, sa, summarySA []int32, buckets map[int32]bucket) {
	var (
		b         bucket
		lmsIdx, j int32
	)
	// Scan summary array backwards to place LMS suffixes.
	for i := len(summarySA) - 1; i >= 0; i-- {
		lmsIdx = summarySA[i]
		summarySA[i] = 0
		j = text[lmsIdx]
		b = buckets[j]
		sa[b.end] = lmsIdx
		b.end--
		buckets[j] = b
	}
	// Restore bucket end positions for subsequent steps.
	bucketEnd_arb(buckets)
}

// insertLMS_arb inserts LMS suffixes into the suffix array for arbitrary alphabets.
// It scans the text backwards to identify LMS positions, places them at the ends of their
// character buckets, and marks the last LMS position as empty if multiple LMS suffixes exist.
func insertLMS_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		b                bucket
		l, r, i, lastLMS int32
		numLMS           int
		S                bool
	)
	// Scan text backwards to detect LMS positions.
	for i = int32(len(text) - 1); i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			// Place LMS suffix at bucket end.
			b = buckets[r]
			sa[b.end] = i + 1
			lastLMS = b.end
			numLMS++
			b.end--
			buckets[r] = b
		}
	}
	// Clear last LMS position if multiple exist.
	if numLMS > 1 {
		sa[lastLMS] = 0
	}
	// Restore bucket end positions.
	bucketEnd_arb(buckets)
}

// induceSubL_arb induces L-type suffixes for the summary suffix array with arbitrary alphabets.
// It initializes with the last character, scans forward to place L-type suffixes at the start
// of their character buckets, and marks processed suffixes as negative to avoid reprocessing.
func induceSubL_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		k, j     int32  = int32(len(text) - 1), 0
		l, r     int32  = text[k-1], text[k]
		lastChar int32  = text[len(text)-1]
		b        bucket = buckets[lastChar]
	)
	// Initialize with last character, marking L-type or S-type.
	if l < r {
		k = -k
	}
	// Place suffix at bucket start.
	sa[b.start] = int32(k)
	if b.size > 1 {
		b.start++
		buckets[lastChar] = b
	}

	// Scan forward to induce L-type suffixes.
	for i := 0; i < len(sa); i++ {
		if sa[i] == 0 {
			continue
		}
		j = sa[i]
		if j < 0 {
			// Restore processed suffix.
			sa[i] = -j
			continue
		}
		sa[i] = 0
		k = j - 1
		l, r = text[k-1], text[k]
		// Mark L-type suffix as negative.
		if l < r {
			k = -k
		}
		b = buckets[r]
		sa[b.start] = k
		b.start++
		buckets[r] = b
	}
	// Restore bucket start positions.
	bucketStart_arb(buckets)
}

// induceSubS_arb induces S-type suffixes for the summary suffix array with arbitrary alphabets.
// It scans backward, placing S-type suffixes at the ends of their character buckets,
// moving processed suffixes to the top of the array and marking them as negative.
func induceSubS_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		b          bucket
		j, l, r, k int32
		top        = len(sa)
	)
	// Scan backward to induce S-type suffixes.
	for i := len(sa) - 1; i >= 0; i-- {
		j = sa[i]
		if j == 0 {
			continue
		}
		sa[i] = 0
		if j < 0 {
			// Move processed suffix to top.
			top--
			sa[top] = -j
			continue
		}
		k = j - 1
		l, r = text[k-1], text[k]
		// Mark S-type suffix as negative.
		if l > r {
			k = -k
		}
		b = buckets[r]
		sa[b.end] = k
		b.end--
		buckets[r] = b
	}
	// Restore bucket end positions.
	bucketEnd_arb(buckets)
}

// induceL_arb induces L-type suffixes for the final suffix array with arbitrary alphabets.
// It initializes with the last character, scans forward to place L-type suffixes at the start
// of their character buckets, and updates bucket positions to complete the suffix array.
func induceL_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		k, j     int32  = int32(len(text) - 1), 0
		l, r     int32  = text[k-1], text[k]
		lastChar int32  = text[len(text)-1]
		b        bucket = buckets[lastChar]
	)
	// Initialize with last character, marking L-type or S-type.
	if l < r {
		k = -k
	}
	// Place suffix at bucket start.
	sa[b.start] = int32(k)
	b.start++
	buckets[lastChar] = b

	// Scan forward to induce L-type suffixes.
	for i := 0; i < len(sa); i++ {
		j = sa[i]
		if j <= 0 {
			continue
		}
		k = j - 1
		r = text[k]
		if k > 0 {
			// Check for L-type suffix.
			if l = text[k-1]; l < r {
				k = -k
			}
		}
		b = buckets[r]
		sa[b.start] = k
		b.start++
		buckets[r] = b
	}
	// Restore bucket start positions.
	bucketStart_arb(buckets)
}

// induceS_arb induces S-type suffixes for the final suffix array with arbitrary alphabets.
// It scans backward, restores processed suffixes, and places S-type suffixes at the ends
// of their character buckets to finalize the suffix array construction.
func induceS_arb(text, sa []int32, buckets map[int32]bucket) {
	// Scan backward to induce S-type suffixes.
	for i := len(sa) - 1; i >= 0; i-- {
		j := sa[i]
		if j >= 0 {
			continue
		}
		j = -j
		sa[i] = j
		k := j - 1
		r := text[k]
		if k > 0 {
			// Check for S-type suffix.
			if l := text[k-1]; l <= r {
				k = -k
			}
		}
		b := buckets[r]
		sa[b.end] = k
		b.end--
		buckets[r] = b
	}
	// Restore bucket end positions.
	bucketEnd_arb(buckets)
}
