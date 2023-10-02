package args

import (
	"github.com/spf13/pflag"
)

// Args are the command line arguments.
type Args struct {
	ConfigPath string
}

// Load parses the command line arguments in to a Args struct.
func Load() (*Args, bool) {
	configPath := pflag.StringP(
		"config", "c", "./config.yaml", "path to yaml or json config file",
	)

	help := pflag.BoolP("help", "h", false, "show help")

	pflag.Parse()

	if *help {
		pflag.PrintDefaults()

		return nil, false
	}

	return &Args{
		ConfigPath: *configPath,
	}, true
}
