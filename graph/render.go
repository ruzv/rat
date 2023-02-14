package graph

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"private/rat/graph/render/todo"
	"private/rat/graph/util"

	"github.com/gofrs/uuid"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/pkg/errors"
)

// Render parses nodes content, converts tokens into markdown and renders it to
// HTML.
func (n *Node) Render(p Provider, rend *html.Renderer) string {
	return string(
		markdown.ToHTML(
			[]byte(n.convertContentTokens(p)),
			parser.NewWithExtensions(parser.CommonExtensions),
			rend,
		),
	)
}

var ratTokenRegex = regexp.MustCompile(
	`<rat(\s([^>]+))>`,
)

// converts special tokens in content like
// <rat graph> <rat link> <rat todo>
// into markdown
func (n *Node) convertContentTokens(p Provider) string {
	matches := ratTokenRegex.FindAllIndex([]byte(n.Content), -1)
	parsedSource := n.Content

	for _, match := range util.ReverseSlice(matches) {
		tag := n.Content[match[0]:match[1]]

		parsed, err := n.parseRatToken(p, tag)
		if err != nil {
			log.Warning("failed to parse rat token", tag, err)
		}

		left := parsedSource[:match[0]]
		right := parsedSource[match[1]:]

		parsedSource = left + parsed + right
	}

	return parsedSource
}

//nolint:grouper
const (
	ratTagKeywordLink  = "link"
	ratTagKeywordGraph = "graph"
	ratTagKeywordTodo  = "todo"
	ratTagKeywordImg   = "img"
)

// parses a single rat tag
// <rat keyword arg1 arg2> .
func (n *Node) parseRatToken(p Provider, token string) (string, error) {
	// rat keyword arg1 arg2
	pTag := strings.Trim(token, "<>")

	// keyword arg1 arg2
	args := strings.Fields(pTag)[1:]

	// keyword
	keyword := args[0]

	// arg1 arg2
	args = args[1:]

	switch keyword {
	case ratTagKeywordLink:
		if len(args) == 2 {
			return parseRatTagLink(p, args[0], args[1])
		}

		if len(args) == 1 {
			return parseRatTagLink(p, args[0], "")
		}

		return "", errors.New("too many arguments")

	case ratTagKeywordGraph:
		depth := -1 // no depth limit

		if len(args) == 1 {
			var err error

			depth, err = strconv.Atoi(args[0])
			if err != nil {
				return "", errors.Wrap(err, "failed to parse depth")
			}
		}

		return n.parseRatTagGraph(p, depth)
	case ratTagKeywordTodo:
		return n.parseRatTagTodo(p, args)

	default:
		return "", errors.New("unknown keyword")
	}
}

func parseRatTagLink(p Provider, linkID, name string) (string, error) {
	id, err := uuid.FromString(linkID)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse link id")
	}

	node, err := p.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to get node by ID")
	}

	if name == "" {
		name = node.Name
	}

	return util.Link(node.Path, name), nil
}

func (n *Node) parseRatTagGraph(p Provider, depth int) (string, error) {
	var links []string

	err := n.Walk(
		p,
		func(d int, node *Node) bool {
			if d == depth {
				return false
			}

			links = append(
				links,
				fmt.Sprintf(
					"%s- %s",
					strings.Repeat("\t", depth),
					util.Link(node.Path, node.Name),
				),
			)

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	return strings.Join(links, "\n"), nil
}

// -------------------------------------------------------------------------- //
// todo
// -------------------------------------------------------------------------- //

var todoRegex = regexp.MustCompile(
	"```todo\n((?:.*\n)*?)```",
)

//nolint:gocognit,gocyclo,cyclop
func (n *Node) parseRatTagTodo(p Provider, args []string) (string, error) {
	var (
		err            error
		parent         *Node
		depth          = math.MaxInt
		filterPriority bool
		argIdx         int
	)

	for argIdx < len(args) {
		err := func() error {
			// consumes one
			defer func() { argIdx++ }()

			arg := args[argIdx]

			if arg == "filter" {
				filterPriority = true

				return nil
			}

			// can consume two
			if argIdx+1 >= len(args) {
				return nil
			}

			// consumes two
			defer func() { argIdx++ }()

			if arg == "parent" {
				parentID, err := uuid.FromString(args[argIdx+1])
				if err != nil {
					return errors.Wrap(err, "failed to parse parent id")
				}

				parent, err = p.GetByID(parentID)
				if err != nil {
					return errors.Wrap(err, "failed to get parent node")
				}

				return nil
			}

			if arg == "depth" {
				depth, err = strconv.Atoi(args[argIdx+1])
				if err != nil {
					return errors.Wrap(err, "failed to parse depth")
				}

				return nil
			}

			return nil
		}()
		if err != nil {
			return "", err
		}
	}

	if parent == nil {
		parent, err = n.Parent(p)
		if err != nil {
			return "", errors.Wrap(err, "failed to get parent")
		}
	}

	var (
		todoLists []*todo.TodoList
		derr      error
	)

	err = parent.Walk(
		p,
		func(i int, leaf *Node) bool {
			if i == depth {
				return false
			}

			matches := todoRegex.FindAllStringSubmatch(leaf.Content, -1)
			for _, match := range matches {
				todoL, err := todo.Parse(match[1])
				if err != nil {
					log.Warning("failed to parse todo", err)
					derr = errors.Wrap(err, "failed to parse todo")

					return false
				}

				notDone := todoL.NotDone()
				if notDone.Empty() {
					continue
				}

				if filterPriority && !notDone.HasPriority() {
					continue
				}

				notDone.SourceNodePath = leaf.Path
				todoLists = append(todoLists, notDone)
			}

			return true
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to walk graph")
	}

	if derr != nil {
		return derr.Error(), derr
	}

	sort.SliceStable(
		todoLists,
		func(iIdx, jIdx int) bool {
			i := todoLists[iIdx]
			j := todoLists[jIdx]

			if !i.HasPriority() && !j.HasPriority() {
				return false
			}

			if i.HasPriority() && !j.HasPriority() {
				return true
			}

			if !i.HasPriority() && j.HasPriority() {
				return false
			}

			return i.Priority() > j.Priority()
		},
	)

	rawTodoLists := make([]string, 0, len(todoLists))

	for _, todoList := range todoLists {
		rawTodoLists = append(
			rawTodoLists,
			todoList.Markdown(),
		)
	}

	return strings.Join(rawTodoLists, "\n"), nil
}
