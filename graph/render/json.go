package render

import (
	"fmt"
	"strings"

	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/render/token"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/pkg/errors"
)

// listTypes maps ast.ListType to string.
var listTypes = map[ast.ListType]string{
	ast.ListTypeOrdered:    "ordered",
	ast.ListTypeDefinition: "definition",
	ast.ListTypeTerm:       "term",
}

func RenderJSON(n *graph.Node, p graph.Provider) (*jsonast.AstPart, error) {
	root := jsonast.NewRootAstPart()
	part := root

	var err error

	ast.WalkFunc(
		parser.NewWithExtensions(
			parser.NoIntraEmphasis|
				parser.Tables|
				parser.FencedCode|
				parser.Autolink|
				parser.Strikethrough|
				parser.SpaceHeadings|
				parser.HeadingIDs|
				parser.BackslashLineBreak|
				parser.DefinitionLists|
				parser.MathJax|
				parser.LaxHTMLBlocks|
				parser.AutoHeadingIDs|
				parser.Attributes|
				parser.SuperSubscript,
		).Parse([]byte(token.WrapContentTokens(n.Content))),
		func(node ast.Node, entering bool) ast.WalkStatus {
			part, err = renderNode(part, n, p, node, entering)
			if err != nil {
				return ast.Terminate
			}

			return ast.GoToNext
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed walk ast and render")
	}

	return root, nil
}

func renderNode(
	part *jsonast.AstPart,
	n *graph.Node,
	p graph.Provider,
	node ast.Node,
	entering bool,
) (*jsonast.AstPart, error) {
	switch node := node.(type) {
	case *ast.Document:
		if entering {
			part.Type = "document"
		}
	case *ast.Text:
		part.AddLeaf(
			&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.Link:
		part = part.AddContainer(
			&jsonast.AstPart{
				Type: "link",
				Attributes: jsonast.AstAttributes{
					"destination": string(node.Destination),
					"title":       string(node.Title),
				},
			},
			entering,
		)
	case *ast.List:
		fmt.Println("LIST", node.ListFlags)
		fmt.Println("LIST Type term", node.ListFlags&ast.ListTypeTerm)
		fmt.Println("LIST Type ordered", node.ListFlags&ast.ListTypeOrdered)
		fmt.Println("LIST Type definition", node.ListFlags&ast.ListTypeDefinition)

		part = part.AddContainer(
			&jsonast.AstPart{
				Type: "list",
				Attributes: jsonast.AstAttributes{
					"ordered":    node.ListFlags&ast.ListTypeOrdered != 0,
					"definition": node.ListFlags&ast.ListTypeDefinition != 0,
					"term":       node.ListFlags&ast.ListTypeTerm != 0,
				},
			},
			entering,
		)
	case *ast.ListItem:
		part = part.AddContainer(
			&jsonast.AstPart{
				Type: "list_item",
				Attributes: jsonast.AstAttributes{
					"type": listTypes[node.ListFlags],
				},
			},
			entering,
		)
	case *ast.Heading:
		part = part.AddContainer(
			&jsonast.AstPart{
				Type: "heading",
				Attributes: jsonast.AstAttributes{
					"level": node.Level,
				},
			},
			entering,
		)
	case *ast.HTMLSpan:
		part.AddLeaf(
			&jsonast.AstPart{
				Type: "span",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.HTMLBlock:
		literal := string(node.Literal)
		raw := strings.TrimPrefix(literal, "<div>")
		raw = strings.TrimSuffix(raw, "</div>")

		if strings.HasPrefix(literal, "<div>") &&
			strings.HasSuffix(literal, "</div>") &&
			token.IsToken(raw) {
			err := renderTokenNode(part, n, p, raw)
			if err != nil {
				return nil, errors.Wrap(err, "failed to render token node")
			}

			break
		}

		part.AddLeaf(
			&jsonast.AstPart{
				Type: "html_block",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
				},
			},
		)

	case *ast.Paragraph:
		part = part.AddContainer(
			&jsonast.AstPart{
				Type: "paragraph",
			},
			entering,
		)
	case *ast.Code:
		part.AddLeaf(
			&jsonast.AstPart{
				Type: "code",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.CodeBlock:
		part.AddLeaf(
			&jsonast.AstPart{
				Type: "code_block",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
					"info": string(node.Info),
				},
			},
		)
	default:
		if node.AsLeaf() == nil { // container
			part = part.AddContainer(
				&jsonast.AstPart{
					Type: "unknown",
					Attributes: jsonast.AstAttributes{
						"text": fmt.Sprintf("%T", node),
					},
				},
				entering,
			)
		} else {
			part.AddLeaf(
				&jsonast.AstPart{
					Type: "unknown",
					Attributes: jsonast.AstAttributes{
						"text": fmt.Sprintf("%T", node),
					},
				},
			)
		}
	}

	return part, nil
}

func renderTokenNode(
	part *jsonast.AstPart, n *graph.Node, p graph.Provider, rawToken string,
) error {
	t, err := token.NewToken(rawToken)
	if err != nil {
		return errors.Wrapf(err, "failed to parse token - %q", rawToken)
	}

	// TODO: pass render function here to allow todos and other shits to have
	// full markdown access.
	err = t.Render(part, n, p)
	if err != nil {
		return errors.Wrap(err, "failed to transform")
	}

	return nil
}
