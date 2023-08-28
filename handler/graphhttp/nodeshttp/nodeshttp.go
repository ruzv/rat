package nodeshttp

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"rat/graph"
	"rat/graph/util"
	pathutil "rat/graph/util/path"
	"rat/handler/shared"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var (
	log = logging.MustGetLogger("nodeshttp")

	// listTypes maps ast.ListType to string.
	listTypes = map[ast.ListType]string{
		ast.ListTypeOrdered:    "ordered",
		ast.ListTypeDefinition: "definition",
		ast.ListTypeTerm:       "term",
	}
)

type handler struct {
	ss *shared.Services
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, ss *shared.Services,
) error {
	h := &handler{
		ss: ss,
	}

	nodesRouter := router.PathPrefix("/nodes").Subrouter()

	nodesRouter.HandleFunc("/", shared.Wrap(h.read)).Methods(http.MethodGet)
	nodesRouter.HandleFunc("/", shared.Wrap(h.create)).Methods(http.MethodPost)

	pathRe := regexp.MustCompile(
		`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll
	)

	nodeRouter := nodesRouter.
		PathPrefix(fmt.Sprintf("/{path:%s}", pathRe.String())).
		Subrouter()

	nodeRouter.HandleFunc("/", shared.Wrap(h.deconstruct)).
		Methods(http.MethodGet)
	nodeRouter.HandleFunc("/", shared.Wrap(h.create)).Methods(http.MethodPost)

	return nil
}

// AstAttributes describes a abstract syntax tree part attributes.
type AstAttributes map[string]any

// AstPart describes a abstract syntax tree part.
type AstPart struct {
	Type       string        `json:"type"`
	Attributes AstAttributes `json:"attributes,omitempty"`
	Children   []*AstPart    `json:"children,omitempty"`
	parent     *AstPart
}

// AddLeaf adds a leaf (a ast part that can not contain other ast parts) to
// the ast part.
func (p *AstPart) AddLeaf(leaf *AstPart) {
	leaf.parent = p
	p.Children = append(p.Children, leaf)
}

// AddContainer on entering adds a container (a ast part that can contain other
// ast parts) to the ast part, on exit moves the target part back one parent.
func (p *AstPart) AddContainer(child *AstPart, entering bool) *AstPart {
	if !entering { // exiting
		return p.parent
	}

	child.parent = p
	p.Children = append(p.Children, child)

	return child
}

// RenderNode parses a single ast node into a struct, that after parsing can
// be serialised into JSON. Used int ast.WalkFunc.
//
//nolint:gocyclo,cyclop
func (p *AstPart) RenderNode(node ast.Node, entering bool) *AstPart {
	switch node := node.(type) {
	case *ast.Document:
		if entering {
			p.Type = "document"
		}
	case *ast.Text:
		p.AddLeaf(
			&AstPart{
				Type: "text",
				Attributes: AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.Link:
		p = p.AddContainer(
			&AstPart{
				Type: "link",
				Attributes: AstAttributes{
					"destination": string(node.Destination),
					"title":       string(node.Title),
				},
			},
			entering,
		)
	case *ast.List:
		p = p.AddContainer(
			&AstPart{
				Type: "list",
				Attributes: AstAttributes{
					"type": listTypes[node.ListFlags],
				},
			},
			entering,
		)
	case *ast.ListItem:
		p = p.AddContainer(
			&AstPart{
				Type: "list_item",
				Attributes: AstAttributes{
					"type": listTypes[node.ListFlags],
				},
			},
			entering,
		)
	case *ast.Heading:
		p = p.AddContainer(
			&AstPart{
				Type: "heading",
				Attributes: AstAttributes{
					"level": node.Level,
				},
			},
			entering,
		)
	case *ast.HTMLSpan:
		p.AddLeaf(
			&AstPart{
				Type: "span",
				Attributes: AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.Paragraph:
		p = p.AddContainer(
			&AstPart{
				Type: "paragraph",
			},
			entering,
		)
	case *ast.Code:
		p.AddLeaf(
			&AstPart{
				Type: "code",
				Attributes: AstAttributes{
					"text": string(node.Literal),
				},
			},
		)
	case *ast.CodeBlock:
		p.AddLeaf(
			&AstPart{
				Type: "code_block",
				Attributes: AstAttributes{
					"text": string(node.Literal),
					"info": string(node.Info),
				},
			},
		)
	default:
		if node.AsLeaf() == nil { // container
			p = p.AddContainer(
				&AstPart{
					Type: "unknown",
					Attributes: AstAttributes{
						"text": fmt.Sprintf("%T", node),
					},
				},
				entering,
			)
		} else {
			p.AddLeaf(
				&AstPart{
					Type: "unknown",
					Attributes: AstAttributes{
						"text": fmt.Sprintf("%T", node),
					},
				},
			)
		}
	}

	return p
}

func (h *handler) deconstruct(w http.ResponseWriter, r *http.Request) error {
	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node error")
	}

	childNodePaths, err := h.getChildNodes(w, n.Path)
	if err != nil {
		return errors.Wrap(err, "failed to get child node paths")
	}

	rootPart := &AstPart{}
	rootPart.parent = rootPart

	part := rootPart

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
		).Parse([]byte(n.Content)),
		func(node ast.Node, entering bool) ast.WalkStatus {
			part = part.RenderNode(node, entering)

			return ast.GoToNext
		},
	)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = shared.WriteResponse(
		w,
		http.StatusOK,
		struct {
			ID             uuid.UUID           `json:"id"`
			Name           string              `json:"name"`
			Path           pathutil.NodePath   `json:"path"`
			ChildNodePaths []pathutil.NodePath `json:"childNodePaths"`
			Length         int                 `json:"length"`
			AST            *AstPart            `json:"ast"`
		}{
			ID:             n.ID,
			Name:           n.Name,
			Path:           n.Path,
			Length:         len(strings.Split(n.Content, "\n")),
			ChildNodePaths: childNodePaths,
			AST:            rootPart,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) error {
	body, err := shared.Body[struct {
		Name string `json:"name" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node error")
	}

	_, err = n.AddLeaf(h.ss.Graph, body.Name)
	if err != nil {
		shared.WriteError(
			w, http.StatusInternalServerError, "failed to create node",
		)

		return errors.Wrap(err, "failed to create node")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *handler) read(w http.ResponseWriter, r *http.Request) error {
	resp := struct {
		graph.Node
		Leafs []pathutil.NodePath `json:"leafs,omitempty"`
	}{}

	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	resp.Node = *n
	resp.Node.Content = h.ss.Renderer.Render(n)

	if includeLeafs(r) {
		leafs, err := h.getChildNodes(w, n.Path)
		if err != nil {
			return errors.Wrap(err, "failed to get leaf paths")
		}

		resp.Leafs = leafs
	}

	err = shared.WriteResponse(w, http.StatusOK, resp)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) getChildNodes(
	w http.ResponseWriter,
	path pathutil.NodePath,
) ([]pathutil.NodePath, error) {
	childNodes, err := h.ss.Graph.GetLeafs(path)
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get leafs",
		)

		return nil, errors.Wrap(err, "failed to get leafs")
	}

	return util.Map(
			childNodes,
			func(n *graph.Node) pathutil.NodePath {
				return n.Path
			},
		),
		nil
}

func includeLeafs(r *http.Request) bool {
	leafsParam := r.URL.Query().Get("leafs")
	if leafsParam == "" {
		return false
	}

	l, err := strconv.ParseBool(leafsParam)
	if err != nil {
		log.Debug("failed to parse leafs param", err)

		return false
	}

	return l
}

func (h *handler) getNode(
	w http.ResponseWriter,
	r *http.Request,
) (*graph.Node, error) {
	path := mux.Vars(r)["path"]

	var (
		n   *graph.Node
		err error
	)

	if path == "" {
		n, err = h.ss.Graph.Root()
		if err != nil {
			shared.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	} else {
		n, err = h.ss.Graph.GetByPath(pathutil.NodePath(path))
		if err != nil {
			shared.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	}

	return n, nil
}
