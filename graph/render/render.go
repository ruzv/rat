package render

import (
	"fmt"
	"strings"

	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/render/todo"
	"rat/graph/render/token"
	"rat/logr"

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

var _ jsonast.Renderer = (*JSONRenderer)(nil)

// JSONRenderer renders a nodes markdown content to JSON representation of the
// markdown AST.
type JSONRenderer struct {
	p   graph.Provider
	log *logr.LogR
}

// NewJSONRenderer creates a new JSONRenderer.
func NewJSONRenderer(p graph.Provider, log *logr.LogR) *JSONRenderer {
	return &JSONRenderer{
		p:   p,
		log: log.Prefix("json-renderer"),
	}
}

// Render renders the markdown content of the specified node to JSON.
func (jr *JSONRenderer) Render(
	root *jsonast.AstPart, n *graph.Node,
) error {
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
			part, err = jr.renderNode(part, n, node, entering)
			if err != nil {
				return ast.Terminate
			}

			return ast.GoToNext
		},
	)

	if err != nil {
		return errors.Wrap(err, "failed walk ast and render")
	}

	return nil
}

//nolint:cyclop,gocyclo
func (jr *JSONRenderer) renderNode(
	part *jsonast.AstPart,
	n *graph.Node,
	node ast.Node,
	entering bool,
) (*jsonast.AstPart, error) {
	switch node := node.(type) {
	case *ast.Document:
		part = part.AddContainer(&jsonast.AstPart{Type: "document"}, entering)
	case *ast.Text:
		part.AddLeaf(
			&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.HorizontalRule:
		part.AddLeaf(&jsonast.AstPart{Type: "horizontal_rule"})
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
			err := token.Render(part, raw, n, jr.p, jr)
			if err != nil {
				return nil, errors.Wrap(err, "failed to render token")
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
		_, ok := node.GetParent().(*ast.ListItem)
		if !ok { // render paragraphs that are not part of lists.
			part = part.AddContainer(
				&jsonast.AstPart{
					Type: "paragraph",
				},
				entering,
			)
		}
	case *ast.Code:
		jr.log.Infof("%q", string(node.Literal))

		part.AddLeaf(
			&jsonast.AstPart{
				Type: "code",
				Attributes: jsonast.AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.CodeBlock:
		switch string(node.Info) {
		case "graphviz":
			part.AddLeaf(
				&jsonast.AstPart{
					Type: "graphviz",
					Attributes: jsonast.AstAttributes{
						"text": string(node.Literal),
					},
				},
			)
		case "todo":
			t, err := todo.Parse(string(node.Literal))
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse todo")
			}

			t.Render(part)
		default:
			part.AddLeaf(
				&jsonast.AstPart{
					Type: "code_block",
					Attributes: jsonast.AstAttributes{
						"text": string(node.Literal),
						"info": string(node.Info),
					},
				},
			)
		}
	case *ast.Table:
		part = part.AddContainer(
			&jsonast.AstPart{Type: "table"},
			entering,
		)
	case *ast.TableHeader:
		part = part.AddContainer(
			&jsonast.AstPart{Type: "table_header"},
			entering,
		)
	case *ast.TableBody:
		part = part.AddContainer(
			&jsonast.AstPart{Type: "table_body"},
			entering,
		)
	case *ast.TableRow:
		part = part.AddContainer(
			&jsonast.AstPart{Type: "table_row"},
			entering,
		)
	case *ast.TableCell:
		part = part.AddContainer(
			&jsonast.AstPart{Type: "table_cell"},
			entering,
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
