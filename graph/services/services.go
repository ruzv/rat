package services

import (
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

type FileURLResolver struct {
	configs []*config.FileserverConfig
	log     *logr.LogR
}

func (r *FileURLResolver) Resolve(path string) (string, error) {
	dest, err := url.Parse(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse file path as url")
	}

	if dest.IsAbs() {
		return path, nil
	}

	for _, c := range r.configs {
		dest.Host = c.Host
		dest.Scheme = c.Scheme

		destURL := dest.String()

		resp, err := http.Head(destURL)
		if err != nil {
			r.log.Debugf("head request to %q failed: %s", destURL, err.Error())

			continue
		}

		if resp.StatusCode != http.StatusOK {
			r.log.Debugf(
				"head request to %q returned status code %d",
				destURL,
				resp.StatusCode,
			)

			continue
		}

		return destURL, nil
	}

	return "", errors.Errorf("failed to resolve file url %q", path)
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
