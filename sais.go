// Copyright (c) 2025 Nikita Kamenev
// Licensed under the MIT License. See LICENSE file in the project root for details.
package suffixarr

// sais constructs a suffix array for the given text using the SA-IS algorithm.
func sais(text []int32) []int32 {
	if len(text) == 0 {
		return []int32{} // Empty text has no suffixes.
	} else if len(text) == 1 {
		return []int32{0} // Single character text has one suffix at index 0.
	}
	return _sais(text, nil, nil, 0)
}

// _sais is the core recursive implementation of the SA-IS algorithm.
// It analyzes the text to determine character range and LMS (Left-Most S-type) suffixes,
// then delegates to induced sorting based on alphabet size. For large alphabets, it uses
// arbitrary alphabet sorting; otherwise, it optimizes for small alphabets (<= 256).
// srcAlphaSize specifies the original alphabet size for recursive calls.
func _sais(text, sa, data []int32, srcAlphaSize int32) []int32 {
	var (
		minChar, maxChar int32 = text[0], text[0]
		l, r, numLMS     int32
		S                bool
	)
	// Scan text backwards to find min/max characters and count LMS suffixes.
	for i := len(text) - 1; i >= 0; i-- {
		l, r = text[i], l
		if l < minChar {
			minChar = l
		}
		if l > maxChar {
			maxChar = l
		}
		// Identify S-type (l < r) and LMS (l > r after S-type) positions.
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			numLMS++
		}
	}
	// Compute current alphabet size from character range.
	currAlphaSize := maxChar - minChar + 1
	if sa == nil {
		// Allocate suffix array if not provided.
		srcAlphaSize = currAlphaSize
		sa = make([]int32, len(text))
	}
	// Switch to arbitrary alphabet sorting for large or recursive alphabets.
	if currAlphaSize > 256 || currAlphaSize > srcAlphaSize {
		return induceSort_arb(text, sa, data, numLMS)
	}
	return induceSort(text, sa, data, minChar, numLMS, srcAlphaSize, currAlphaSize)
}

// induceSort builds the suffix array using induced sorting for small alphabets (<= 256).
// It places LMS suffixes into buckets, induces L-type and S-type suffixes, and recursively
// processes a summary string if LMS substrings are not unique. Final L-type and S-type
// induction completes the suffix array.
// Parameters include minChar (minimum character), numLMS (number of LMS suffixes),
// srcAlphaSize (original alphabet size), and currAlphaSize (current alphabet size).
func induceSort(text, sa, data []int32, minChar, numLMS, srcAlphaSize, currAlphaSize int32) []int32 {
	// Allocate or reuse auxiliary array for frequency and buckets.
	if data == nil || len(data) < int(srcAlphaSize)*2 {
		data = make([]int32, srcAlphaSize*2)
	}
	var summary []int32
	freq := data[:currAlphaSize]
	buckets := data[srcAlphaSize : srcAlphaSize+currAlphaSize]
	frequency(text, freq, minChar)

	// Insert LMS suffixes into their bucket ends.
	insertLMS(text, sa, freq, buckets, minChar)
	if numLMS > 1 {
		// Induce L-type suffixes for summary array.
		induceSubL(text, sa, freq, buckets, minChar)
		// Induce S-type suffixes for summary array.
		induceSubS(text, sa, freq, buckets, minChar)
		// Extract LMS substring indices for summary string.
		summary = sa[len(sa)-int(numLMS):]
		maxName := summarise(text, sa, summary, numLMS)

		summarySA := sa[:numLMS]
		if maxName < numLMS {
			// Recursively build suffix array for summary string if LMS substrings repeat.
			_sais(summary, summarySA, data, srcAlphaSize)
			// Map summary indices back to original text positions.
			unmap(text, sa, summarySA, summary)
		} else {
			// Use summary directly if all LMS substrings are unique.
			copy(summarySA, summary)
			clear(sa[numLMS:])
		}
		// Expand LMS suffixes to final positions using bucket sorting.
		expand(text, sa, summarySA, freq, buckets, minChar)
	}
	// Final induction to complete L-type suffixes.
	induceL(text, sa, freq, buckets, minChar)
	// Final induction to complete S-type suffixes.
	induceS(text, sa, freq, buckets, minChar)
	return sa
}

// unmap maps LMS substring indices from the summary suffix array back to their original
// positions in the text. It collects LMS positions in reverse order and reassigns them
// based on the summary suffix array to prepare for expansion.
func unmap(text, sa, summarySA, LMS []int32) {
	var (
		j    int32 = int32(len(LMS))
		l, r int32
		S    bool
	)
	// Scan text backwards to collect LMS positions.
	for i := len(text) - 1; i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			j--
			LMS[j] = int32(i) + 1 // Store LMS position.
		}
	}
	// Map summary indices to original LMS positions.
	for i := 0; i < len(LMS); i++ {
		j = summarySA[i]
		sa[i] = LMS[j]
		LMS[j] = 0 // Clear temporary storage.
	}
}

// expand places LMS suffixes into their final positions in the suffix array using bucket
// sorting. It uses the summary suffix array to determine correct bucket ends for each LMS
// suffix, preparing the array for final induction steps.
func expand(text, sa, summarySA, freq, bucket []int32, minChar int32) {
	frequency(text, freq, minChar)
	bucketEnd(freq, bucket)
	var lmsIdx, b, j int32
	// Insert LMS suffixes at bucket ends in reverse order.
	for i := len(summarySA) - 1; i >= 0; i-- {
		lmsIdx = summarySA[i]
		summarySA[i] = 0
		j = text[lmsIdx] - minChar
		b = bucket[j]
		sa[b] = lmsIdx
		bucket[j] = b - 1 // Update bucket position.
	}
}

// frequency counts occurrences of each character in the text.
func frequency(text, freq []int32, minChar int32) {
	clear(freq)
	for _, v := range text {
		freq[v-minChar]++
	}
}

// bucketStart calculates starting positions for L-type suffix buckets.
func bucketStart(freq, bucket []int32) {
	var offset int32
	for i, n := range freq {
		if n > 0 {
			bucket[i] = offset
			offset += n
		}
	}
}

// bucketEnd calculates ending positions for S-type suffix buckets.
func bucketEnd(freq, bucket []int32) {
	var offset int32
	for i, n := range freq {
		if n > 0 {
			offset += n
			bucket[i] = offset - 1
		}
	}
}

// insertLMS inserts LMS suffixes into their bucket ends in the suffix array.
// It scans the text backwards to identify LMS positions and places them at the end
// of their respective character buckets, marking the last LMS position as empty if
// multiple LMS suffixes exist.
func insertLMS(text, sa, freq, bucket []int32, minChar int32) {
	bucketEnd(freq, bucket)
	var (
		l, r, i, j, b, lastLMS int32
		numLMS                 int
		S                      bool
	)
	// Scan backwards to find LMS positions.
	for i = int32(len(text) - 1); i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			// Place LMS suffix at bucket end.
			j = r - minChar
			b = bucket[j]
			bucket[j] = b - 1
			sa[b] = i + 1
			lastLMS = b
			numLMS++
		}
	}
	// Clear last LMS position if multiple exist.
	if numLMS > 1 {
		sa[lastLMS] = 0
	}
}

// induceSubL induces L-type suffixes for the summary suffix array.
// It starts with the last character and scans forward, placing L-type suffixes
// at the start of their character buckets and marking processed suffixes as negative.
func induceSubL(text, sa, freq, bucket []int32, minChar int32) {
	bucketStart(freq, bucket)
	var (
		k, j     int32 = int32(len(text) - 1), 0
		l, r     int32 = text[k-1], text[k]
		lastChar int32 = text[len(text)-1]
		b        int32 = bucket[lastChar-minChar]
	)
	// Initialize with last character, marking L-type or S-type.
	if l < r {
		k = -k
	}
	bucket[lastChar-minChar] = b + 1
	sa[b] = int32(k)

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
		// Place at bucket start.
		b = bucket[r-minChar]
		bucket[r-minChar] = b + 1
		sa[b] = k
	}
}

// induceSubS induces S-type suffixes for the summary suffix array.
// It scans backward, placing S-type suffixes at the end of their character buckets
// and moving processed suffixes to the top of the array, marking them as negative.
func induceSubS(text, sa, freq, bucket []int32, minChar int32) {
	bucketEnd(freq, bucket)
	var (
		j, b, l, r, k int32
		top           = len(sa)
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
		// Place at bucket end.
		b = bucket[r-minChar]
		bucket[r-minChar] = b - 1
		sa[b] = k
	}
}

// induceL induces L-type suffixes for the final suffix array.
// It starts with the last character and scans forward, placing L-type suffixes
// at the start of their character buckets to complete the suffix array.
func induceL(text, sa, freq, bucket []int32, minChar int32) {
	bucketStart(freq, bucket)
	var (
		k, j     int32 = int32(len(text) - 1), 0
		l, r     int32 = text[k-1], text[k]
		lastChar int32 = text[len(text)-1]
		b        int32 = bucket[lastChar-minChar]
	)
	// Initialize with last character, marking L-type or S-type.
	if l < r {
		k = -k
	}
	bucket[lastChar-minChar] = b + 1
	sa[b] = int32(k)

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
		// Place at bucket start.
		b = bucket[r-minChar]
		bucket[r-minChar] = b + 1
		sa[b] = k
	}
}

// induceS induces S-type suffixes for the final suffix array.
// It scans backward, restoring processed suffixes and placing S-type suffixes
// at the end of their character buckets to finalize the suffix array.
func induceS(text, sa, freq, bucket []int32, minChar int32) {
	bucketEnd(freq, bucket)
	var (
		j, l, r, k, b int32
	)
	// Scan backward to induce S-type suffixes.
	for i := len(sa) - 1; i >= 0; i-- {
		j = sa[i]
		if j >= 0 {
			continue
		}
		j = -j
		sa[i] = j
		k = j - 1
		r = text[k]
		if k > 0 {
			// Check for S-type suffix.
			if l = text[k-1]; l <= r {
				k = -k
			}
		}
		// Place at bucket end.
		b = bucket[r-minChar]
		bucket[r-minChar] = b - 1
		sa[b] = k
	}
}

// lengthLMS computes the lengths of LMS substrings and stores them temporarily in sa.
func lengthLMS(text, sa []int32) {
	var (
		l, r int32
		prev int32 = int32(len(text)) - 1
		S    bool
	)
	// Scan backwards to calculate LMS substring lengths.
	for i := len(text) - 1; i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			// Store length of LMS substring.
			sa[(i+1)/2] = prev - int32(i)
			prev = int32(i)
		}
	}
}

// equalLMS checks if two LMS substrings are identical.
func equalLMS(text []int32, l, r, lLen, rLen int32) bool {
	if lLen != rLen {
		return false
	}
	for lLen > 0 {
		if text[l] != text[r] {
			return false
		}
		l++
		r++
		lLen--
	}
	return true
}

// summarise creates a summary string by assigning unique names to LMS substrings.
// It computes LMS substring lengths, compares adjacent substrings to assign names,
// and collects names into the summary array for recursive processing. Returns the
// maximum name assigned, indicating the number of unique LMS substrings.
func summarise(text, sa, summary []int32, numLMS int32) int32 {
	// Compute LMS substring lengths.
	lengthLMS(text, sa)
	var (
		name, maxName int32 = 1, 1
		posLMS              = summary
		prev, curr    int32 = sa[posLMS[0]], 0
		prevLen       int32 = sa[posLMS[0]/2]
	)
	// Assign initial name to first LMS substring.
	sa[posLMS[0]/2] = name
	// Compare consecutive LMS substrings to assign unique names.
	for i := 1; i < len(posLMS); i++ {
		prev = posLMS[i-1]
		curr = posLMS[i]
		// Increment name for distinct LMS substrings.
		if !equalLMS(text, prev, curr, prevLen, sa[curr/2]) {
			name++
			maxName++
		}
		prevLen = sa[curr/2]
		sa[curr/2] = name
	}
	if maxName >= numLMS {
		return maxName
	}
	// Collect names into summary array, clearing temporary storage.
	var j int
	for i := 0; i < len(sa)/2; i++ {
		curr := sa[i]
		if curr <= 0 {
			continue
		}
		sa[i], summary[j] = 0, curr
		j++
	}
	return maxName
}
