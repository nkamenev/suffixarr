// Copyright (c) 2025 Nikita Kamenev
// Licensed under the MIT License. See LICENSE file in the project root for details.
package suffixarr

import (
	"sort"
)

// SuffixArray represents a suffix array for a given text.
// It stores the original text and the suffix array (indices of suffixes in lexicographical order).
type SuffixArray struct {
	text, sa []int32
}

// New creates a new SuffixArray for the given text.
// Parameters:
// - text: input text as a slice of int32.
// Returns a pointer to the constructed SuffixArray.
func New(text []int32) *SuffixArray {
	return &SuffixArray{text, sais(text)}
}

// comparePrefix compares a suffix with a prefix to determine their lexicographical order.
// Parameters:
// - suf: suffix to compare (slice of int32).
// - prefix: prefix to compare against (slice of int32).
// Returns:
// - -1 if suf < prefix.
// - 0 if suf starts with prefix.
// - 1 if suf > prefix.
func comparePrefix(suf, prefix []int32) int {
	minLen := len(suf)
	if minLen > len(prefix) {
		minLen = len(prefix)
	}
	for i := 0; i < minLen; i++ {
		if suf[i] < prefix[i] {
			return -1
		}
		if suf[i] > prefix[i] {
			return 1
		}
	}
	if len(suf) < len(prefix) {
		return -1
	}
	return 0
}

// Lookup finds all suffixes that start with the given prefix.
// Parameters:
// - prefix: prefix to search for (slice of int32).
// Returns a slice of indices from the suffix array where suffixes start with the prefix,
// in lexicographical order of the suffixes.
func (sa *SuffixArray) Lookup(prefix []int32) []int32 {
	if len(prefix) == 0 {
		return sa.sa
	}
	if len(sa.sa) == 0 {
		return []int32{}
	}
	l := sort.Search(len(sa.sa), func(i int) bool {
		suf := sa.text[sa.sa[i]:]
		return comparePrefix(suf, prefix) >= 0
	})
	r := l + sort.Search(len(sa.sa)-l, func(i int) bool {
		suf := sa.text[sa.sa[l+i]:]
		return comparePrefix(suf, prefix) > 0
	})
	return sa.sa[l:r]
}

// LookupTextOrd finds all suffixes that start with the given prefix and returns their indices
// in the order they appear in the original text.
// Parameters:
// - prefix: prefix to search for (slice of int32).
// Returns a slice of indices sorted by their position in the text.
func (sa *SuffixArray) LookupTextOrd(prefix []int32) []int32 {
	indices := sa.Lookup(prefix)
	cp := make([]int32, len(indices))
	copy(cp, indices)

	sort.Slice(cp, func(i, j int) bool {
		return cp[i] < cp[j]
	})

	return cp
}
