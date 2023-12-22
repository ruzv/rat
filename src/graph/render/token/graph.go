package token

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
)

func (t *Token) renderGraph(
	part *jsonast.AstPart, n *graph.Node, p graph.Provider,
) error {
	depth, err := t.getArgDepth()
	if err != nil {
		return errors.Wrap(err, "failed to get depth")
	}

	err = renderGraphTokenWithDepth(part, n, p, depth, 0)
	if err != nil {
		return errors.Wrap(err, "failed to walk graph")
	}

	return nil
}

func renderGraphTokenWithDepth(
	part *jsonast.AstPart,
	n *graph.Node,
	p graph.Provider,
	depth, d int,
) error {
	if depth != -1 && d >= depth {
		return nil
	}

	children, err := n.GetLeafs(p)
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	if len(children) == 0 {
		return nil
	}

	listPart := part.AddContainer(
		&jsonast.AstPart{
			Type: "list",
		},
		true,
	)

	for _, child := range children {
		listPart.AddContainer(
			&jsonast.AstPart{
				Type: "list_item",
			},
			true,
		).AddLeaf(&jsonast.AstPart{
			Type: "graph_link",
			Attributes: jsonast.AstAttributes{
				"title":       child.Name(),
				"destination": fmt.Sprintf("/view/%s", child.Path),
			},
		})

		err := renderGraphTokenWithDepth(listPart, child, p, depth, d+1)
		if err != nil {
			return err
		}
	}

	return nil
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
