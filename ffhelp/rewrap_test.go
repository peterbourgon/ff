package ffhelp_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/fftest"
)

func TestRewrapAt(t *testing.T) {
	t.Parallel()

	for i, testcase := range []struct {
		s    string
		max  int
		want string
	}{
		{
			s:    loremIpsum120,
			max:  80,
			want: loremIpsum80,
		},
		{
			s:    loremIpsum40,
			max:  120,
			want: loremIpsum120,
		},
		{
			s:    paragraphs80,
			max:  120,
			want: paragraphs120,
		},
		{
			s:    paragraphs120,
			max:  80,
			want: paragraphs80,
		},
		{
			s:    paragraphs80Split,
			max:  120,
			want: paragraphs120,
		},
	} {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			if want, have := testcase.want, ffhelp.RewrapAt(testcase.s, testcase.max); want != have {
				t.Errorf("\n%s", fftest.DiffString(want, have))
			}
		})
	}
}

var loremIpsum40 = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate,
vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia
nostra, per inceptos himenaeos. Mauris
venenatis felis orci, ac consectetur mi
molestie ac. Integer pharetra pharetra
odio. Maecenas metus eros, viverra eget
efficitur ut, feugiat in tortor.
`)

var loremIpsum80 = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Mauris
venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra pharetra
odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in tortor.
`)

var loremIpsum120 = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros, vestibulum at pulvinar vulputate, vehicula id
lacus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Mauris venenatis
felis orci, ac consectetur mi molestie ac. Integer pharetra pharetra odio. Maecenas metus eros, viverra eget efficitur
ut, feugiat in tortor.
`)

var paragraphs120 = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros, vestibulum at pulvinar vulputate, vehicula id
lacus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros, vestibulum at pulvinar vulputate, vehicula id
lacus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Mauris venenatis
felis orci, ac consectetur mi molestie ac. Integer pharetra pharetra odio. Maecenas metus eros, viverra eget efficitur
ut, feugiat in tortor.
`)

var paragraphs80 = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Mauris
venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra pharetra
odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in tortor.
`)

var paragraphs80Split = strings.TrimSpace(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.



Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam diam eros,
vestibulum at pulvinar vulputate, vehicula id lacus. Class aptent taciti
sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Mauris
venenatis felis orci, ac consectetur mi molestie ac. Integer pharetra pharetra
odio. Maecenas metus eros, viverra eget efficitur ut, feugiat in tortor.
`)
