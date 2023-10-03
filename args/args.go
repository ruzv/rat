package args

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// Args are the command line arguments.
type Args struct {
	ConfigPath string
}

// Load parses the command line arguments in to a Args struct.
func Load(ratVersion string) (*Args, bool) {
	configPath := pflag.StringP(
		"config", "c", "./config.yaml", "path to yaml or json config file",
	)

	help := pflag.BoolP("help", "h", false, "show help")
	version := pflag.BoolP("version", "v", false, "show version")

	pflag.Parse()

	if *help {
		pflag.PrintDefaults()

		return nil, false
	}

	if *version {
		fmt.Fprintf(os.Stdout, "rat version %s\n", ratVersion)

		return nil, false
	}

	return &Args{
		ConfigPath: *configPath,
	}, true
}
