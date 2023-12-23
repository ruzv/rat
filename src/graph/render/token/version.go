package token

import (
	"rat/buildinfo"
	"rat/graph/render/jsonast"
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
