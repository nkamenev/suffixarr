package suffixarr

import (
	"math/rand"
	"slices"
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
		input[i] = rand.Int31()
	}
	return input
}

func makeSA(text []int32) []int32 {
	sa := make([]int32, len(text))
	for i := range len(text) {
		sa[i] = int32(i)
	}
	sort.Slice(sa, func(i int, j int) bool {
		return slices.Compare(text[sa[i]:], text[sa[j]:]) < 0
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
		text,
		prefix,
		suffix,
		lexOrdExp,
		textOrdExp []int32
		prefixExp int
		sufExp    int
	}{
		"empty text": {
			text:       []int32{},
			prefix:     []int32("a"),
			suffix:     []int32("a"),
			lexOrdExp:  []int32{},
			textOrdExp: []int32{},
			prefixExp:  -2,
			sufExp:     -1,
		},
		"empty prefix": {
			text:       []int32("aaaaaaa"),
			prefix:     []int32{},
			suffix:     []int32{},
			lexOrdExp:  []int32{6, 5, 4, 3, 2, 1, 0},
			textOrdExp: []int32{0, 1, 2, 3, 4, 5, 6},
			prefixExp:  -1,
			sufExp:     7,
		},
		"same characters": {
			text:       []int32("aaaaaaa"),
			prefix:     []int32("a"),
			suffix:     []int32("a"),
			lexOrdExp:  []int32{6, 5, 4, 3, 2, 1, 0},
			textOrdExp: []int32{0, 1, 2, 3, 4, 5, 6},
			prefixExp:  0,
			sufExp:     6,
		},
		"banana": {
			text:       []int32("banana"),
			prefix:     []int32("banana"),
			suffix:     []int32("banana"),
			lexOrdExp:  []int32{0},
			textOrdExp: []int32{0},
			prefixExp:  0,
			sufExp:     0,
		},
		"anana": {
			text:       []int32("banana"),
			prefix:     []int32("banan"),
			suffix:     []int32("anana"),
			lexOrdExp:  []int32{1},
			textOrdExp: []int32{1},
			prefixExp:  0,
			sufExp:     1,
		},
		"nana": {
			text:       []int32("banana"),
			prefix:     []int32("bana"),
			suffix:     []int32("nana"),
			lexOrdExp:  []int32{2},
			textOrdExp: []int32{2},
			prefixExp:  0,
			sufExp:     2,
		},
		"ana": {
			text:       []int32("banana"),
			prefix:     []int32("ban"),
			suffix:     []int32("ana"),
			lexOrdExp:  []int32{3, 1},
			textOrdExp: []int32{1, 3},
			prefixExp:  0,
			sufExp:     3,
		},
		"na": {
			text:       []int32("banana"),
			suffix:     []int32("na"),
			prefix:     []int32("ba"),
			lexOrdExp:  []int32{4, 2},
			textOrdExp: []int32{2, 4},
			prefixExp:  0,
			sufExp:     4,
		},
		"a": {
			text:       []int32("banana"),
			prefix:     []int32("b"),
			suffix:     []int32("a"),
			lexOrdExp:  []int32{5, 3, 1},
			textOrdExp: []int32{1, 3, 5},
			prefixExp:  0,
			sufExp:     5,
		},
		"not found": {
			text:       []int32("banana"),
			prefix:     []int32("ab"),
			suffix:     []int32("ab"),
			lexOrdExp:  []int32{},
			textOrdExp: []int32{},
			prefixExp:  -2,
			sufExp:     -1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.lexOrdExp, New(tc.text).Lookup(tc.suffix))
			assert.Equal(t, tc.textOrdExp, New(tc.text).LookupTextOrder(tc.suffix))
			assert.Equal(t, tc.sufExp, New(tc.text).LookupSuffix(tc.suffix))
			assert.Equal(t, tc.prefixExp, New(tc.text).LookupPrefix(tc.prefix))
		})
	}
}

func TestGSA(t *testing.T) {
	tests := map[string]struct {
		text_32 [][]int32
		text    []string
		prefix  []int32
		exp     []Index
	}{
		"empty text": {
			text_32: [][]int32{},
			text:    []string{},
			prefix:  []int32("ab"),
			exp:     []Index{},
		},
		"empty prefix": {
			text_32: [][]int32{[]int32("aaaaaaa")},
			text:    []string{"aaaaaaa"},
			prefix:  []int32{},
			exp:     []Index{{0, []int32{0, 1, 2, 3, 4, 5, 6}}},
		},
		"single": {
			text_32: [][]int32{[]int32("a")},
			text:    []string{"a"},
			prefix:  []int32("a"),
			exp:     []Index{{0, []int32{0}}},
		},
		"all same in one string": {
			text_32: [][]int32{[]int32("aaaaaaa")},
			text:    []string{"aaaaaaa"},
			prefix:  []int32("a"),
			exp:     []Index{{0, []int32{0, 1, 2, 3, 4, 5, 6}}},
		},
		"all same in multiple strings": {
			text_32: [][]int32{[]int32("aaaaaaa"), []int32("aaaaa")},
			text:    []string{"aaaaaaa", "aaaaa"},
			prefix:  []int32("a"),
			exp:     []Index{{0, []int32{0, 1, 2, 3, 4, 5, 6}}, {1, []int32{0, 1, 2, 3, 4}}},
		},
		"one different string": {
			text_32: [][]int32{[]int32("abbacdababaaaaaab")},
			text:    []string{"abbacdababaaaaaab"},
			prefix:  []int32("ab"),
			exp:     []Index{{0, []int32{0, 6, 8, 15}}},
		},
		"multiple strings with many occurrences": {
			text_32: [][]int32{
				[]int32("abzababab"),
				[]int32("babaxyzab"),
				[]int32("jvoabbabrpvpabewge"),
				[]int32("wcccchervabgimeog"),
				[]int32("xqabqqqhfimmoabmhbaabfiq"),
				[]int32("cqoiwhoihabewqh"),
				[]int32("xxhoiababhehqab"),
				[]int32("qihcoiabhwca"),
				[]int32("qoixh79bbab"),
				[]int32("oihcqoihoieabicq"),
				[]int32("abababababababab"),
				[]int32("ociioimcwwwababa"),
				[]int32("aboiqhconhwiehcoiqwwfab"),
				[]int32("pqcpmwpeoicwq"),
				[]int32("mevmbxouccoiwq"),
				[]int32("bababicqqqqqqk"),
				[]int32("bbbbbbbbbbbbbbb"),
				[]int32("aaaaaaaaaaaabbbb"),
				[]int32("bbbaaaabbbaaaabab"),
				[]int32("xxxxxxxyyyyyyyyzzzz"),
			},
			text: []string{
				"abzababab",
				"babaxyzab",
				"jvoabbabrpvpabewge",
				"wcccchervabgimeog",
				"xqabqqqhfimmoabmhbaabfiq",
				"cqoiwhoihabewqh",
				"xxhoiababhehqab",
				"qihcoiabhwca",
				"qoixh79bbab",
				"oihcqoihoieabicq",
				"abababababababab",
				"ociioimcwwwababa",
				"aboiqhconhwiehcoiqwwfab",
				"pqcpmwpeoicwq",
				"mevmbxouccoiwq",
				"bababicqqqqqqk",
				"bbbbbbbbbbbbbbb",
				"aaaaaaaaaaaabbbb",
				"bbbaaaabbbaaaabab",
				"xxxxxxxyyyyyyyyzzzz",
			},
			prefix: []int32("ab"),
			exp: []Index{
				{0, []int32{0, 3, 5, 7}},
				{1, []int32{1, 7}},
				{2, []int32{3, 6, 12}},
				{3, []int32{9}},
				{4, []int32{2, 13, 19}},
				{5, []int32{9}},
				{6, []int32{5, 7, 13}},
				{7, []int32{6}},
				{8, []int32{9}},
				{9, []int32{11}},
				{10, []int32{0, 2, 4, 6, 8, 10, 12, 14}},
				{11, []int32{11, 13}},
				{12, []int32{0, 21}},
				{15, []int32{1, 3}},
				{17, []int32{11}},
				{18, []int32{6, 13, 15}},
			},
		},
		"multiple strings with one occurrence": {
			text_32: [][]int32{
				[]int32("cnklnldskk"),
				[]int32("jwofjpppmcppppppppppw"),
				[]int32("oqccpowcccwq"),
				[]int32("poqcurmpowww"),
				[]int32("ouqcomopooew"),
				[]int32("cqoiwhoihewqh"),
				[]int32("xxhoihehq"),
				[]int32("qihcoihwc"),
				[]int32("qoixh79"),
				[]int32("oihcqoihoieicq"),
				[]int32("ociioimcwwwababa"),
				[]int32("oiqhconhwiehcoiqwwf"),
				[]int32("pqcpmwpeoicwq"),
				[]int32("mevmbxouccoiwq"),
				[]int32("bababicqqqqqqk"),
				[]int32("bbbbbbbbbbbbbbb"),
				[]int32("aaaaaaaaaaaabbbb"),
				[]int32("bbbaaaabbbaaaabab"),
				[]int32("xxxxxxxyyyyyyyyzzzz"),
			},
			text: []string{
				"cnklnldskk",
				"jwofjpppmcppppppppppw",
				"oqccpowcccwq",
				"poqcurmpowww",
				"ouqcomopooew",
				"cqoiwhoihewqh",
				"xxhoihehq",
				"qihcoihwc",
				"qoixh79",
				"oihcqoihoieicq",
				"ociioimcwwwababa",
				"oiqhconhwiehcoiqwwf",
				"pqcpmwpeoicwq",
				"mevmbxouccoiwq",
				"bababicqqqqqqk",
				"bbbbbbbbbbbbbbb",
				"aaaaaaaaaaaabbbb",
				"bbbaaaabbbaaaabab",
				"xxxxxxxyyyyyyyyzzzz",
			},
			prefix: []int32("pmwpeo"),
			exp: []Index{
				{12, []int32{3}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gsa_32 := NewGSA_32(tc.text_32)
			gsa := NewGSA(tc.text)
			if name == "empty text" {
				assert.Nil(t, gsa_32)
				assert.Nil(t, gsa)
				return
			}
			assert.Equal(t, tc.exp, gsa_32.LookupTextOrder(tc.prefix))
			assert.Equal(t, tc.exp, gsa.LookupTextOrder(tc.prefix))
		})
	}
}

func TestGSALookup(t *testing.T) {
	tests := map[string]struct {
		text            [][]int32
		prefix, suffix  []int32
		expPref, expSuf []Index
	}{
		"empty text": {
			text:    [][]int32{},
			prefix:  []int32("ab"),
			suffix:  []int32("ab"),
			expPref: []Index{},
			expSuf:  []Index{},
		},
		"empty suffix": {
			text: [][]int32{
				[]int32("aaa"),
				[]int32("bbbb"),
				[]int32("ccccc"),
			},
			prefix: []int32{},
			suffix: []int32{},
			expPref: []Index{
				{0, []int32{-1}},
				{1, []int32{-1}},
				{2, []int32{-1}},
			},
			expSuf: []Index{
				{0, []int32{3}},
				{1, []int32{4}},
				{2, []int32{5}},
			},
		},
		"not found": {
			text: [][]int32{
				[]int32("aaa"),
				[]int32("bbbb"),
				[]int32("ccccc"),
			},
			prefix:  []int32("x"),
			suffix:  []int32("x"),
			expPref: []Index{},
			expSuf:  []Index{},
		},
		"single": {
			text:    [][]int32{[]int32("a")},
			prefix:  []int32("a"),
			suffix:  []int32("a"),
			expPref: []Index{{0, []int32{0}}},
			expSuf:  []Index{{0, []int32{0}}},
		},
		"all same in one string": {
			text:    [][]int32{[]int32("aaaaaaa")},
			prefix:  []int32("a"),
			suffix:  []int32("a"),
			expPref: []Index{{0, []int32{0}}},
			expSuf:  []Index{{0, []int32{6}}},
		},
		"all same in multiple strings": {
			text:    [][]int32{[]int32("aaaaaaa"), []int32("aaaaa")},
			prefix:  []int32("a"),
			suffix:  []int32("a"),
			expPref: []Index{{0, []int32{0}}, {1, []int32{0}}},
			expSuf:  []Index{{0, []int32{6}}, {1, []int32{4}}},
		},
		"one different string": {
			text:    [][]int32{[]int32("abbacdababaaaaaab")},
			prefix:  []int32("ab"),
			suffix:  []int32("ab"),
			expPref: []Index{{0, []int32{0}}},
			expSuf:  []Index{{0, []int32{15}}},
		},
		"multiple strings with many occurrences": {
			text: [][]int32{
				[]int32("abazabababxyz"),                      // 0
				[]int32("abacwimrivwwoiwmcxyz"),               // 1
				[]int32("abajomcoojwpmw438xyz"),               // 2
				[]int32("kssshvliwii"),                        // 3
				[]int32("abaisssmmmmmmi643xyyz"),              // 4
				[]int32("abaisssmmmmmmi643xyz"),               // 5
				[]int32("abalkmlclwwc6496593527983269854xyz"), // 6
				[]int32("abaxyz"),                             // 7
				[]int32("abaxyzxyz"),                          // 8
			},
			prefix: []int32("aba"),
			suffix: []int32("xyz"),
			expPref: []Index{
				{0, []int32{0}},
				{1, []int32{0}},
				{2, []int32{0}},
				{4, []int32{0}},
				{5, []int32{0}},
				{6, []int32{0}},
				{7, []int32{0}},
				{8, []int32{0}},
			},
			expSuf: []Index{
				{0, []int32{10}},
				{1, []int32{17}},
				{2, []int32{17}},
				{5, []int32{17}},
				{6, []int32{31}},
				{7, []int32{3}},
				{8, []int32{6}},
			},
		},
		"multiple strings with one occurrence": {
			text: [][]int32{
				[]int32("cnklnldskk"),
				[]int32("jwofjpppmcppppppppppw"),
				[]int32("oqccpowcccwq"),
				[]int32("poqcurmpowww"),
				[]int32("ouqcomopooew"),
				[]int32("cqoiwhoihewqh"),
				[]int32("xxhoihehq"),
				[]int32("abaqihcoihwc"),
				[]int32("qoixh79"),
				[]int32("oihcqoihoieicq"),
				[]int32("ociioimcwwwababa"),
				[]int32("oiqhconhwiehcoiqwwf"),
				[]int32("pqcpmwpeoicwq"),
				[]int32("mevmbxouccoiwq"),
				[]int32("bababicqqqqqqk"),
				[]int32("bbbbbbbbbbbbbbb"),
				[]int32("aaaaaaaaaaaabbbb"),
				[]int32("bbbaaaabbbaaaabab"),
				[]int32("xxxxxxxyyyyyyyyzzzz"),
			},
			prefix:  []int32("aba"),
			suffix:  []int32("wwababa"),
			expPref: []Index{{7, []int32{0}}},
			expSuf:  []Index{{10, []int32{9}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gsa := NewGSA_32(tc.text)
			if name == "empty text" {
				assert.Nil(t, gsa)
				return
			}
			assert.Equal(t, tc.expSuf, gsa.LookupSuffix(tc.suffix))
			assert.Equal(t, tc.expPref, gsa.LookupPrefix(tc.prefix))
		})
	}
}

func TestGSAWithSingleText(t *testing.T) {
	text := [][]int32{
		[]int32("abzababab"),
		[]int32("babaxyzab"),
		[]int32("jvoabbabrpvpabewge"),
		[]int32("wcccchervabgimeog"),
		[]int32("xqabqqqhfimmoabmhbaabfiq"),
		[]int32("cqoiwhoihabewqh"),
		[]int32("xxhoiababhehqab"),
		[]int32("qihcoiabhwca"),
		[]int32("qoixh79bbab"),
		[]int32("oihcqoihoieabicq"),
		[]int32("abababababababab"),
		[]int32("ociioimcwwwababa"),
		[]int32("aboiqhconhwiehcoiqwwfab"),
		[]int32("pqcpmwpeoicwq"),
		[]int32("mevmbxouccoiwq"),
		[]int32("bababicqqqqqqk"),
		[]int32("bbbbbbbbbbbbbbb"),
		[]int32("aaaaaaaaaaaabbbb"),
		[]int32("bbbaaaabbbaaaabab"),
		[]int32("xxxxxxxyyyyyyyyzzzz"),
	}
	tests := map[string]struct {
		prefix []int32
		exp    []Index
	}{
		"ab": {
			prefix: []int32("ab"),
			exp: []Index{
				{0, []int32{0, 3, 5, 7}},
				{1, []int32{1, 7}},
				{2, []int32{3, 6, 12}},
				{3, []int32{9}},
				{4, []int32{2, 13, 19}},
				{5, []int32{9}},
				{6, []int32{5, 7, 13}},
				{7, []int32{6}},
				{8, []int32{9}},
				{9, []int32{11}},
				{10, []int32{0, 2, 4, 6, 8, 10, 12, 14}},
				{11, []int32{11, 13}},
				{12, []int32{0, 21}},
				{15, []int32{1, 3}},
				{17, []int32{11}},
				{18, []int32{6, 13, 15}},
			},
		},
		"aba": {
			prefix: []int32("aba"),
			exp: []Index{
				{0, []int32{3, 5}},
				{1, []int32{1}},
				{6, []int32{5}},
				{10, []int32{0, 2, 4, 6, 8, 10, 12}},
				{11, []int32{11, 13}},
				{15, []int32{1}},
				{18, []int32{13}},
			},
		},
		"pmwpeo": {
			prefix: []int32("pmwpeo"),
			exp: []Index{
				{13, []int32{3}},
			},
		},
	}
	gsa := NewGSA_32(text)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.exp, gsa.LookupTextOrder(tc.prefix))
		})
	}
}

func BenchmarkGSALookup(b *testing.B) {
	tests := map[string]struct {
		text   [][]int32
		prefix []int32
	}{
		"single": {
			text:   [][]int32{[]int32("a")},
			prefix: []int32("a"),
		},
		"all same in one string": {
			text:   [][]int32{[]int32("aaaaaaa")},
			prefix: []int32("a"),
		},
		"all same in multiple strings": {
			text:   [][]int32{[]int32("aaaaaaa"), []int32("aaaaa")},
			prefix: []int32("a"),
		},
		"one different string": {
			text:   [][]int32{[]int32("abbacdababaaaaaab")},
			prefix: []int32("ab"),
		},
		"multiple strings with many occurrences": {
			text: [][]int32{
				[]int32("abzababab"),
				[]int32("babaxyzab"),
				[]int32("jvoabbabrpvpabewge"),
				[]int32("wcccchervabgimeog"),
				[]int32("xqabqqqhfimmoabmhbaabfiq"),
				[]int32("cqoiwhoihabewqh"),
				[]int32("xxhoiababhehqab"),
				[]int32("qihcoiabhwca"),
				[]int32("qoixh79bbab"),
				[]int32("oihcqoihoieabicq"),
				[]int32("abababababababab"),
				[]int32("ociioimcwwwababa"),
				[]int32("aboiqhconhwiehcoiqwwfab"),
				[]int32("pqcpmwpeoicwq"),
				[]int32("mevmbxouccoiwq"),
				[]int32("bababicqqqqqqk"),
				[]int32("bbbbbbbbbbbbbbb"),
				[]int32("aaaaaaaaaaaabbbb"),
				[]int32("bbbaaaabbbaaaabab"),
				[]int32("xxxxxxxyyyyyyyyzzzz"),
			},
			prefix: []int32("ab"),
		},
		"multiple strings with one occurrence": {
			text: [][]int32{
				[]int32("cnklnldskk"),
				[]int32("jwofjpppmcppppppppppw"),
				[]int32("oqccpowcccwq"),
				[]int32("poqcurmpowww"),
				[]int32("ouqcomopooew"),
				[]int32("cqoiwhoihewqh"),
				[]int32("xxhoihehq"),
				[]int32("qihcoihwc"),
				[]int32("qoixh79"),
				[]int32("oihcqoihoieicq"),
				[]int32("ociioimcwwwababa"),
				[]int32("oiqhconhwiehcoiqwwf"),
				[]int32("pqcpmwpeoicwq"),
				[]int32("mevmbxouccoiwq"),
				[]int32("bababicqqqqqqk"),
				[]int32("bbbbbbbbbbbbbbb"),
				[]int32("aaaaaaaaaaaabbbb"),
				[]int32("bbbaaaabbbaaaabab"),
				[]int32("xxxxxxxyyyyyyyyzzzz"),
			},
			prefix: []int32("pmwpeo"),
		},
	}
	for name, tc := range tests {
		gsa := NewGSA_32(tc.text)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				gsa.LookupTextOrder(tc.prefix)
			}
		})
	}
}

func TestNewGSA(t *testing.T) {
	text := [][]int32{
		[]int32("abzababab"),
		[]int32("babaxyzab"),
		[]int32("jvorpvpewge"),
		[]int32("wcccchervgimeog"),
		[]int32("xqqqqhfimmomhfiq"),
		[]int32("cqoiwhoihewqh"),
		[]int32("xxhoihehq"),
		[]int32("qihcoihwc"),
		[]int32("qoixh79"),
		[]int32("oihcqoihoieicq"),
		[]int32("abababababababab"),
		[]int32("ociioimcwwwababa"),
		[]int32("oiqhconhwiehcoiqwwf"),
		[]int32("pqcpmwpeoicwq"),
		[]int32("mevmbxouccoiwq"),
		[]int32("bababicqqqqqqk"),
		[]int32("bbbbbbbbbbbbbbb"),
		[]int32("aaaaaaaaaaaabbbb"),
		[]int32("bbbaaaabbbaaaabab"),
		[]int32("xxxxxxxyyyyyyyyzzzz"),
	}
	NewGSA_32(text)
}

func BenchmarkNewGSA_32(b *testing.B) {
	tests := map[string]struct {
		text [][]int32
	}{
		"single": {
			text: [][]int32{[]int32("a")},
		},
		"all same in one string": {
			text: [][]int32{[]int32("aaaaaaa")},
		},
		"all same in multiple strings": {
			text: [][]int32{[]int32("aaaaaaa"), []int32("aaaaa")},
		},
		"one different string": {
			text: [][]int32{[]int32("abbacdababaaaaaab")},
		},
		"multiple different strings": {
			text: [][]int32{
				[]int32("abzababab"),
				[]int32("babaxyzab"),
				[]int32("jvorpvpewge"),
				[]int32("wcccchervgimeog"),
				[]int32("xqqqqhfimmomhfiq"),
				[]int32("cqoiwhoihewqh"),
				[]int32("xxhoihehq"),
				[]int32("qihcoihwc"),
				[]int32("qoixh79"),
				[]int32("oihcqoihoieicq"),
				[]int32("abababababababab"),
				[]int32("ociioimcwwwababa"),
				[]int32("oiqhconhwiehcoiqwwf"),
				[]int32("pqcpmwpeoicwq"),
				[]int32("mevmbxouccoiwq"),
				[]int32("bababicqqqqqqk"),
				[]int32("bbbbbbbbbbbbbbb"),
				[]int32("aaaaaaaaaaaabbbb"),
				[]int32("bbbaaaabbbaaaabab"),
				[]int32("xxxxxxxyyyyyyyyzzzz"),
			},
		},
	}
	for name, tc := range tests {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NewGSA_32(tc.text)
			}
		})
	}
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
