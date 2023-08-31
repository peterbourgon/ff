package fftest

import (
	"strings"
)

// DiffString produces a git-like diff of two multi-line strings.
func DiffString(a, b string) string {
	var (
		chunks = diffChunks(strings.Split(a, "\n"), strings.Split(b, "\n"))
		lines  = make([]string, 0, len(chunks))
	)
	for _, c := range chunks {
		lines = append(lines, c.String())
	}
	return strings.Join(lines, "\n")
}

// diffChunks adapted from github.com/kylelemons/godebug.
func diffChunks(a, b []string) []chunk {
	var (
		alen    = len(a)
		blen    = len(b)
		maxPath = alen + blen
	)
	if maxPath == 0 {
		return nil
	}

	var (
		v  = make([]int, 2*maxPath+1)
		vs = make([][]int, 0, 8)

		save = func() {
			dup := make([]int, len(v))
			copy(dup, v)
			vs = append(vs, dup)
		}
	)

	d := func() int {
		var d int
		for d = 0; d <= maxPath; d++ {
			for diag := -d; diag <= d; diag += 2 { // k
				var aindex int // x
				switch {
				case diag == -d:
					aindex = v[maxPath-d+1] + 0
				case diag == d:
					aindex = v[maxPath+d-1] + 1
				case v[maxPath+diag+1] > v[maxPath+diag-1]:
					aindex = v[maxPath+diag+1] + 0
				default:
					aindex = v[maxPath+diag-1] + 1
				}
				bindex := aindex - diag // y
				for aindex < alen && bindex < blen && a[aindex] == b[bindex] {
					aindex++
					bindex++
				}
				v[maxPath+diag] = aindex
				if aindex >= alen && bindex >= blen {
					save()
					return d
				}
			}
			save()
		}
		return d
	}()
	if d == 0 {
		return nil
	}

	var (
		chunks = make([]chunk, d+1)
		x, y   = alen, blen
	)
	for d := d; d > 0; d-- {
		var (
			endpoint = vs[d]
			diag     = x - y
			insert   = diag == -d || (diag != d && endpoint[maxPath+diag-1] < endpoint[maxPath+diag+1])
			x1       = endpoint[maxPath+diag]
			kk       int
			x0       int
			y0       int
			xM       int
		)
		if insert {
			kk = diag + 1
			x0 = endpoint[maxPath+kk]
			y0 = x0 - kk
			xM = x0
		} else {
			kk = diag - 1
			x0 = endpoint[maxPath+kk]
			y0 = x0 - kk
			xM = x0 + 1
		}

		var c chunk
		if insert {
			c.added = b[y0:][:1]
		} else {
			c.deleted = a[x0:][:1]
		}
		if xM < x1 {
			c.equal = a[xM:][:x1-xM]
		}

		chunks[d] = c
		x, y = x0, y0
	}

	if x > 0 {
		chunks[0].equal = a[:x]
	}

	if chunks[0].empty() {
		chunks = chunks[1:]
	}

	if len(chunks) == 0 {
		return nil
	}

	return chunks
}

type chunk struct {
	added   []string
	deleted []string
	equal   []string
}

func (c *chunk) empty() bool {
	return len(c.added) == 0 && len(c.deleted) == 0 && len(c.equal) == 0
}

func (c *chunk) String() string {
	var lines []string
	for _, s := range c.added {
		lines = append(lines, "+ "+s)
	}
	for _, s := range c.deleted {
		lines = append(lines, "- "+s)
	}
	for _, s := range c.equal {
		lines = append(lines, "  "+s)
	}
	return strings.Join(lines, "\n")
}
