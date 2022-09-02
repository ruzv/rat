package args

import "github.com/spf13/pflag"

type Args struct {
	ConfigPath string
	LogPath    string
}

func Load() (*Args, bool) {
	configPath := pflag.StringP(
		"config", "c", "./config.json", "path to config file",
	)
	logPath := pflag.StringP(
		"log", "l", "./logs.log", "path to log file",
	)
	help := pflag.BoolP("help", "h", false, "show help")

	pflag.Parse()

	if *help {
		pflag.PrintDefaults()

		return nil, false
	}

	return &Args{
		ConfigPath: *configPath,
		LogPath:    *logPath,
	}, true
}
