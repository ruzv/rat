package token

import (
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
	"rat/graph/render/todo"
	"rat/graph/util"
)

func (t *Token) renderTodo(
	part *jsonast.AstPart,
	n *graph.Node,
	p graph.Provider,
	r jsonast.Renderer,
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
		t.Render(part, n, r)
	}

	return nil
}

//nolint:gocyclo,cyclop //TODO: fix
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

		err = exNode.Walk(
			p,
			func(_ int, node *graph.Node) (bool, error) {
				excludedNodes[node.Header.ID] = true

				return true, nil
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to walk exclude nodes")
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

		err = inNode.Walk(
			p,
			func(_ int, node *graph.Node) (bool, error) {
				if excludedNodes[node.Header.ID] {
					return false, nil
				}

				nodeTodos, err := todo.ParseNode(node) //nolint:govet
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

func (t *Token) getArgSources() ([]uuid.UUID, []uuid.UUID, error) {
	columnsArg, ok := t.Args["sources"]
	if !ok {
		return nil, nil, errors.New(`missing "sources" argument`)
	}

	parts := strings.Split(columnsArg, ",")

	var ( //nolint:prealloc // unknown sizes.
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

	if len(include) == 0 {
		return nil, nil, errors.New(
			"sources arg must contain at least one include ID",
		)
	}

	return include, exclude, nil
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
