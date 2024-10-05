package mathjax

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type mathJaxBlockParser struct {
}

var defaultMathJaxBlockParser = &mathJaxBlockParser{}

type mathBlockData struct {
	indent   int
	isInline bool
}

var mathBlockInfoKey = parser.NewContextKey()

func NewMathJaxBlockParser() parser.BlockParser {
	return defaultMathJaxBlockParser
}

func (b *mathJaxBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos == -1 {
		return nil, parser.NoChildren
	}

	// Check for multi-line math block
	if len(line) >= pos+2 && line[pos] == '$' && line[pos+1] == '$' {
		pc.Set(mathBlockInfoKey, &mathBlockData{indent: pos, isInline: false})
		node := NewMathBlock()
		return node, parser.NoChildren
	}

	return nil, parser.NoChildren
}

func (b *mathJaxBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	dataInterface := pc.Get(mathBlockInfoKey)
	if dataInterface == nil {
		return parser.Close
	}

	_, ok := dataInterface.(*mathBlockData)
	if !ok {
		return parser.Close
	}

	// Check for closing $$
	if bytes.HasPrefix(bytes.TrimSpace(line), []byte("$$")) {
		if !bytes.Equal(bytes.TrimSpace(line), []byte("$$")) {
			// If there's content after $$, we need to split it
			parts := bytes.SplitN(bytes.TrimSpace(line), []byte("$$"), 2)
			if len(parts) > 1 && len(parts[1]) > 0 {
				// Add the content after $$ to a new paragraph
				para := ast.NewParagraph()
				para.AppendChild(para, ast.NewTextSegment(text.NewSegment(segment.Start+bytes.Index(line, parts[1]), segment.Stop)))
				node.Parent().InsertAfter(node.Parent(), node, para)
			}
		}
		reader.Advance(segment.Stop - segment.Start)
		return parser.Close
	}

	node.(*MathBlock).Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (b *mathJaxBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	mathBlock, ok := node.(*MathBlock)
	if !ok {
		return
	}

	// Remove leading and trailing $$ from the content
	if mathBlock.Lines().Len() > 0 {
		firstLine := mathBlock.Lines().At(0)
		lastLine := mathBlock.Lines().At(mathBlock.Lines().Len() - 1)

		firstLineContent := reader.Value(firstLine)
		lastLineContent := reader.Value(lastLine)

		if bytes.HasPrefix(firstLineContent, []byte("$$")) {
			mathBlock.Lines().Set(0, text.NewSegment(firstLine.Start+2, firstLine.Stop))
		}
		if bytes.HasSuffix(lastLineContent, []byte("$$")) {
			mathBlock.Lines().Set(mathBlock.Lines().Len()-1, text.NewSegment(lastLine.Start, lastLine.Stop-2))
		}
	}

	pc.Set(mathBlockInfoKey, nil)
}

func (b *mathJaxBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *mathJaxBlockParser) CanAcceptIndentedLine() bool {
	return true
}

func (b *mathJaxBlockParser) Trigger() []byte {
	return []byte{'$'}
}
