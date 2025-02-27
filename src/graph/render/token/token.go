package token

import (
	"strings"
	"text/scanner"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/services/urlresolve"
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
	// Time token renders the result of different time calculations.
	Time Type = "time"
)

// ErrMissingArgument error returned when and argument is missing in token.
var (
	// ErrMissingArgument error returned when and argument is missing in token.
	ErrMissingArgument = errors.New("missing argument")
	// ErrUnknownTokenType error returned when token type is not recognised by
	// token parser.
	ErrUnknownTokenType = errors.New("unknown token type")
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

	tokenType, err := sf.MustPop()
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
		Type: Type(tokenType),
		Args: args,
	}, nil
}

// Render renders a token to JSON AST.
func (t *Token) Render(
	root *jsonast.AstPart,
	n *graph.Node,
	p graph.Provider,
	resolver *urlresolve.Resolver,
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
		return t.renderEmbed(root, resolver)
	case Version:
		renderVersion(root)

		return nil
	case Time:
		return t.renderTime(root)
	default:
		return errors.Wrapf(
			ErrUnknownTokenType, "unknown token type - %q", t.Type,
		)
	}
}
