package token

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
)

func (t *Token) renderGraphic(
	root *jsonast.AstPart,
	n *graph.Node,
	p graph.Provider,
) error {
	settings := map[string]string{}

	engine, ok := t.Args["engine"]
	if !ok {
		engine = "circo"
		settings["mindist"] = "0.01"
		settings["bgcolor"] = "transparent"
	}

	if !slices.Contains(
		[]string{
			"circo",
			"dot",
			"fdp",
			"neato",
			"osage",
			"patchwork",
			"twopi",
		},
		engine,
	) {
		return errors.Errorf("unknown graphviz engine - %q", engine)
	}

	for k, v := range t.Args {
		setting, ok := strings.CutPrefix(k, "setting_")
		if !ok {
			continue
		}

		settings[setting] = v
	}

	var (
		err      error
		sourceID = n.Header.ID
	)

	source, ok := t.Args["source"]
	if ok {
		sourceID, err = uuid.FromString(source)
		if err != nil {
			return errors.Wrapf(err, "failed to parse source id %q", source)
		}
	}

	depth, err := t.getArgDepth()
	if err != nil {
		return errors.Wrap(err, "failed to get depth")
	}

	sourceNode, err := p.GetByID(sourceID)
	if err != nil {
		return errors.Wrap(err, "failed to get source node")
	}

	buff := bytes.Buffer{}

	buff.WriteString("graph {\n")

	for k, v := range settings {
		buff.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	fmt.Fprintf(
		&buff,
		"%q [label=\"%s\",shape=circle,width=%f,color=\"#f3715d\",fillcolor=\"#f3715d\",style=filled,fixedsize=true]\n",
		sourceNode.Header.ID.String(),
		sourceNode.Name(),
		3.0,
	)

	err = renderGraphicWithDepth(
		&buff,
		sourceNode,
		p,
		depth,
		0,
	)
	if err != nil {
		return errors.Wrap(err, "failed to render graphic")
	}

	buff.WriteString("}\n")

	root.AddLeaf(
		&jsonast.AstPart{
			Type: "graphviz",
			Attributes: jsonast.AstAttributes{
				"text":   buff.String(),
				"engine": engine,
			},
		},
	)

	return nil
}

func renderGraphicWithDepth(
	buff *bytes.Buffer,
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

	for _, child := range children {
		fmt.Fprintf(
			buff,
			"%q [label=\"%s\",shape=circle,width=%f,color=\"#f3715d\",fillcolor=\"#f3715d\",style=filled,fixedsize=true]\n",
			child.Header.ID.String(),
			child.Name(),
			2.0/float64(d+1),
		)

		fmt.Fprintf(
			buff,
			"%q -- %q [color=\"#f3715d\"]\n",
			n.Header.ID.String(),
			child.Header.ID.String(),
		)

		err := renderGraphicWithDepth(buff, child, p, depth, d+1)
		if err != nil {
			return err
		}
	}

	return nil
}
