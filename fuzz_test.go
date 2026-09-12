//go:build go1.18
// +build go1.18

package mathjax

import (
	"io"
	"testing"

	"github.com/yuin/goldmark"
)

func FuzzMathParsing(f *testing.F) {
	for _, source := range []string{
		"$$\nx\n$$\n$$\ny\n$$",
		"- item\r\n\r\n\t$$x$$ after\r\n",
		"> $$\n> a < b & c\n> $$ tail\n",
		"$$x$$ tail\nnext line\n",
		"$x + \\$5$ and `$$y$$`",
		"$$\n\tx$$\n",
	} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		md := goldmark.New(goldmark.WithExtensions(MathJax))
		if err := md.Convert([]byte(source), io.Discard); err != nil {
			t.Fatal(err)
		}
	})
}
