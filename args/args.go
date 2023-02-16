package args

import "github.com/spf13/pflag"

// Args are the command line arguments.
type Args struct {
	ConfigPath string
	Embed      bool
}

// Load parses the command line arguments in to a Args struct.
func Load() (*Args, bool) {
	configPath := pflag.StringP(
		"config", "c", "./config.json", "path to config file",
	)
	embed := pflag.BoolP(
		"embed", "e", true, "flag to toggle usage of embedded files",
	)

	help := pflag.BoolP("help", "h", false, "show help")

	pflag.Parse()

	if *help {
		pflag.PrintDefaults()

		return nil, false
	}

	return &Args{
		ConfigPath: *configPath,
		Embed:      *embed,
	}, true
}
