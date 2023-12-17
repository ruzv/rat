package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"rat/config"
	"rat/graph"
	"rat/graph/index"
	"rat/graph/pathcache"
	"rat/graph/singlefile"
	"rat/graph/sync"
	"rat/logr"
)

// GraphServices contains service components of a graph.
type GraphServices struct {
	Graph           graph.Provider
	Syncer          *sync.Syncer
	Index           *index.GraphIndex
	FileURLResolver *FileURLResolver
	log             *logr.LogR
}

// FileURLResolver resolves relative file URLs in node content to absolute urls
// to password protected, pre-configured fileservers.
type FileURLResolver struct {
	configs []*config.FileserverConfig
	log     *logr.LogR
}

// Resolve iterates configured fileservers until a match is found and a server
// returns a 200 OK for the specified path. Returns the absolute URL to the
// file.
func (r *FileURLResolver) Resolve(path string) (string, error) {
	dest, err := url.Parse(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse file path as url")
	}

	if dest.IsAbs() {
		return path, nil
	}

	for _, c := range r.configs {
		destURL, err := r.resolve(c, path)
		if err != nil {
			r.log.Debugf("failed to resolve file url %q: %s", path, err.Error())

			continue
		}

		return destURL, nil
	}

	return "", errors.Errorf("failed to resolve file url %q", path)
}

func (r *FileURLResolver) resolve(
	c *config.FileserverConfig, path string,
) (string, error) {
	dest, err := url.Parse(c.Authority)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse fileserver authority")
	}

	dest.Path = path
	dest.User = url.UserPassword(c.User, c.Password)

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

// NewGraphServices creates a new graph services.
func NewGraphServices(
	log *logr.LogR, graphConf *config.GraphConfig,
) (*GraphServices, error) {
	log = log.Prefix("services")

	gs := &GraphServices{
		FileURLResolver: &FileURLResolver{
			configs: graphConf.Fileservers,
			log:     log.Prefix("file-url-resolver"),
		},
		log: log,
	}

	sf, err := singlefile.NewSingleFile(graphConf.Name, graphConf.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create single file graph")
	}

	gs.Graph = pathcache.NewPathCache(sf)

	if graphConf.Sync != nil {
		log.Infof("starting syncer")

		gs.Syncer, err = sync.NewSyncer(log, graphConf.Path, graphConf.Sync)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create syncer")
		}

		gs.Syncer.Start()
	}

	gs.Index, err = index.NewIndex(log, gs.Graph)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	logMetrics(log.Prefix("metrics"), gs.Graph)

	return gs, nil
}

// Close closes the graph services.
func (gs *GraphServices) Close() error {
	if gs.Syncer != nil {
		gs.log.Infof("stopping syncer")

		gs.Syncer.Stop()
	}

	err := gs.Index.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close index")
	}

	return nil
}

func logMetrics(log *logr.LogR, p graph.Provider) {
	r, err := p.Root()
	if err != nil {
		log.Errorf("failed to log metrics: %s", err.Error())

		return
	}

	m, err := r.Metrics(p)
	if err != nil {
		log.Errorf("failed to log metrics: %s", err.Error())

		return
	}

	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		log.Errorf("failed to log metrics: %s", err.Error())

		return
	}

	log.Infof(string(b))
}
