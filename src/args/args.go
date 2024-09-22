package args

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"rat/buildinfo"
)

// ErrExitZero indicates that no error occurred.
// The process should exit with 0.
var ErrExitZero = errors.New("exit 0")

// Args are the command line arguments.
type Args struct {
	ConfigPath string
}

// Load parses the command line arguments in to a Args struct.
func Load() (*Args, error) {
	configPath := pflag.StringP(
		"config", "c", "./config.yaml", "path to yaml or json config file",
	)

	help := pflag.BoolP("help", "h", false, "show help")
	version := pflag.BoolP("version", "v", false, "show version")

	pflag.Parse()

	if *help {
		pflag.PrintDefaults()

		return nil, ErrExitZero
	}

	if *version {
		_, err := fmt.Fprintf(
			os.Stdout,
			"rat server version %s\n",
			buildinfo.Version(),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to print version")
		}

		return nil, ErrExitZero
	}

	return &Args{ConfigPath: *configPath}, nil
}
