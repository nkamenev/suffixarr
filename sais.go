// Copyright (c) 2025 Nikita Kamenev
// Licensed under the MIT License. See LICENSE file in the project root for details.
package suffixarr

// sais constructs a suffix array for the given text using the SA-IS algorithm.
// Returns an array of int32 representing the starting indices of suffixes in lexicographical order.
func sais(text []int32) []int32 {
	if len(text) == 0 {
		return []int32{}
	} else if len(text) == 1 {
		return []int32{0}
	}
	return _sais(text, nil, nil, 0)
}

// _sais is a recursive helper for sais, implementing the core SA-IS algorithm.
// Parameters:
// - text: input text as a slice of int32.
// - sa: suffix array to store results, or nil to allocate.
// - data: auxiliary array for frequency and bucket calculations, or nil to allocate.
// - srcAlphaSize: size of the alphabet for the original text.
// Returns the constructed suffix array.
func _sais(text, sa, data []int32, srcAlphaSize int32) []int32 {
	var (
		minChar, maxChar int32 = text[0], text[0]
		l, r, numLMS     int32
		S                bool
	)
	// Scan text to find min/max characters and count LMS (Left-Most S-type) suffixes.
	for i := len(text) - 1; i >= 0; i-- {
		l, r = text[i], l
		if l < minChar {
			minChar = l
		}
		if l > maxChar {
			maxChar = l
		}
		// Detect S-type positions (l < r) and LMS positions (l > r after S-type).
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			numLMS++
		}
	}
	// Calculate current alphabet size based on character range.
	currAlphaSize := maxChar - minChar + 1
	if sa == nil {
		// Allocate suffix array if not provided.
		srcAlphaSize = currAlphaSize
		sa = make([]int32, len(text))
	}
	// Use arbitrary alphabet sorting for large alphabets or recursive calls.
	if currAlphaSize > 256 || currAlphaSize > srcAlphaSize {
		return induceSort_arb(text, sa, data, numLMS)
	}
	return induceSort(text, sa, data, minChar, numLMS, srcAlphaSize, currAlphaSize)
}

// induceSort constructs the suffix array using induced sorting for small alphabets (<= 256).
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - data: auxiliary array for frequency and bucket calculations.
// - minChar: minimum character in the text.
// - numLMS: number of LMS (Left-Most S-type) suffixes.
// - srcAlphaSize: size of the original alphabet.
// - currAlphaSize: size of the current alphabet.
// Returns the constructed suffix array.
func induceSort(text, sa, data []int32, minChar, numLMS, srcAlphaSize, currAlphaSize int32) []int32 {
	// Allocate or reuse auxiliary array for frequency and buckets.
	if data == nil || len(data) < int(srcAlphaSize)*2 {
		data = make([]int32, srcAlphaSize*2)
	}
	var summary []int32
	freq := data[:currAlphaSize]
	buckets := data[srcAlphaSize : srcAlphaSize+currAlphaSize]
	frequency(text, freq, minChar)

	// Place LMS suffixes into their buckets.
	insertLMS(text, sa, freq, buckets, minChar)
	if numLMS > 1 {
		// Induce L-type and S-type suffixes for the summary array.
		induceSubL(text, sa, freq, buckets, minChar)
		induceSubS(text, sa, freq, buckets, minChar)
		// Extract LMS substring indices for summary string.
		summary = sa[len(sa)-int(numLMS):]
		maxName := summarise(text, sa, summary, numLMS)

		summarySA := sa[:numLMS]
		if maxName < numLMS {
			// Recursively construct suffix array for summary string if unique LMS substrings exist.
			_sais(summary, summarySA, data, srcAlphaSize)
			// Map summary indices back to original text.
			unmap(text, sa, summarySA, summary)
		} else {
			// Copy summary directly if all LMS substrings are unique.
			copy(summarySA, summary)
			clear(sa[numLMS:])
		}
		// Expand LMS suffixes to their correct positions.
		expand(text, sa, summarySA, freq, buckets, minChar)
	}
	// Final induction steps to complete the suffix array.
	induceL(text, sa, freq, buckets, minChar)
	induceS(text, sa, freq, buckets, minChar)
	return sa
}

// unmap maps LMS substring indices from the summary suffix array back to the original text.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - summarySA: suffix array of the summary string.
// - LMS: array to store LMS indices temporarily.
func unmap(text, sa, summarySA, LMS []int32) {
	var (
		j    int32 = int32(len(LMS))
		l, r int32
		S    bool
	)
	// Collect LMS indices in reverse order of appearance in text.
	for i := len(text) - 1; i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			j--
			LMS[j] = int32(i) + 1
		}
	}
	// Map summary suffix array indices to original LMS positions.
	for i := 0; i < len(LMS); i++ {
		j = summarySA[i]
		sa[i] = LMS[j]
		LMS[j] = 0
	}
}

// expand places LMS suffixes into the suffix array using bucket sorting.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - summarySA: suffix array of the summary string.
// - freq: frequency array for characters.
// - bucket: bucket array for sorting.
// - minChar: minimum character in the text.
func expand(text, sa, summarySA, freq, bucket []int32, minChar int32) {
	frequency(text, freq, minChar)
	bucketEnd(freq, bucket)
	var lmsIdx, b, j int32
	// Place LMS suffixes into their bucket ends in reverse order.
	for i := len(summarySA) - 1; i >= 0; i-- {
		lmsIdx = summarySA[i]
		summarySA[i] = 0
		j = text[lmsIdx] - minChar
		b = bucket[j]
		sa[b] = lmsIdx
		bucket[j] = b - 1
	}
}

// frequency calculates the frequency of each character in the text.
// Parameters:
// - text: input text.
// - freq: array to store character frequencies.
// - minChar: minimum character in the text.
func frequency(text, freq []int32, minChar int32) {
	clear(freq)
	for _, v := range text {
		freq[v-minChar]++
	}
}

// bucketStart computes the starting positions of buckets for L-type suffixes.
// Parameters:
// - freq: frequency array for characters.
// - bucket: array to store bucket start positions.
func bucketStart(freq, bucket []int32) {
	var offset int32
	for i, n := range freq {
		if n > 0 {
			bucket[i] = offset
			offset += n
		}
	}
}

// bucketEnd computes the ending positions of buckets for S-type suffixes.
// Parameters:
// - freq: frequency array for characters.
// - bucket: array to store bucket end positions.
func bucketEnd(freq, bucket []int32) {
	var offset int32
	for i, n := range freq {
		if n > 0 {
			offset += n
			bucket[i] = offset - 1
		}
	}
}

// insertLMS inserts LMS suffixes into the suffix array.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - freq: frequency array for characters.
// - bucket: bucket array for sorting.
// - minChar: minimum character in the text.
func insertLMS(text, sa, freq, bucket []int32, minChar int32) {
	bucketEnd(freq, bucket)
	var (
		l, r, i, j, b, lastLMS int32
		numLMS                 int
		S                      bool
	)

	// Scan text backwards to identify LMS positions and insert them into buckets.
	for i = int32(len(text) - 1); i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			// Insert LMS suffix at the end of its character's bucket.
			j = r - minChar
			b = bucket[j]
			bucket[j] = b - 1
			sa[b] = i + 1
			lastLMS = b
			numLMS++
		}
	}
	// Mark the last LMS position as empty if multiple LMS suffixes exist.
	if numLMS > 1 {
		sa[lastLMS] = 0
	}
}

// induceSubL induces L-type suffixes for the summary suffix array.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - freq: frequency array for characters.
// - bucket: bucket array for sorting.
// - minChar: minimum character in the text.
func induceSubL(text, sa, freq, bucket []int32, minChar int32) {
	bucketStart(freq, bucket)
	var (
		k, j     int32 = int32(len(text) - 1), 0
		l, r     int32 = text[k-1], text[k]
		lastChar int32 = text[len(text)-1]
		b        int32 = bucket[lastChar-minChar]
	)
	// Initialize with the last character, marking it as L-type or S-type.
	if l < r {
		k = -k
	}
	bucket[lastChar-minChar] = b + 1
	sa[b] = int32(k)

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
		b = bucket[r-minChar]
		bucket[r-minChar] = b + 1
		sa[b] = k
	}
}

// induceSubS induces S-type suffixes for the summary suffix array.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - freq: frequency array for characters.
// - bucket: bucket array for sorting.
// - minChar: minimum character in the text.
func induceSubS(text, sa, freq, bucket []int32, minChar int32) {
	bucketEnd(freq, bucket)
	var (
		j, b, l, r, k int32
		top           = len(sa)
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
		b = bucket[r-minChar]
		bucket[r-minChar] = b - 1
		sa[b] = k
	}
}

// induceL induces L-type suffixes for the final suffix array.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - freq: frequency array for characters.
// - bucket: bucket array for sorting.
// - minChar: minimum character in the text.
func induceL(text, sa, freq, bucket []int32, minChar int32) {
	bucketStart(freq, bucket)
	var (
		k, j     int32 = int32(len(text) - 1), 0
		l, r     int32 = text[k-1], text[k]
		lastChar int32 = text[len(text)-1]
		b        int32 = bucket[lastChar-minChar]
	)
	// Initialize with the last character, marking it as L-type or S-type.
	if l < r {
		k = -k
	}
	bucket[lastChar-minChar] = b + 1
	sa[b] = int32(k)

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
		b = bucket[r-minChar]
		bucket[r-minChar] = b + 1
		sa[b] = k
	}
}

// induceS induces S-type suffixes for the final suffix array.
// Parameters:
// - text: input text.
// - sa: suffix array to store results.
// - freq: frequency array for characters.
// - bucket: bucket array for sorting.
// - minChar: minimum character in the text.
func induceS(text, sa, freq, bucket []int32, minChar int32) {
	bucketEnd(freq, bucket)
	var (
		j, l, r, k, b int32
	)

	// Scan suffix array backwards to induce S-type suffixes.
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
			// Check if the preceding character forms an S-type suffix.
			if l = text[k-1]; l <= r {
				k = -k
			}
		}
		// Place suffix at the end of its character's bucket.
		b = bucket[r-minChar]
		bucket[r-minChar] = b - 1
		sa[b] = k
	}
}

// lengthLMS computes the lengths of LMS substrings and stores them in sa.
// Parameters:
// - text: input text.
// - sa: suffix array to store LMS lengths temporarily.
func lengthLMS(text, sa []int32) {
	var (
		l, r int32
		prev int32 = int32(len(text)) - 1
		S    bool
	)
	// Scan text backwards to compute LMS substring lengths.
	for i := len(text) - 1; i >= 0; i-- {
		l, r = text[i], l
		if l < r {
			S = true
		} else if l > r && S {
			S = false
			// Store length of LMS substring ending at prev, starting at i+1.
			sa[(i+1)/2] = prev - int32(i)
			prev = int32(i)
		}
	}
}

// equalLMS checks if two LMS substrings are equal.
// Parameters:
// - text: input text.
// - l, r: starting positions of the LMS substrings.
// - lLen, rLen: lengths of the LMS substrings.
// Returns true if the LMS substrings are equal, false otherwise.
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

// summarise creates a summary string from LMS substrings and assigns names.
// Parameters:
// - text: input text.
// - sa: suffix array used for temporary storage.
// - summary: array to store the summary string.
// - numLMS: number of LMS substrings.
// Returns the maximum name assigned to LMS substrings.
func summarise(text, sa, summary []int32, numLMS int32) int32 {
	// Compute LMS substring lengths and store in sa.
	lengthLMS(text, sa)
	var (
		name, maxName int32 = 1, 1
		posLMS              = summary
		prev, curr    int32 = sa[posLMS[0]], 0
		prevLen       int32 = sa[posLMS[0]/2]
	)
	// Assign initial name to the first LMS substring.
	sa[posLMS[0]/2] = name
	// Compare consecutive LMS substrings to assign unique names.
	for i := 1; i < len(posLMS); i++ {
		prev = posLMS[i-1]
		curr = posLMS[i]
		// Increment name if LMS substrings differ.
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
	// Collect names into summary array, clearing sa.
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
