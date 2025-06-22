package token

import (
	"github.com/ruzv/rat/internal/buildinfo"
	"github.com/ruzv/rat/internal/graph/render/jsonast"
)

func renderVersion(part *jsonast.AstPart) {
	part.AddLeaf(
		&jsonast.AstPart{
			Type: "code",
			Attributes: jsonast.AstAttributes{
				"text": buildinfo.Version(),
			},
		},
	)
}
