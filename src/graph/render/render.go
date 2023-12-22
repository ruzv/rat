package render

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/render/todo"
	"rat/graph/render/token"
	"rat/logr"
)

var _ jsonast.Renderer = (*JSONRenderer)(nil)

// JSONRenderer renders a nodes markdown content to JSON representation of the
// markdown AST.
type JSONRenderer struct {
	p   graph.Provider
	log *logr.LogR
}

// NewJSONRenderer creates a new JSONRenderer.
func NewJSONRenderer(
	log *logr.LogR,
	p graph.Provider,
) *JSONRenderer {
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

//nolint:cyclop,gocyclo,maintidx // big switch.
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
		p, err := jr.renderGraphLink(part, node, entering)
		if err != nil {
			return part.AddContainer( //nolint:nilerr,lll // render link if graph link fails.
				&jsonast.AstPart{
					Type: "link",
					Attributes: jsonast.AstAttributes{
						"title":       string(node.Title),
						"destination": string(node.Destination),
					},
				},
				entering,
			), nil
		}

		return p, nil
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
					"type": map[ast.ListType]string{
						ast.ListTypeOrdered:    "ordered",
						ast.ListTypeDefinition: "definition",
						ast.ListTypeTerm:       "term",
					}[node.ListFlags],
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
						"text": strings.TrimSpace(string(node.Literal)),
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
	case *ast.Strong:
		part = part.AddContainer(
			&jsonast.AstPart{Type: "strong"},
			entering,
		)
	case *ast.Image:
		part = part.AddContainer(
			&jsonast.AstPart{
				Type: "image",
				Attributes: jsonast.AstAttributes{
					"src": resolveFileURL(string(node.Destination)),
				},
			},
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

func (jr *JSONRenderer) renderGraphLink(
	part *jsonast.AstPart, link *ast.Link, entering bool,
) (*jsonast.AstPart, error) {
	id, err := uuid.FromString(string(link.Destination))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uuid")
	}

	destNode, err := jr.p.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by id")
	}

	dest, err := destNode.Path.ViewURL()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get view url")
	}

	title := string(link.Title)

	if title == "" {
		title = destNode.Name()
	}

	return part.AddContainer(
			&jsonast.AstPart{
				Type: "graph_link",
				Attributes: jsonast.AstAttributes{
					"title":       title,
					"destination": dest,
				},
			},
			entering,
		),
		nil
}

func resolveFileURL(file string) string {
	parsed, err := url.Parse(file)
	if err != nil {
		return file
	}

	if parsed.IsAbs() {
		return file
	}

	res, err := url.JoinPath("/graph/file/", file)
	if err != nil {
		return file
	}

	return res
}
