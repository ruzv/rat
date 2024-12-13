package urlresolve

import (
	"context"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"rat/logr"
)

// Config contains url resolver configuration parameters.
type Config struct {
	Fileservers []*FileserverConfig `yaml:"fileservers"`
}

// FileserverConfig defines configuration parameters for a fileserver that
// web app can use to retrieve files.
type FileserverConfig struct {
	Authority string `yaml:"authority" validate:"nonzero"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
}

// Resolver resolves relative file URLs in node content to absolute urls
// to password protected, pre-configured fileservers.
type Resolver struct {
	log          *logr.LogR
	fileservers  []*FileserverConfig
	apiAuthority string
}

// NewResolver creates a new relative url resolver, that tries to match a
// relative URL to a fileserver the resolver is configured to, returning
// absolute URL to the resource on a fileserver.
func NewResolver(
	c *Config, log *logr.LogR, apiAuthority string,
) (*Resolver, error) {
	var fileservers []*FileserverConfig

	if c != nil {
		fileservers = c.Fileservers
	}

	// validate here, at startup to avoid unexpected runtime errors when
	// resolving urls.
	_, err := url.Parse(apiAuthority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pares API authority")
	}

	return &Resolver{
		log:          log.Prefix("url-resolver"),
		fileservers:  fileservers,
		apiAuthority: apiAuthority,
	}, nil
}

// Resolve iterates configured fileservers until a match is found and a server
// returns a 200 OK for the specified path. Returns the absolute URL to the
// file.
func (r *Resolver) Resolve(path string) (string, error) {
	dest, err := url.Parse(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse file path as url")
	}

	if dest.IsAbs() {
		return path, nil
	}

	if len(r.fileservers) == 0 {
		return "", errors.Errorf("no fileservers configured")
	}

	for _, fs := range r.fileservers {
		destURL, err := r.resolve(fs, path)
		if err != nil {
			r.log.Debugf("failed to resolve file url %q: %s", path, err.Error())

			continue
		}

		return destURL, nil
	}

	return "", errors.Errorf("failed to resolve file url %q", path)
}

// PrefixResolverEndpoint adds the endpoint resolver API route prefix to the
// path if it is not an absolute URL.
func (r *Resolver) PrefixResolverEndpoint(path string) string {
	parsed, err := url.Parse(path)
	if err != nil {
		r.log.Errorf("failed to parse URL: %s", err.Error())

		return path
	}

	if parsed.IsAbs() {
		return path
	}

	res, err := url.JoinPath("/graph/file/", path)
	if err != nil {
		return path
	}

	u, err := url.Parse(r.apiAuthority)
	if err != nil {
		r.log.Errorf("failed to parse API authority: %s", err.Error())

		return path
	}

	u.Path = res

	return u.String()
}

func (r *Resolver) resolve(
	c *FileserverConfig, path string,
) (string, error) {
	dest, err := url.Parse(c.Authority)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse fileserver authority")
	}

	dest.Path = path

	if c.User != "" && c.Password != "" {
		dest.User = url.UserPassword(c.User, c.Password)
	}

	redactedDestURL := dest.Redacted()
	destURL := dest.String()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodHead,
		destURL,
		http.NoBody,
	)
	if err != nil {
		return "", errors.Wrapf(
			err, "failed to create head request to %q", redactedDestURL,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(
			err, "head request to %q failed", redactedDestURL,
		)
	}

	defer resp.Body.Close() //nolint:errcheck // ignore.

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf(
			"head request to %q returned status code %d",
			redactedDestURL,
			resp.StatusCode,
		)
	}

	r.log.Debugf(
		"head request to %q returned Content-Type %s",
		redactedDestURL,
		resp.Header.Get("Content-Type"),
	)

	return destURL, nil
}
