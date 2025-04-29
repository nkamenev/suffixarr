// Copyright (c) 2025 Nikita Kamenev
// Licensed under the MIT License. See LICENSE file in the project root for details.
package suffixarr

import (
	"slices"
	"sort"
	"unicode/utf8"
)

// sep is a special character used to separate strings in the generalized suffix array.
// It is chosen from the Unicode Private Use Area (PUA), U+E000, to avoid
// conflicts with actual text characters.
const sep int32 = 0xE000

// SuffixArray holds a text and its suffix array.
type SuffixArray struct {
	text, sa []int32
}

// New creates a suffix array for the given text.
func New(text []int32) *SuffixArray {
	return &SuffixArray{text, sais(text)}
}

// comparePrefix compares a suffix with a prefix lexicographically.
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

// lookup finds suffixes starting with the given prefix.
func lookup(text, sa, prefix []int32) []int32 {
	if len(prefix) == 0 {
		return sa
	}
	if len(sa) == 0 {
		return []int32{}
	}
	// Find left boundary where suffix >= prefix.
	l := sort.Search(len(sa), func(i int) bool {
		suf := text[sa[i]:]
		return comparePrefix(suf, prefix) >= 0
	})
	// Find right boundary where suffix > prefix.
	r := l + sort.Search(len(sa)-l, func(i int) bool {
		suf := text[sa[l+i]:]
		return comparePrefix(suf, prefix) > 0
	})
	return sa[l:r]
}

// lookupTextOrder finds suffixes starting with the prefix, sorted by text position.
func lookupTextOrder(text, sa, prefix []int32) []int32 {
	indices := lookup(text, sa, prefix)
	cp := make([]int32, len(indices))
	copy(cp, indices)
	// Sort indices by their position in the original text.
	sort.Slice(cp, func(i, j int) bool {
		return cp[i] < cp[j]
	})
	return cp
}

// Lookup finds suffixes starting with the given prefix.
func (sa *SuffixArray) Lookup(prefix []int32) []int32 {
	return lookup(sa.text, sa.sa, prefix)
}

// LookupTextOrder finds suffixes starting with the prefix, sorted by text position.
func (sa *SuffixArray) LookupTextOrder(prefix []int32) []int32 {
	return lookupTextOrder(sa.text, sa.sa, prefix)
}

// LookupSuffix finds the exact suffix in the text.
// For an empty suffix, returns len(sa) as it occurs at the end of the string.
// Otherwise, returns the starting index or -1 if not found.
func (sa *SuffixArray) LookupSuffix(suffix []int32) int {
	if len(suffix) == 0 {
		return len(sa.sa) // Empty suffix is at the end of the string.
	}
	if len(sa.sa) == 0 || len(suffix) > len(sa.text) {
		return -1
	}
	// Check if the suffix matches the end of the text.
	l := len(sa.text) - len(suffix)
	if slices.Compare(sa.text[l:], suffix) == 0 {
		return l
	}
	return -1
}

// LookupPrefix checks if the text starts with the given prefix.
// For an empty prefix, returns -1 as it precedes the first character.
// Returns 0 if matched, -2 otherwise.
func (sa *SuffixArray) LookupPrefix(prefix []int32) int {
	if len(prefix) == 0 {
		return -1 // Empty prefix is invalid, precedes first character.
	}
	if len(sa.sa) == 0 || len(prefix) > len(sa.text) {
		return -2
	}
	if slices.Compare(sa.text[:len(prefix)], prefix) == 0 {
		return 0
	}
	return -2
}

// index stores metadata(l, i) and buffer for a substring in the generalized suffix array.
type index struct {
	l, i int
	sa   []int32
}

// GSA represents a generalized suffix array for multiple strings.
type GSA struct {
	src              [][]int32 // Original strings.
	text, sa, strIdx []int32   // Concatenated text, suffix array, and string indices.
	idx              []index   // Buffer and metadata for each substring.
	index            []Index   // Buffer for occurrence indices for lookup results.
}

// newGSA_32 builds a generalized suffix array for int32 strings.
func newGSA_32(src [][]int32, strNum int) *GSA {
	// Allocate buffer for text, string indices, and suffix arrays.
	textSz := strNum + len(src) + 1
	buf := make([]int32, textSz*2+strNum)
	text := buf[:textSz]
	strIdx, idxBuf := buf[textSz:textSz*2], buf[textSz*2:]
	idx := make([]index, len(src))

	// Initialize text with separator.
	text[0] = sep
	var (
		l, r    int        // Buffer boundaries for each substring.
		ll, pos int = 1, 1 // Left boundary and current position in text.
	)
	// Concatenate strings with separators, track indices.
	for i := 0; i < len(src); i++ {
		for j := 0; j < len(src[i]); j++ {
			text[pos], strIdx[pos] = src[i][j], int32(i)
			pos++
		}
		r += len(src[i])
		// Store string metadata.
		curr := idx[i]
		curr.l, curr.sa = ll, idxBuf[l:r]
		idx[i], strIdx[pos], text[pos] = curr, int32(i), sep
		pos++
		ll += len(src[i]) + 1
		l = r
	}
	// Build suffix array for concatenated text.
	sa := sais(text)
	return &GSA{src, text, sa, strIdx, idx, make([]Index, len(src))}
}

// NewGSA creates a generalized suffix array from strings.
func NewGSA(src []string) *GSA {
	if len(src) == 0 {
		return nil
	}
	// Convert strings to int32 slices.
	src32 := make([][]int32, len(src))
	var sz int
	for i := 0; i < len(src); i++ {
		sz += utf8.RuneCountInString(src[i])
		src32[i] = []int32(src[i])
	}
	return newGSA_32(src32, sz)
}

// NewGSA_32 creates a generalized suffix array from int32 slices.
func NewGSA_32(src [][]int32) *GSA {
	if len(src) == 0 {
		return nil
	}
	// Calculate total character count.
	var sz int
	for i := 0; i < len(src); i++ {
		sz += len(src[i])
	}
	return newGSA_32(src, sz)
}

// fillIdx fill gsa.idx with indexes from sa according to substrings
// Returns the number of strings with occurrences.
func (gsa *GSA) fillIdx(sa []int32) (sz int) {
	var prev int32 // Previous processed sa index
	for i := 0; i < len(sa); i++ {
		j := sa[i]
		// Skip separator unless followed by a valid character.
		if gsa.text[j] == sep {
			if int(j) == len(gsa.text)-1 {
				break
			}
			j++
		}
		// Avoid duplicate indices.
		if j == prev {
			continue
		}
		str := gsa.strIdx[j]
		curr := gsa.idx[str]
		// Increment unique string count on first occurrence.
		if curr.i == 0 {
			sz++
		}
		// Store offset relative to string start.
		curr.sa[curr.i] = j - int32(curr.l)
		curr.i++
		gsa.idx[str] = curr
		prev = j
	}
	return
}

// Index holds a string's occurrences in the generalized suffix array.
type Index struct {
	String     int32
	Occurences []int32
}

// makeIndex generates occurrence indices for strings.
func (gsa *GSA) makeIndex(sa []int32, sz int) []Index {
	index := gsa.index[:sz]
	var (
		k    int   // Current index in result.
		prev int32 // Previous processed sa index.
	)
	for i := 0; i < len(sa); i++ {
		j := sa[i]
		// Skip separator unless followed by a valid character.
		if gsa.text[j] == sep {
			if int(j) == len(gsa.text)-1 {
				break
			}
			j++
		}
		if j == prev {
			continue
		}
		str := gsa.strIdx[j]
		idx := gsa.idx[str]
		if idx.i == 0 {
			continue
		}
		// Store string index and its occurrences.
		curr := Index{str, idx.sa[:idx.i]}
		gsa.idx[str].i = 0
		index[k] = curr
		k++
	}
	return index
}

// LookupTextOrder finds prefix occurrences in the generalized suffix array, sorted by text position.
func (gsa *GSA) LookupTextOrder(prefix []int32) []Index {
	res := lookupTextOrder(gsa.text, gsa.sa, prefix)
	sz := gsa.fillIdx(res)
	return gsa.makeIndex(res, sz)
}

// LookupSuffix finds suffix occurrences in the generalized suffix array, sorted by text position.
func (gsa *GSA) LookupSuffix(suf []int32) []Index {
	if len(suf) == 0 {
		// Returns the length of each substring as the index of the empty suffix.
		for i := 0; i < len(gsa.src); i++ {
			l := len(gsa.idx[i].sa)
			gsa.idx[i].sa[0] = int32(l)
			gsa.index[i] = Index{int32(i), gsa.idx[i].sa[:1]}
		}
		return gsa.index
	}
	// Append separator to ensure exact suffix match.
	suf = append(suf, sep)
	res := lookupTextOrder(gsa.text, gsa.sa, suf)
	sz := gsa.fillIdx(res)
	return gsa.makeIndex(res, sz)
}

// LookupPrefix finds prefix occurrences in the generalized suffix array, sorted by text position.
func (gsa *GSA) LookupPrefix(suf []int32) []Index {
	if len(suf) == 0 {
		// Return -1 for each string if prefix is empty.
		for i := 0; i < len(gsa.src); i++ {
			gsa.idx[i].sa[0] = -1
			gsa.index[i] = Index{int32(i), gsa.idx[i].sa[:1]}
		}
		return gsa.index
	}
	// Prepend separator to match string start.
	cp := make([]int32, len(suf)+1)
	cp[0] = sep
	copy(cp[1:], suf)
	res := lookupTextOrder(gsa.text, gsa.sa, cp)
	sz := gsa.fillIdx(res)
	return gsa.makeIndex(res, sz)
}
