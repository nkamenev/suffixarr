package suffixarr

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genRandText_8_32(size int) []int32 {
	input := make([]int32, size)
	for i := 0; i < size; i++ {
		input[i] = rand.Int31n(255)
	}
	return input
}

func genRandText_32(size int) []int32 {
	input := make([]int32, size)
	for i := 0; i < size; i++ {
		input[i] = rand.Int31n(10000)
	}
	return input
}

func makeSA(text []int32) []int32 {
	sa := make([]int32, len(text))
	for i := range len(text) {
		sa[i] = int32(i)
	}
	sort.Slice(sa, func(i int, j int) bool {
		return string(text[sa[i]:]) < string(text[sa[j]:])
	})
	return sa
}

func TestSAIS(t *testing.T) {
	tests := map[string]struct {
		input []int32
	}{
		"empty string": {
			input: []int32{},
		},
		"single character": {
			input: []int32{100},
		},
		"same characters": {
			input: []int32("aaaaaaaaaaaaaaaaaaaaa"),
		},
		"1 LMS": {
			input: []int32("aabab"),
		},
		"2 LMS": {
			input: []int32("aababab"),
		},
		"banana": {
			input: []int32("banana"),
		},
		"repeated pattern": {
			input: []int32{1, 2, 1, 2, 1, 2, 1, 2},
		},
		"reverse sorted": {
			input: []int32{5, 4, 3, 2, 1},
		},
		"abracadabra": {
			input: []int32("abracadabra"),
		},
		"ACGTGCCTAGCCTACCGTGCC": {
			input: []int32("ACGTGCCTAGCCTACCGTGCC"),
		},
		"min/max edges": {
			input: []int32{0, 255},
		},
		"alternating pattern": {
			input: []int32{3, 1, 3, 1, 3, 1},
		},
		"zero characters": {
			input: []int32{0, 0, 0, 1, 1, 1},
		},
		"long random string 8": {
			input: genRandText_8_32(1000),
		},
		"long random string 32": {
			input: genRandText_32(1000),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, makeSA(tc.input), sais(tc.input))
		})
	}
}

func TestLookup(t *testing.T) {
	tests := map[string]struct {
		text, prefix, lexOrdExp, textOrdExp []int32
	}{
		"empty": {
			text:       []int32{},
			prefix:     []int32("a"),
			lexOrdExp:  []int32{},
			textOrdExp: []int32{},
		},
		"same characters": {
			text:       []int32("aaaaaaa"),
			prefix:     []int32("a"),
			lexOrdExp:  []int32{6, 5, 4, 3, 2, 1, 0},
			textOrdExp: []int32{0, 1, 2, 3, 4, 5, 6},
		},
		"banana": {
			text:       []int32("banana"),
			prefix:     []int32("banana"),
			lexOrdExp:  []int32{0},
			textOrdExp: []int32{0},
		},
		"anana": {
			text:       []int32("banana"),
			prefix:     []int32("anana"),
			lexOrdExp:  []int32{1},
			textOrdExp: []int32{1},
		},
		"nana": {
			text:       []int32("banana"),
			prefix:     []int32("nana"),
			lexOrdExp:  []int32{2},
			textOrdExp: []int32{2},
		},
		"ana": {
			text:       []int32("banana"),
			prefix:     []int32("ana"),
			lexOrdExp:  []int32{3, 1},
			textOrdExp: []int32{1, 3},
		},
		"na": {
			text:       []int32("banana"),
			prefix:     []int32("na"),
			lexOrdExp:  []int32{4, 2},
			textOrdExp: []int32{2, 4},
		},
		"a": {
			text:       []int32("banana"),
			prefix:     []int32("a"),
			lexOrdExp:  []int32{5, 3, 1},
			textOrdExp: []int32{1, 3, 5},
		},
		"not found": {
			text:       []int32("banana"),
			prefix:     []int32("ab"),
			lexOrdExp:  []int32{},
			textOrdExp: []int32{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.lexOrdExp, New(tc.text).Lookup(tc.prefix))
			assert.Equal(t, tc.textOrdExp, New(tc.text).LookupTextOrd(tc.prefix))
		})
	}
}

func eqSA(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func BenchmarkSAIS(b *testing.B) {
	tests := []struct {
		name     string
		input_32 []int32
	}{
		{"empty", []int32{}},
		{"single", []int32{100}},
		{"all same", []int32{5, 5, 5, 5, 5, 5}},
		{"unique", []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"repeated pattern", []int32{1, 2, 1, 2, 1, 2, 1, 2}},
		{"ACGTGCCTAGCCTACCGTGCC", []int32("ACGTGCCTAGCCTACCGTGCC")},
		{"long random string", genRandText_32(10000)},
		{"long random string 8", genRandText_8_32(10000)},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sais(tt.input_32)
			}
		})
	}
}
