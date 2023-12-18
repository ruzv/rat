package token

import (
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render/jsonast"
)

func (t *Token) renderKanban(
	part *jsonast.AstPart,
	p graph.Provider,
	r jsonast.Renderer,
) error {
	cols, err := t.getArgColumns()
	if err != nil {
		return errors.Wrap(err, "failed to get columns")
	}

	kanbanPart := part.AddContainer(
		&jsonast.AstPart{
			Type: "kanban",
		},
		true,
	)

	for _, col := range cols {
		n, err := p.GetByID(col)
		if err != nil {
			return errors.Wrap(err, "failed to get node")
		}

		children, err := n.GetLeafs(p)
		if err != nil {
			return errors.Wrap(err, "failed to get child nodes")
		}

		colPart := kanbanPart.AddContainer(
			&jsonast.AstPart{
				Type: "kanban_column",
				Attributes: jsonast.AstAttributes{
					"id":   n.Header.ID.String(),
					"name": n.Name,
					"path": n.Path.String(),
				},
			},
			true,
		)

		for _, child := range children {
			cardPart := colPart.AddContainer(
				&jsonast.AstPart{
					Type: "kanban_card",
					Attributes: jsonast.AstAttributes{
						"id":   child.Header.ID.String(),
						"name": child.Name,
						"path": child.Path.String(),
					},
				},
				true,
			)

			err := r.Render(cardPart, child)
			if err != nil {
				return errors.Wrap(err, "failed to render child node")
			}
		}
	}

	return nil
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
