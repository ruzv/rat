package token

import (
	"github.com/pkg/errors"
	"rat/graph/render/jsonast"
	"rat/graph/services/urlresolve"
)

func (t *Token) renderEmbed(part *jsonast.AstPart) error {
	embedURL, ok := t.Args["url"]
	if !ok {
		return errors.Wrap(
			errMissingArgument, "missing url arg for embed token",
		)
	}

	part.AddLeaf(
		&jsonast.AstPart{
			Type: "embed",
			Attributes: jsonast.AstAttributes{
				"url": urlresolve.PrefixResolverEndpoint(
					embedURL,
				),
			},
		},
	)

	return nil
}
