package token

import (
	"fmt"
	"regexp"
	"strings"
	"text/scanner"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/util"
)

const (
	// GraphTokenType graph tokens provide an overview of a nodes child nodes.
	// Graph tokens get substituted with a list tree of links to child nodes of
	// specified depth. Unlimited depth if omitted.
	GraphTokenType Type = "graph"
	// TodoTokenType todo searches for todos in child nodes and collects them
	// into a large singular todo. Token args can be used to specify search
	// options.
	TodoTokenType Type = "todo"
	// KanbanTokenType kanban tokens provide a kanban board of child nodes.
	KanbanTokenType Type = "kanban"
)

var (
	errUnknownTokenType = errors.New("unknown token type")
	tokenRegex          = regexp.MustCompile(`<rat(?:\s((?:.|\s)+?))/>`)
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

// IsToken returns true if the raw string is a token.
func IsToken(raw string) bool {
	return tokenRegex.MatchString(raw)
}

// WrapContentTokens wraps tokens in content with <div></div> tags. Allowing
// markdown parser to parse them as HTML blocks.
func WrapContentTokens(content string) string {
	matches := tokenRegex.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		return content
	}

	var (
		prevTokenEnd int
		parts        = make([]string, 0, 2*len(matches)+1)
	)

	for _, match := range matches {
		tokenStart := match[0]
		tokenEnd := match[1]

		parts = append(
			parts,
			content[prevTokenEnd:tokenStart],
			fmt.Sprintf("<div>%s</div>", content[tokenStart:tokenEnd]),
		)

		prevTokenEnd = tokenEnd
	}

	// last tokens end to end of content
	parts = append(parts, content[matches[len(matches)-1][1]:])

	return strings.Join(parts, "")
}

// Render renders a token to JSON AST.
func Render(
	root *jsonast.AstPart,
	rawToken string,
	n *graph.Node,
	p graph.Provider,
	r jsonast.Renderer,
) error {
	t, err := parse(rawToken)
	if err != nil {
		return errors.Wrapf(err, "failed to parse token - %q", rawToken)
	}

	switch t.Type {
	case TodoTokenType:
		return t.renderTodo(root, p)
	case GraphTokenType:
		return t.renderGraph(root, n, p)
	case KanbanTokenType:
		return t.renderKanban(root, p, r)
	default:
		return errors.Errorf("unknown token type - %s", t.Type)
	}
}

// parse attempts to parse a new woken from raw string.
//
//nolint:cyclop,gocyclo //TODO: fix.
func parse(raw string) (*Token, error) {
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

	err := sf.PopParts("<", "rat")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse beginning of token")
	}

	rawTokenType, err := sf.MustPop()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token type")
	}

	tokenType, err := func(raw string) (Type, error) {
		target := Type(raw)

		for _, valid := range []Type{
			GraphTokenType,
			TodoTokenType,
			KanbanTokenType,
		} {
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

	for {
		err := sf.PopParts("/", ">")
		if err == nil {
			break
		}

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
