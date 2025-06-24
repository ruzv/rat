package token

import (
	"strconv"

	"github.com/gomarkdown/markdown/ast"
	"github.com/pkg/errors"
	"github.com/ruzv/rat/internal/graph/render/jsonast"
	"github.com/ruzv/rat/internal/graph/render/mdast"
)

type tocEntry struct {
	name       string
	id         string
	level      int
	subEntries []*tocEntry
	prev       *tocEntry
}

//nolint:cyclop
func (t *Token) renderToC(
	root *jsonast.AstPart,
	rootMarkdownNode ast.Node,
) error {
	depth, err := t.getArgDepth()
	if err != nil {
		return errors.Wrap(err, "failed to get depth")
	}

	includeTitle, err := t.getArgIncludeTitle()
	if err != nil {
		return errors.Wrap(err, "failed to get include_title")
	}

	rootToC := &tocEntry{
		name:       "first", // this gets ignored, not rendered
		id:         "",
		level:      0,
		subEntries: []*tocEntry{},
		prev:       nil,
	}
	rootToC.prev = rootToC

	toc := rootToC

	ast.WalkFunc(
		rootMarkdownNode,
		func(node ast.Node, entering bool) ast.WalkStatus {
			if !entering {
				return ast.GoToNext
			}

			heading, ok := node.(*ast.Heading)
			if !ok {
				return ast.GoToNext
			}

			switch {
			case heading.Level == toc.level:
				// new same level entry
				toc = toc.prev
			case heading.Level < toc.level:
				// new higher level entry
				for heading.Level <= toc.level {
					toc = toc.prev
				}
			default: // heading.Level > toc.level new lower level entry toc=toc
			}

			// add new entry
			sub := &tocEntry{
				name:       mdast.GetTextOfSubNodes(node),
				id:         string(heading.Attribute.ID),
				level:      heading.Level,
				subEntries: []*tocEntry{},
				prev:       toc,
			}

			toc.subEntries = append(toc.subEntries, sub)

			toc = sub

			return ast.GoToNext
		},
	)

	// single title (h1 in doc)
	if !includeTitle && len(rootToC.subEntries) == 1 {
		renderToCEntry(root, rootToC.subEntries[0], depth, 0)

		return nil
	}

	renderToCEntry(root, rootToC, depth, 0)

	return nil
}

func renderToCEntry(
	part *jsonast.AstPart,
	entry *tocEntry,
	depth, d int,
) {
	if depth != -1 && d >= depth {
		return
	}

	listPart := part.AddContainer(
		&jsonast.AstPart{
			Type: "list",
			Attributes: jsonast.AstAttributes{
				"type": "unordered",
			},
		},
		true,
	)

	for _, subEntry := range entry.subEntries {
		linkPart := listPart.AddContainer(
			&jsonast.AstPart{
				Type: "list_item",
				Attributes: jsonast.AstAttributes{
					"type": "unordered",
				},
			},
			true,
		).AddContainer(
			&jsonast.AstPart{
				Type: "link",
				Attributes: jsonast.AstAttributes{
					"destination": "#" + subEntry.id,
				},
			},
			true,
		)

		linkPart.AddLeaf(
			&jsonast.AstPart{
				Type: "text",
				Attributes: jsonast.AstAttributes{
					"text": subEntry.name,
				},
			},
		)

		renderToCEntry(listPart, subEntry, depth, d+1)
	}
}

func (t *Token) getArgIncludeTitle() (bool, error) {
	includeTitleArg, ok := t.Args["include_title"]
	if !ok {
		return false, nil
	}

	includeTitle, err := strconv.ParseBool(includeTitleArg)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"failed to parse %q as exclude_title",
			includeTitleArg,
		)
	}

	return includeTitle, nil
}
