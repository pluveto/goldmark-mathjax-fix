goldmark-mathjax
=========================

goldmark-mathjax is an extension for the [goldmark](http://github.com/yuin/goldmark) 
that adds both block math and inline math support

It translate inline math equation quoted by `$` and display math block quoted by `$$` into MathJax compatible format.
hyphen `_` won't break LaTeX render within a math element any more.

```
$$
\left[ \begin{array}{a} a^l_1 \\ ⋮ \\ a^l_{d_l} \end{array}\right]
= \sigma(
 \left[ \begin{matrix} 
    w^l_{1,1} & ⋯  & w^l_{1,d_{l-1}} \\  
    ⋮ & ⋱  & ⋮  \\ 
    w^l_{d_l,1} & ⋯  & w^l_{d_l,d_{l-1}} \\  
 \end{matrix}\right]  ·
 \left[ \begin{array}{x} a^{l-1}_1 \\ ⋮ \\ ⋮ \\ a^{l-1}_{d_{l-1}} \end{array}\right] + 
 \left[ \begin{array}{b} b^l_1 \\ ⋮ \\ b^l_{d_l} \end{array}\right])
 $$
```


Borrow the idea from pandoc and this [blackfriday PR](https://github.com/russross/blackfriday/pull/412/)

The implementation is heavily inspired by the Fenced Code Block and CodeSpan of goldmark

Maintained fork
--------------------

This fork continues pluveto's `feat/intended-block` fixes. Import this module
directly; no replacement of the upstream module is required.

Supported delimiters are `$...$` for inline math and `$$...$$` for display
math (on one line or across multiple lines). Display blocks work in lists and
blockquotes, with LF or CRLF input, and may be adjacent without a blank line.
Text after a closing display delimiter is parsed as Markdown. Escaped dollar
signs do not close math. Fenced and inline code remain literal. An unclosed
display block extends to the end of its container, like a fenced code block.
TeX is escaped when written to HTML so `<`, `>` and `&` reach MathJax intact.

The parser does not validate TeX or load the browser's MathJax runtime.

Installation
--------------------

```
go get github.com/pluveto/goldmark-mathjax-fix@codex/fix-math-parsing
```

Usage
--------------------

The fixes currently live on `codex/fix-math-parsing`, based on
`feat/intended-block`. Go records an immutable pseudo-version in `go.mod`.

```go
package main

import (
	"bytes"
	"fmt"

	mathjax "github.com/pluveto/goldmark-mathjax-fix"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func main() {
	md := goldmark.New(
		goldmark.WithExtensions(mathjax.MathJax),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// todo more control on the parsing process
	var html bytes.Buffer
	mdContent := []byte(`
$$
\mathbb{E}(X) = \int x d F(x) = \left\{ \begin{aligned} \sum_x x f(x) \; & \text{ if } X \text{ is discrete} 
\\ \int x f(x) dx \; & \text{ if } X \text{ is continuous }
\end{aligned} \right.
$$


Inline math $\frac{1}{2}$
`)
	if err := md.Convert(mdContent, &html); err != nil {
		fmt.Println(err)
	}
	fmt.Println(html.String())
}
```

License
--------------------
MIT

