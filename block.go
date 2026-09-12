package mathjax

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type mathJaxBlockParser struct{}

var defaultMathJaxBlockParser = &mathJaxBlockParser{}

func NewMathJaxBlockParser() parser.BlockParser { return defaultMathJaxBlockParser }

// mathClosure finds an unescaped pair of dollars, not part of a longer run.
func mathClosure(line []byte, start int) int {
	for i := start; i+1 < len(line); i++ {
		if line[i] != '$' || line[i+1] != '$' {
			continue
		}
		if (i > 0 && line[i-1] == '$') || (i+2 < len(line) && line[i+2] == '$') {
			continue
		}
		if !escapedDollar(line, i) {
			return i
		}
	}
	return -1
}

func escapedDollar(line []byte, offset int) bool {
	backslashes := 0
	for i := offset - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 != 0
}

func (b *mathJaxBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || pos+2 > len(line) || !bytes.HasPrefix(line[pos:], []byte("$$")) ||
		(pos+2 < len(line) && line[pos+2] == '$') {
		return nil, parser.NoChildren
	}
	node := NewMathBlock()
	start := pos + 2
	if close := mathClosure(line, start); close >= 0 {
		node.Lines().Append(text.NewSegment(segment.Start-segment.Padding+start, segment.Start-segment.Padding+close))
		node.closed = true
		// Open can return only one block. Save trailing Markdown as source lines
		// and insert a paragraph on Close, before Goldmark's inline parsing pass.
		tail := close + 2
		for tail < len(line) && (line[tail] == ' ' || line[tail] == '\t') {
			tail++
		}
		if len(bytes.TrimSpace(line[tail:])) != 0 {
			seg := text.NewSegment(segment.Start-segment.Padding+tail, segment.Stop)
			for seg.Stop > seg.Start && (reader.Source()[seg.Stop-1] == '\r' || reader.Source()[seg.Stop-1] == '\n') {
				seg.Stop--
			}
			node.tail = &seg
		}
	} else if len(bytes.TrimSpace(line[start:])) != 0 {
		node.Lines().Append(text.NewSegment(segment.Start-segment.Padding+start, segment.Stop))
	}
	return node, parser.NoChildren
}

func (b *mathJaxBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	block := node.(*MathBlock)
	if block.closed {
		return parser.Close
	}
	line, segment := reader.PeekLine()
	if close := mathClosure(line, 0); close >= 0 {
		if len(bytes.TrimSpace(line[:close])) != 0 {
			prefix := segment
			prefix.Stop = segment.Start - segment.Padding + close
			block.Lines().Append(prefix)
		}
		// Leave both the line ending and any trailing Markdown for Goldmark.
		// Consuming the newline here skips into the next formula's opener.
		reader.Advance(close + 2)
		block.closed = true
		return parser.Close
	}
	// Reader segments already account for list/blockquote indentation and
	// virtual tab padding. A second dedent can truncate actual TeX characters.
	block.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (b *mathJaxBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	block := node.(*MathBlock)
	if block.tail != nil {
		paragraph := ast.NewParagraph()
		paragraph.Lines().Append(*block.tail)
		parent := block.Parent()
		parent.InsertAfter(parent, block, paragraph)
		block.tail = nil
	}
}

func (b *mathJaxBlockParser) CanInterruptParagraph() bool { return true }
func (b *mathJaxBlockParser) CanAcceptIndentedLine() bool { return true }
func (b *mathJaxBlockParser) Trigger() []byte             { return []byte{'$'} }
