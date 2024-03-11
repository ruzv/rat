package render

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown/ast"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/render/todo"
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
	root *jsonast.AstPart, n *graph.Node, data string,
) {
	jr.log.Debugf("rendering for node %q %s", n.Path, logr.Preview(data))

	part := root

	ast.WalkFunc(
		Parse(data),
		func(node ast.Node, entering bool) ast.WalkStatus {
			newPart, err := jr.renderNode(part, n, node, entering)
			if err != nil {
				part.AddLeaf(
					&jsonast.AstPart{
						Type: "rat_error",
						Attributes: jsonast.AstAttributes{
							"err": err.Error(),
						},
					},
				)

				return ast.GoToNext
			}

			part = newPart

			return ast.GoToNext
		},
	)
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
	case *RatTokenNode:
		err := node.Token.Render(part, n, jr.p, jr)
		if err != nil {
			return nil, errors.Wrapf(
				err, "failed to render %q token", node.Token.Type,
			)
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
		listType := "unordered"
		if node.ListFlags&ast.ListTypeOrdered != 0 {
			listType = "ordered"
		}

		part = part.AddContainer(
			&jsonast.AstPart{
				Type:       "list",
				Attributes: jsonast.AstAttributes{"type": listType},
			},
			entering,
		)
	case *ast.ListItem:
		listItemType := "unordered"
		if node.ListFlags&ast.ListTypeOrdered != 0 {
			listItemType = "ordered"
		}

		part = part.AddContainer(
			&jsonast.AstPart{
				Type:       "list_item",
				Attributes: jsonast.AstAttributes{"type": listItemType},
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
		jr.log.Infof("html block %s", logr.Preview(string(node.Literal)))
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
			jr.log.Debugf(
				"parsing todo %s",
				logr.Preview(string(node.Literal)),
			)

			t, err := todo.Parse(string(node.Literal))
			if err != nil {
				jr.log.Errorf("failed to parse todo: %s", err)

				return nil, errors.Wrap(err, "failed to parse todo")
			}

			jr.log.Debugf(
				"rendering todo %s",
				logr.Preview(string(node.Literal)),
			)

			t.Render(part, n, jr)

			jr.log.Debugf("rendered todo")
		default:
			part.AddLeaf(
				&jsonast.AstPart{
					Type: "code_block",
					Attributes: jsonast.AstAttributes{
						"text":     strings.TrimSpace(string(node.Literal)),
						"language": string(node.Info),
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
	case *RatErrorNode:
		// for when markdown parser for rat tokens faild with something
		part.AddLeaf(
			&jsonast.AstPart{
				Type: "rat_error",
				Attributes: jsonast.AstAttributes{
					"err": node.Err.Error(),
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

	linkPart := part.AddContainer(
		&jsonast.AstPart{
			Type: "graph_link",
			Attributes: jsonast.AstAttributes{
				"destination": dest,
			},
		},
		entering,
	)
	if entering &&
		strings.TrimSpace(string(link.Title)) == "" &&
		len(link.GetChildren()) == 0 {
		linkPart.AddLeaf(
			&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": destNode.Name(),
				},
			},
		)
	}

	return linkPart, nil
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
