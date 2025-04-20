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

// linearCount estimates the number of unique characters in the text using a probabilistic counting algorithm.
// Parameters:
// - text: input text as a slice of int32.
// - tmp: temporary array for bit storage.
// Returns an estimate of the number of unique characters.
func linearCount(text, tmp []int32) uint64 {
	n := len(text)
	totalBits := uint64(n * 32)

	var buf [4]byte
	h := fnv.New64a()

	// Use FNV hash to map characters to bit positions in tmp.
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(buf[:], uint32(text[i]))
		h.Reset()
		h.Write(buf[:])
		x := h.Sum64()
		bitIndex := x % totalBits
		slot := bitIndex / 32
		bit := uint32(bitIndex % 32)
		tmp[slot] |= int32(1 << bit)
	}

	// Count zero bits to estimate unique characters.
	zeroBits := 0
	for i := 0; i < n; i++ {
		val := uint32(tmp[i])
		zeroBits += bits.OnesCount32(^val)
		tmp[i] = 0
	}

	if zeroBits == 0 {
		return totalBits
	}
	// Apply linear counting formula: -N * ln(Z/N), where Z is zero bits.
	estimate := -float64(totalBits) * math.Log(float64(zeroBits)/float64(totalBits))
	return uint64(estimate + 0.5)
}

// makeBucketsMap creates a map of buckets for characters in the text.
// Parameters:
// - sa: suffix array used temporarily to store the alphabet.
// - text: input text.
// Returns:
// - bucketsMap: map of character to bucket (start, end, size).
// - alphaSize: size of the alphabet.
func makeBucketsMap(sa, text []int32) (map[int32]bucket, int32) {
	// Estimate alphabet size using linear counting.
	lc := int(linearCount(text, sa))
	// Add 10% error.
	sz := lc + int(float32(lc)*0.1)
	bucketsMap := make(map[int32]bucket, sz)
	var alphaSize int32
	// Count character frequencies and collect unique characters.
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
	// Sort unique characters to define bucket order.
	alphabet := sa[:alphaSize]
	slices.Sort(alphabet)
	var (
		offset, n int32
		curr      bucket
	)
	// Assign start and end positions to buckets based on character frequencies.
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

// induceSort_arb constructs a suffix array for arbitrary alphabets using induced sorting.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - data: auxiliary array (unused in this version).
// - numLMS: number of LMS (Left-Most S-type) suffixes.
// Returns the constructed suffix array.
func induceSort_arb(text, sa, data []int32, numLMS int32) []int32 {
	// Create bucket map for character sorting.
	bucketsMap, alphaSize := makeBucketsMap(sa, text)
	var summary []int32

	// Place LMS suffixes into their buckets.
	insertLMS_arb(text, sa, bucketsMap)
	if numLMS > 1 {
		// Induce L-type and S-type suffixes for the summary array.
		induceSubL_arb(text, sa, bucketsMap)
		induceSubS_arb(text, sa, bucketsMap)
		// Extract LMS substring indices for summary string.
		summary = sa[len(sa)-int(numLMS):]
		maxName := summarise(text, sa, summary, numLMS)

		summarySA := sa[:numLMS]
		if maxName < numLMS {
			// Recursively construct suffix array for summary string if unique LMS substrings exist.
			_sais(summary, summarySA, data, alphaSize)
			// Map summary indices back to original text.
			unmap(text, sa, summarySA, summary)
		} else {
			// Copy summary directly if all LMS substrings are unique.
			copy(summarySA, summary)
			clear(sa[numLMS:])
		}
		// Expand LMS suffixes to their correct positions.
		expand_arb(text, sa, summarySA, bucketsMap)
	}
	// Final induction steps to complete the suffix array.
	induceL_arb(text, sa, bucketsMap)
	induceS_arb(text, sa, bucketsMap)
	return sa
}

// bucketStart_arb updates the start positions of buckets for L-type suffixes.
// Parameters:
// - buckets: map of character to bucket.
func bucketStart_arb(buckets map[int32]bucket) {
	// Reset start positions to the beginning of each bucket for L-type sorting.
	for ch, b := range buckets {
		b.start = b.end - b.size + 1
		buckets[ch] = b
	}
}

// bucketEnd_arb updates the end positions of buckets for S-type suffixes.
// Parameters:
// - buckets: map of character to bucket.
func bucketEnd_arb(buckets map[int32]bucket) {
	// Reset end positions to the end of each bucket for S-type sorting.
	for ch, b := range buckets {
		b.end = b.start + b.size - 1
		buckets[ch] = b
	}
}

// expand_arb places LMS suffixes into the suffix array using bucket sorting for arbitrary alphabets.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - summarySA: suffix array of the summary string.
// - buckets: map of character to bucket.
func expand_arb(text, sa, summarySA []int32, buckets map[int32]bucket) {
	var (
		b         bucket
		lmsIdx, j int32
	)
	// Place LMS suffixes into their bucket ends in reverse order.
	for i := len(summarySA) - 1; i >= 0; i-- {
		lmsIdx = summarySA[i]
		summarySA[i] = 0
		j = text[lmsIdx]
		b = buckets[j]
		sa[b.end] = lmsIdx
		b.end--
		buckets[j] = b
	}
	// Restore bucket end positions for subsequent operations.
	bucketEnd_arb(buckets)
}

// insertLMS_arb inserts LMS suffixes into the suffix array for arbitrary alphabets.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - buckets: map of character to bucket.
func insertLMS_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		b                bucket
		l, r, i, lastLMS int32
		numLMS           int
		S                bool
	)

	// Scan text backwards to identify LMS positions and insert them into buckets.
	for i = int32(len(text) - 1); i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			// Insert LMS suffix at the end of its character's bucket.
			b = buckets[r]
			sa[b.end] = i + 1
			lastLMS = b.end
			numLMS++
			b.end--
			buckets[r] = b
		}
	}
	// Mark the last LMS position as empty if multiple LMS suffixes exist.
	if numLMS > 1 {
		sa[lastLMS] = 0
	}
	// Restore bucket end positions.
	bucketEnd_arb(buckets)
}

// induceSubL_arb induces L-type suffixes for the summary suffix array with arbitrary alphabets.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - buckets: map of character to bucket.
func induceSubL_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		k, j     int32  = int32(len(text) - 1), 0
		l, r     int32  = text[k-1], text[k]
		lastChar int32  = text[len(text)-1]
		b        bucket = buckets[lastChar]
	)
	// Initialize with the last character, marking it as L-type or S-type.
	if l < r {
		k = -k
	}
	// Place suffix at the start of its character's bucket.
	sa[b.start] = int32(k)
	if b.size > 1 {
		b.start++
		buckets[lastChar] = b
	}

	// Induce L-type suffixes by scanning the suffix array.
	for i := 0; i < len(sa); i++ {
		if sa[i] == 0 {
			continue
		}
		j = sa[i]
		if j < 0 {
			// Restore negative (already processed) suffix.
			sa[i] = -j
			continue
		}
		sa[i] = 0
		k = j - 1
		l, r = text[k-1], text[k]
		// Mark as L-type (negative) if the preceding character forms an L-type suffix.
		if l < r {
			k = -k
		}
		// Place suffix at the start of its character's bucket.
		b = buckets[r]
		sa[b.start] = k
		b.start++
		buckets[r] = b
	}
	// Reset bucket start positions for L-type sorting.
	bucketStart_arb(buckets)
}

// induceSubS_arb induces S-type suffixes for the summary suffix array with arbitrary alphabets.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - buckets: map of character to bucket.
func induceSubS_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		b          bucket
		j, l, r, k int32
		top        = len(sa)
	)
	// Scan suffix array backwards to induce S-type suffixes.
	for i := len(sa) - 1; i >= 0; i-- {
		j = sa[i]
		if j == 0 {
			continue
		}
		sa[i] = 0
		if j < 0 {
			// Move negative (processed) suffix to the top of the array.
			top--
			sa[top] = -j
			continue
		}
		k = j - 1
		l, r = text[k-1], text[k]

		// Mark as S-type (negative) if the preceding character forms an S-type suffix.
		if l > r {
			k = -k
		}
		// Place suffix at the end of its character's bucket.
		b = buckets[r]
		sa[b.end] = k
		b.end--
		buckets[r] = b
	}
	// Reset bucket end positions for S-type sorting.
	bucketEnd_arb(buckets)
}

// induceL_arb induces L-type suffixes for the final suffix array with arbitrary alphabets.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - buckets: map of character to bucket.
func induceL_arb(text, sa []int32, buckets map[int32]bucket) {
	var (
		k, j     int32  = int32(len(text) - 1), 0
		l, r     int32  = text[k-1], text[k]
		lastChar int32  = text[len(text)-1]
		b        bucket = buckets[lastChar]
	)
	// Initialize with the last character, marking it as L-type or S-type.
	if l < r {
		k = -k
	}
	// Place suffix at the start of its character's bucket.
	sa[b.start] = int32(k)
	b.start++
	buckets[lastChar] = b

	// Induce L-type suffixes by scanning the suffix array.
	for i := 0; i < len(sa); i++ {
		j = sa[i]
		if j <= 0 {
			continue
		}

		k = j - 1
		r = text[k]
		if k > 0 {
			// Check if the preceding character forms an L-type suffix.
			if l = text[k-1]; l < r {
				k = -k
			}
		}
		// Place suffix at the start of its character's bucket.
		b = buckets[r]
		sa[b.start] = k
		b.start++
		buckets[r] = b
	}
	// Reset bucket start positions for L-type sorting.
	bucketStart_arb(buckets)
}

// induceS_arb induces S-type suffixes for the final suffix array with arbitrary alphabets.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - buckets: map of character to bucket.
func induceS_arb(text, sa []int32, buckets map[int32]bucket) {
	// Scan suffix array backwards to induce S-type suffixes.
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
			// Check if the preceding character forms an S-type suffix.
			if l := text[k-1]; l <= r {
				k = -k
			}
		}
		// Place suffix at the end of its character's bucket.
		b := buckets[r]
		sa[b.end] = k
		b.end--
		buckets[r] = b
	}
	// Reset bucket end positions for S-type sorting.
	bucketEnd_arb(buckets)
}
