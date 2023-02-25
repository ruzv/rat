package token

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"private/rat/graph"
	"private/rat/graph/render/todo"
	"private/rat/graph/util"

	"github.com/gofrs/uuid"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var tokenRegex = regexp.MustCompile(
	`<rat(\s([^>]+))>`,
)

var log = logging.MustGetLogger("graph")

// TransformContentTokens converts special tokens in content like
// <rat graph> <rat link> <rat todo>
// into markdown.
func TransformContentTokens(n *graph.Node, p graph.Provider) string {
	matches := tokenRegex.FindAllIndex([]byte(n.Content), -1)

	if len(matches) == 0 {
		return n.Content
	}

	var prevTokenEnd int

	parts := make([]string, 0, 2*len(matches)+1)

	for _, match := range matches {
		tokenStart := match[0]
		tokenEnd := match[1]

		res, err := func() (string, error) {
			t, err := NewToken(n.Content[tokenStart:tokenEnd])
			if err != nil {
				return "", errors.Wrap(err, "failed to create token")
			}

			res, err := t.Transform(n, p)
			if err != nil {
				return "", errors.Wrap(err, "failed to transform")
			}

			return res, nil
		}()
		if err != nil {
			log.Warningf("failed to tansform token: %s", err.Error())
			res = fmt.Sprintf("`TOKEN ERROR: %s`", err.Error())
		}

		parts = append(
			parts,
			n.Content[prevTokenEnd:tokenStart],
			res,
		)

		prevTokenEnd = tokenEnd
	}

	// last tokens end to end of content
	parts = append(parts, n.Content[matches[len(matches)-1][1]:])

	return strings.Join(parts, "")
}

// TokenType describes rat token types.
type TokenType string

const (
	// GraphTokenType graph tokens provide an overview of a nodes child nodes.
	// Graph tokens get substituted with a list tree of links to child nodes of
	// specified depth. Unlimited depth if omitted.
	GraphTokenType TokenType = "graph"
	// TodoTokenType todo searches for todos in child nodes and collects them
	// into a large singular todo. Token args can be used to specify search
	// options.
	TodoTokenType TokenType = "todo"
)

// Token describes a single rat token. Rat tokens are special html tab like
// strings that when present nodes content have special handling. Mode detailed
// explication of what each token type does is available at TokenType constat's
// definitions comments.
// Tokens are in form of:
// <rat type key=value key=value>
// - type - the tokens type
// - key=value - 0 or more key value pairs.
type Token struct {
	Type TokenType
	Args map[string]string
}

// NewToken attempts to create a new woken from raw string.
func NewToken(raw string) (*Token, error) {
	parts := strings.Fields(strings.Trim(raw, "<>"))

	if len(parts) <= 1 {
		return nil, errors.New("cannot create token without a type")
	}

	parts = parts[1:]

	tokenType, err := func(raw string) (TokenType, error) {
		target := TokenType(raw)

		for _, valid := range []TokenType{
			GraphTokenType,
			TodoTokenType,
		} {
			if target == valid {
				return target, nil
			}
		}

		return "", errors.Errorf("unknown token type %s", target)
	}(parts[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token type")
	}

	args := func(raw []string) map[string]string {
		args := make(map[string]string)

		for _, r := range raw {
			parts := strings.Split(r, "=")
			if len(parts) != 2 {
				continue
			}

			args[parts[0]] = parts[1]
		}

		return args
	}(parts[1:])

	return &Token{
		Type: tokenType,
		Args: args,
	}, nil
}

// Transform produces the expanded version of the token.
func (t *Token) Transform(n *graph.Node, p graph.Provider) (string, error) {
	transformers := map[TokenType]func(
		n *graph.Node, p graph.Provider,
	) (string, error){
		GraphTokenType: t.transformGraphToken,
		TodoTokenType:  t.transformTodoToken,
	}

	trans, ok := transformers[t.Type]
	if !ok {
		return "", errors.Errorf("unknown token type - %s", t.Type)
	}

	return trans(n, p)
}

func (t *Token) transformGraphToken(
	n *graph.Node,
	p graph.Provider,
) (string, error) {
	depth, err := t.getArgDepth()
	if err != nil {
		return "", errors.Wrap(err, "failed to get depth")
	}

	b := &bytes.Buffer{}

	var buffErr error

	err = n.Walk(
		p,
		func(d int, node *graph.Node) bool {
			if d == depth {
				return false
			}

			if buffErr != nil {
				return false
			}

			_, err := b.WriteString(
				fmt.Sprintf(
					"%s- %s \n",
					strings.Repeat("\t", d),
					util.Link(node.Path, node.Name),
				),
			)
			if err != nil {
				buffErr = err

				return false
			}

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	if buffErr != nil {
		return "", errors.Wrap(buffErr, "failed to prepare read buffer")
	}

	return b.String(), nil
}

func (t *Token) transformTodoToken(
	n *graph.Node,
	p graph.Provider,
) (string, error) {
	id, err := t.getArgParent(n.ID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get parent arg")
	}

	depth, err := t.getArgDepth()
	if err != nil {
		return "", errors.Wrap(err, "failed to get depth")
	}

	parent, err := p.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get parent node")
	}

	var (
		todos    []*todo.Todo
		parseErr error
	)

	err = parent.Walk(
		p,
		func(d int, node *graph.Node) bool {
			if d == depth || parseErr != nil {
				return false
			}

			nodeTodos, err := todo.ParseNode(node)
			if err != nil {
				parseErr = errors.Wrap(err, "failed to parse nodes todos")

				return false
			}

			todos = append(todos, nodeTodos...)

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	if parseErr != nil {
		return "", errors.Wrap(parseErr, "failed to parse todo")
	}

	return strings.Join(
			util.Map(
				todos,
				func(t *todo.Todo) string {
					return t.Markdown()
				},
			),
			"\n",
		),
		nil
}

// by default returns -1, meaning no depth limit.
func (t *Token) getArgDepth() (int, error) {
	depthArg, ok := t.Args["depth"]
	if !ok {
		return -1, nil
	}

	depth, err := strconv.Atoi(depthArg)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse depth")
	}

	if depth < 1 {
		return 0, errors.Errorf(
			"invalid depth - %d, depth must be positive", depth,
		)
	}

	return depth, nil
}

func (t *Token) getArgParent(defaultID uuid.UUID) (uuid.UUID, error) {
	parentArg, ok := t.Args["parent"]
	if !ok {
		return defaultID, nil
	}

	id, err := uuid.FromString(parentArg)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "failed to parse parent id")
	}

	return id, nil
}
