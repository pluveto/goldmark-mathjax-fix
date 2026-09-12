package mathjax

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestMathBlockBoundaries(t *testing.T) {
	tests := []struct {
		name, source string
		formulas     []string
		tail         string
	}{
		{"display", "$$\nx = 1\n$$\n", []string{"x = 1"}, ""},
		{"adjacent", "$$\nx = 1\n$$\n$$\ny = 2\n$$\n", []string{"x = 1", "y = 2"}, ""},
		{"single line", "$$x = 1$$\n", []string{"x = 1"}, ""},
		{"single line at EOF", "$$x = 1$$", []string{"x = 1"}, ""},
		{"opening line content", "$$x = 1\ny = 2\n$$\n", []string{"x = 1\ny = 2"}, ""},
		{"single line tail", "$$x = 1$$ After *math*.\n", []string{"x = 1"}, "<p>After <em>math</em>.</p>"},
		{"closing line tail", "$$\nx = 1\n$$ After *math*.\n", []string{"x = 1"}, "<p>After <em>math</em>.</p>"},
		{"closing line tail at EOF", "$$\nx = 1\n$$ After *math*.", []string{"x = 1"}, "<p>After <em>math</em>.</p>"},
		{"content before closing", "$$\nx = 1$$ After.\n", []string{"x = 1"}, "<p>After.</p>"},
		{"list spaces", "- Item\n\n  $$\n  x = 1\n  $$\n\nAfter.\n", []string{"x = 1"}, "<p>After.</p>"},
		{"list tab", "- Item $n$\n\n\t$$\n\tn = p q\n\t$$\n\nAfter.\n", []string{"n = p q"}, "<p>After.</p>"},
		{"list tab single line", "- Item\n\n\t$$n=pq$$ After.\n", []string{"n=pq"}, "<p>After.</p>"},
		{"nested list", "- Outer\n  - Inner\n\n    $$\n    x=1\n    $$\n\nAfter.\n", []string{"x=1"}, "<p>After.</p>"},
		{"adjacent in quote", "> $$\n> x=1\n> $$\n> $$\n> y=2\n> $$\n", []string{"x=1", "y=2"}, ""},
		{"blockquote", "> $$\n> x = 1\n> $$\n\nAfter.\n", []string{"x = 1"}, "<p>After.</p>"},
		{"unclosed", "$$\nx = 1", []string{"x = 1"}, ""},
		{"escaped dollars", "$$x + \\$$ + y$$\n", []string{`x + \$$ + y`}, ""},
	}
	for _, tc := range tests {
		for _, eol := range []string{"LF", "CRLF"} {
			t.Run(tc.name+"/"+eol, func(t *testing.T) {
				src := tc.source
				if eol == "CRLF" {
					src = strings.ReplaceAll(src, "\n", "\r\n")
				}
				source := []byte(src)
				md := goldmark.New(goldmark.WithExtensions(MathJax))
				doc := md.Parser().Parse(text.NewReader(source))
				var formulas []string
				ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
					if block, ok := n.(*MathBlock); ok && entering {
						var value bytes.Buffer
						for i := 0; i < block.Lines().Len(); i++ {
							line := block.Lines().At(i)
							value.Write(line.Value(source))
						}
						formulas = append(formulas, strings.TrimSpace(strings.ReplaceAll(value.String(), "\r\n", "\n")))
					}
					return ast.WalkContinue, nil
				})
				if len(formulas) != len(tc.formulas) {
					t.Fatalf("formula count: got %q, want %q", formulas, tc.formulas)
				}
				for i := range formulas {
					if formulas[i] != tc.formulas[i] {
						t.Errorf("formula %d: got %q, want %q", i, formulas[i], tc.formulas[i])
					}
				}
				var out bytes.Buffer
				if err := md.Renderer().Render(&out, source, doc); err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(out.String(), tc.tail) {
					t.Errorf("missing trailing Markdown %q in %q", tc.tail, out.String())
				}
			})
		}
	}
}

func TestMathHTMLAndEscapes(t *testing.T) {
	for _, source := range []string{"$a < b & c > d$", "$$\na < b & c > d\n$$"} {
		out, err := renderMarkdown([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), "a &lt; b &amp; c &gt; d") {
			t.Errorf("TeX was not protected from HTML parsing: %s", out)
		}
	}
	out, err := renderMarkdown([]byte(`Before $x + \$5$ after.`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `\(x + \$5\)</span> after.`) {
		t.Errorf("escaped dollar prematurely closed inline math: %s", out)
	}
}

func TestMathDoesNotConsumeCode(t *testing.T) {
	source := "```text\n$x$\n$$y$$\n```\n\n`$z$`\n"
	out, err := renderMarkdown([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"$x$", "$$y$$", "<code>$z$</code>"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("missing literal code %q in %s", want, out)
		}
	}
	if strings.Contains(string(out), `class="math`) {
		t.Fatal("code was interpreted as math")
	}
}

func TestNeighboringInlineMath(t *testing.T) {
	for _, separator := range []string{" ", "+", ",", "\n", "\r\n"} {
		source := []byte("Before $a$" + separator + "$b$ after.")
		md := goldmark.New(goldmark.WithExtensions(MathJax))
		doc := md.Parser().Parse(text.NewReader(source))
		var formulas []string
		ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if math, ok := n.(*InlineMath); ok && entering {
				formulas = append(formulas, string(math.Text(source)))
			}
			return ast.WalkContinue, nil
		})
		if len(formulas) != 2 || formulas[0] != "a" || formulas[1] != "b" {
			t.Errorf("separator %q: got %q", separator, formulas)
		}
	}
}
