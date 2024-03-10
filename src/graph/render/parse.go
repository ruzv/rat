package render

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/pkg/errors"
	"rat/graph/render/token"
)

// RatTokenNode markdown ast node for rat tokens.
type RatTokenNode struct {
	ast.Leaf
	Token *token.Token
}

// RatErrorNode markdown ast node for rat errors.
type RatErrorNode struct {
	ast.Leaf
	Err error
}

// Parse parses markdown content to AST, with rat unique ast nodes, like tokens.
func Parse(content string) ast.Node { //nolint:ireturn
	p := parser.NewWithExtensions(
		parser.NoIntraEmphasis |
			parser.Tables |
			parser.FencedCode |
			parser.Autolink |
			parser.Strikethrough |
			parser.SpaceHeadings |
			parser.HeadingIDs |
			parser.BackslashLineBreak |
			// parser.DefinitionLists | // not part of CommonMark
			parser.MathJax |
			parser.LaxHTMLBlocks |
			parser.AutoHeadingIDs |
			parser.Attributes |
			parser.SuperSubscript,
	)

	p.Opts.ParserHook = ratTokenHook

	prev := p.RegisterInline('<', nil)
	p.RegisterInline('<', ratTokenInlineHook(prev))

	return p.Parse([]byte(content))
}

// ratTokenHook is a markdown parser BlockFunc
// registered on the parser to parse block rat tokens (rat tokens that are
// stand alone, not in paragraph, or header or some other block).
// If successful it returns an ast.Node, a buffer that should be parsed as a
// block and the number of bytes consumed.
func ratTokenHook(data []byte) (ast.Node, []byte, int) { //nolint:ireturn
	astNode, consumed := parseRatTokenData(data)

	// return empty buffer, cause nothing in the returned AST node should be
	// further re-parset as block. its ready as is.
	return astNode, []byte{}, consumed
}

// Parameters:
//
//	p      - parser (lol)
//	data   - content of a block node
//	offset - offset for the data, where the inline parser should begin
//
// Returns:
//
//	int      - bytes consumed
//	ast.Node - parsed ast node
func ratTokenInlineHook(
	prev func(p *parser.Parser, data []byte, offset int) (int, ast.Node),
) func(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	return func(p *parser.Parser, original []byte, offset int) (int, ast.Node) {
		data := original[offset:]

		astNode, consumed := parseRatTokenData(data)
		if astNode == nil { // failed to find start/end markers for rat tokens.
			return prev(p, original, offset)
		}

		return consumed, astNode
	}
}

// Returns:
//
//	*RatTokenNode - node parsed
//	int           - bytes consumed
func parseRatTokenData(data []byte) (ast.Node, int) { //nolint:ireturn
	startMarker := []byte(`<rat `)
	endMarker := []byte(`/>`)

	if !bytes.HasPrefix(data, startMarker) {
		return nil, 0
	}

	end := bytes.Index(data[len(startMarker):], endMarker)
	if end < 0 {
		return nil, 0
	}

	// account for the fact that start marker prefix
	// was not included in the search
	end += len(startMarker)

	raw := string(data[len(startMarker):end])

	t, err := token.Parse(raw)
	if err != nil {
		return &RatErrorNode{
				Err: errors.Wrapf(
					err, `failed to parse "<rat %s/>" as rat token`, raw,
				),
			},
			end + len(endMarker)
	}

	return &RatTokenNode{Token: t}, end + len(endMarker)
}
