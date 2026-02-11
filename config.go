package main

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	TemplateFile string
	OutputFile string
	Prefix string
	Debug bool
	Version bool
}

func parseFlags() Config {
	var config Config
	flag.BoolVar(&config.Debug, "d", false, "enable debug mode")
	flag.StringVar(&config.Prefix, "p", "", "only consider variables starting with prefix")
	flag.StringVar(&config.TemplateFile, "t", "", "template file")
	flag.BoolVar(&config.Version, "v", false, "show version")
	flag.StringVar(&config.OutputFile, "o", "", "output file")
	flag.Parse()

	if config.Version {
		if BuildVersion == "" {
			BuildVersion = "development" // Fallback if not set during build
		}
		fmt.Printf("version %s\n", BuildVersion)
		os.Exit(0)
	}

	if len(config.TemplateFile) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	return config
}
