package mdast

import (
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// GetTextOfSubNodes returns the text of all sub-nodes of the given root.
func GetTextOfSubNodes(
	rootMarkdownNode ast.Node,
) string {
	var sb strings.Builder

	ast.WalkFunc(
		rootMarkdownNode,
		func(node ast.Node, entering bool) ast.WalkStatus {
			if !entering {
				return ast.GoToNext
			}

			leaf := node.AsLeaf()
			if leaf == nil {
				return ast.GoToNext
			}

			sb.Write(leaf.Literal)

			return ast.GoToNext
		},
	)

	return sb.String()
}
