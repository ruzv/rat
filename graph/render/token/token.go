package token

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/scanner"

	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/render/todo"
	"rat/graph/util"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

const (
	// GraphTokenType graph tokens provide an overview of a nodes child nodes.
	// Graph tokens get substituted with a list tree of links to child nodes of
	// specified depth. Unlimited depth if omitted.
	GraphTokenType TokenType = "graph"
	// TodoTokenType todo searches for todos in child nodes and collects them
	// into a large singular todo. Token args can be used to specify search
	// options.
	TodoTokenType TokenType = "todo"
	// KanbanTokenType kanban tokens provide a kanban board of child nodes.
	KanbanTokenType TokenType = "kanban"
)

var (
	errUnknownTokenType = errors.New("unknown token type")
	tokenRegex          = regexp.MustCompile(`<rat(?:\s((?:.|\s)+?))\/>`)
)

// TokenType describes rat token types.
type TokenType string

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
//
//nolint:cyclop,gocyclo
func NewToken(raw string) (*Token, error) {
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

	tokenType, err := func(raw string) (TokenType, error) {
		target := TokenType(raw)

		for _, valid := range []TokenType{
			GraphTokenType,
			TodoTokenType,
			KanbanTokenType,
		} {
			if target == valid {
				return target, nil
			}
		}

		return "",
			errors.Wrapf(errUnknownTokenType, "unknown token type %s", target)
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

// Render produces the expanded version of the token.
func (t *Token) Render(
	part *jsonast.AstPart, n *graph.Node, p graph.Provider,
) error {
	switch t.Type {
	case TodoTokenType:
		return t.renderTodoToken(part, n, p)
	case GraphTokenType:
		return t.renderGraphToken(part, n, p)
	case KanbanTokenType:
		return nil
	default:
		return errors.Errorf("unknown token type - %s", t.Type)
	}
}

func (t *Token) renderGraphToken(
	part *jsonast.AstPart, n *graph.Node, p graph.Provider,
) error {
	depth, err := t.getArgDepth()
	if err != nil {
		return errors.Wrap(err, "failed to get depth")
	}

    graphPart := part

	var f func(graph node *graph.Node, part *jsonast.AstPart, d int) error
	f = func(node *graph.Node, part *jsonast.AstPart, d int) error {
		if depth != -1 && d >= depth {
			return nil
		}

		children, err := node.GetLeafs(p)
		if err != nil {
			return errors.Wrap(err, "failed to get leafs")
		}

		if len(children) == 0 {
			return nil
		}

		listPart := graphPart.AddContainer(
			&jsonast.AstPart{
				Type: "list",
			},
			true,
		)

		for _, child := range children {
			childPart := part.AddContainer(
				&jsonast.AstPart{
					Type: "list_item",
				},
				true,
			)

			childPart.AddLeaf(&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": child.Name,
				},
			})

			err := f(child, childPart, d+1)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = f(n, graphPart, 0)
	if err != nil {
		return errors.Wrap(err, "failed to walk graph")
	}

	return nil
}

func (t *Token) renderTodoToken(
	part *jsonast.AstPart,
	_ *graph.Node,
	p graph.Provider,
) error {
	filtersHas, err := t.getArgFilterHas()
	if err != nil {
		return errors.Wrap(err, "failed to get has hint filter")
	}

	filterValue, err := t.getArgFilterValue()
	if err != nil {
		return errors.Wrap(err, "failed to get value filter")
	}

	includeDone, includeDoneEntries, err := t.getArgInclude()
	if err != nil {
		return errors.Wrap(err, "failed to get include done arg")
	}

	sortRules, err := t.getArgSort()
	if err != nil {
		return errors.Wrap(err, "failed to get sort arg")
	}

	todos, err := t.getTodos(p)
	if err != nil {
		return errors.Wrap(err, "failed to get todos")
	}

	todos = util.Filter(
		todos,
		func(t *todo.Todo) bool {
			if !includeDoneEntries {
				t.RemoveDoneEntries()
			}

			if len(t.Entries) == 0 {
				return false
			}

			if !t.Done() {
				return true
			}

			return includeDone
		},
	)

	todos = todo.FilterHas(todos, filtersHas)
	todos = todo.FilterValue(todos, filterValue)
	sort.SliceStable(todo.NewSorter(sortRules)(todos))

	for _, t := range todos {
		t.JSON(part)
	}

	return nil
}

//nolint:gocyclo,cyclop
func (t *Token) getTodos(p graph.Provider) ([]*todo.Todo, error) {
	includeSources, excludeSources, err := t.getArgSources()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sources arg")
	}

	excludedNodes := make(map[uuid.UUID]bool)

	for _, exID := range excludeSources {
		exNode, err := p.GetByID(exID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get excluded node")
		}

		excludedNodes[exID] = true

		exNodes, err := exNode.ChildNodes(p)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get excluded nodes")
		}

		for _, exNode := range exNodes {
			excludedNodes[exNode.ID] = true
		}
	}

	var todos []*todo.Todo

	for _, inID := range includeSources {
		if excludedNodes[inID] {
			continue
		}

		inNode, err := p.GetByID(inID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get included node")
		}

		nodeTodos, err := todo.ParseNode(inNode)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse node todos")
		}

		todos = append(todos, nodeTodos...)

		err = inNode.Walk(
			p,
			func(_ int, node *graph.Node) (bool, error) {
				if excludedNodes[node.ID] {
					return false, nil
				}

				nodeTodos, err := todo.ParseNode(node)
				if err != nil {
					return false, errors.Wrap(err, "failed to parse node todos")
				}

				todos = append(todos, nodeTodos...)

				return true, nil
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to walk graph")
		}
	}

	return todos, nil
}

func (t *Token) transformKanbanToken(
	_ *graph.Node, _ graph.Provider,
) (string, error) {
	cols, err := t.getArgColumns()
	if err != nil {
		return "", errors.Wrap(err, "failed to get columns")
	}

	return fmt.Sprintf(
			`<div id="kanban">%s</div>`,
			strings.Join(
				util.Map(
					cols,
					func(id uuid.UUID) string {
						return id.String()
					},
				),
				",",
			),
		),
		nil
}

func (t *Token) getArgSources() ([]uuid.UUID, []uuid.UUID, error) {
	columnsArg, ok := t.Args["sources"]
	if !ok {
		return nil, nil, nil
	}

	parts := strings.Split(columnsArg, ",")

	var ( //nolint:prealloc
		include []uuid.UUID
		exclude []uuid.UUID
	)

	for _, raw := range parts {
		raw = strings.TrimSpace(raw)

		if strings.HasPrefix(raw, "-") {
			id, err := uuid.FromString(raw[1:])
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to parse id")
			}

			exclude = append(exclude, id)

			continue
		}

		id, err := uuid.FromString(raw)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse id")
		}

		include = append(include, id)
	}

	return include, exclude, nil
}

func (t *Token) getArgColumns() ([]uuid.UUID, error) {
	columnsArg, ok := t.Args["columns"]
	if !ok {
		return nil, nil
	}

	return parseListOfUUIDs(columnsArg)
}

func parseListOfUUIDs(raw string) ([]uuid.UUID, error) {
	parts := strings.Split(raw, ",")
	ids := make([]uuid.UUID, 0, len(parts))

	for _, raw := range parts {
		raw = strings.TrimSpace(raw)

		id, err := uuid.FromString(raw)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse id")
		}

		ids = append(ids, id)
	}

	return ids, nil
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

func (t *Token) getArgFilterHas() ([]*todo.FilterRule, error) {
	filterHasArg, ok := t.Args["filter_has"]
	if !ok {
		return nil, nil
	}

	rawFilters := strings.Split(filterHasArg, ",")
	filters := make([]*todo.FilterRule, 0, len(rawFilters))

	for _, rawFilter := range rawFilters {
		f, err := todo.ParseFilterRule(rawFilter)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse filter rule")
		}

		filters = append(filters, f)
	}

	return filters, nil
}

func (t *Token) getArgFilterValue() ([]*todo.FilterValueRule, error) {
	filterValueArg, ok := t.Args["filter_value"]
	if !ok {
		return nil, nil
	}

	rawFilters := strings.Split(strings.Trim(filterValueArg, "\""), ",")
	filters := make([]*todo.FilterValueRule, 0, len(rawFilters))

	for _, rawFilter := range rawFilters {
		f, err := todo.ParseFilterValueRule(rawFilter)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse filter rule")
		}

		filters = append(filters, f)
	}

	return filters, nil
}

func (t *Token) getArgInclude() (bool, bool, error) {
	includeArg, ok := t.Args["include"]
	if !ok {
		return false, false, nil
	}

	var done, doneEntries bool

	parts := strings.Split(includeArg, ",")

	for _, part := range parts {
		switch part {
		case "done":
			done = true
		case "done_entries":
			doneEntries = true
		default:
			return false, false, errors.Errorf("unknown include - %s", part)
		}
	}

	return done, doneEntries, nil
}

func (t *Token) getArgSort() ([]*todo.SortRule, error) {
	sortArg, ok := t.Args["sort"]
	if !ok {
		return nil, nil
	}

	rawRules := strings.Split(sortArg, ",")
	rules := make([]*todo.SortRule, 0, len(rawRules))

	for _, rawRule := range rawRules {
		r, err := todo.ParseSortRule(rawRule)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse sort rule")
		}

		rules = append(rules, r)
	}

	return rules, nil
}