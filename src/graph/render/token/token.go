package token

import (
	"strings"
	"text/scanner"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/util"
)

const (
	// Graph tokens provide an overview of a nodes child nodes.
	// Graph tokens get substituted with a list tree of links to child nodes of
	// specified depth. Unlimited depth if omitted.
	Graph Type = "graph"
	// Todo searches for todos in child nodes and collects them
	// into a large singular todo. Token args can be used to specify search
	// options.
	//
	//nolint:godox
	Todo Type = "todo"
	// Kanban tokens provide a kanban board of child nodes.
	Kanban Type = "kanban"
	// Embed tokens allow embeding links.
	Embed Type = "embed"
	// Version token renders the rat server version as a in line code ast node.
	Version Type = "version"
)

var allTypes = []Type{ //nolint:gochecknoglobals
	Graph,
	Todo,
	Kanban,
	Embed,
	Version,
}

var (
	errUnknownTokenType = errors.New("unknown token type")
	errMissingArgument  = errors.New("missing argument")
)

// Type describes rat token types.
type Type string

// Token describes a single rat token. Rat tokens are special html tab like
// strings that when present nodes content have special handling. Mode detailed
// explication of what each token type does is available at TokenType constat's
// definitions comments.
// Tokens are in form of:
// <rat type key=value key=value>
// - type - the tokens type
// - key=value - 0 or more key value pairs.
type Token struct {
	Type Type
	Args map[string]string
}

// Parse attempts to parse a new woken from raw string. The raw string commes
// from rat markdown AST parser and should not have start and end markers.
//
//nolint:cyclop,gocyclo
func Parse(raw string) (*Token, error) {
	s := &scanner.Scanner{}
	s.Init(strings.NewReader(strings.ReplaceAll(raw, "\"", "`")))

	var parts []string

	for {
		if s.Scan() == scanner.EOF {
			break
		}

		parts = append(parts, s.TokenText())
	}

	sf := util.NewStringFeed(parts)

	rawTokenType, err := sf.MustPop()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token type")
	}

	tokenType, err := func(raw string) (Type, error) {
		target := Type(raw)

		for _, valid := range allTypes {
			if target == valid {
				return target, nil
			}
		}

		return "", errors.Wrapf(
			errUnknownTokenType, "unknown token type %s", target,
		)
	}(rawTokenType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token type")
	}

	args := make(map[string]string)

	for sf.More() {
		key, err := sf.MustPop()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get arg key")
		}

		err = sf.PopParts("=")
		if err != nil {
			return nil, errors.Wrap(err, "unexpected token in arg assignment")
		}

		value, err := sf.MustPop()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get arg value")
		}

		args[key] = strings.Trim(value, "`")
	}

	return &Token{
		Type: tokenType,
		Args: args,
	}, nil
}

// Render renders a token to JSON AST.
func (t *Token) Render(
	root *jsonast.AstPart,
	n *graph.Node,
	p graph.Provider,
	r jsonast.Renderer,
) error {
	switch t.Type {
	case Todo:
		return t.renderTodo(root, n, p, r)
	case Graph:
		return t.renderGraph(root, n, p)
	case Kanban:
		return t.renderKanban(root, p, r)
	case Embed:
		return t.renderEmbed(root)
	case Version:
		renderVersion(root)

		return nil
	default:
		return errors.Errorf("unknown token type - %s", t.Type)
	}
}
