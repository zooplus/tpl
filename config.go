package main

import (
	"flag"
)

type Config struct {
	TemplateFile string
	OutputFile   string
	Prefix       string
	Debug        bool
	Version      bool
}

func ParseFlags() Config {
	var config Config
	flag.BoolVar(&config.Debug, "d", false, "enable debug mode")
	flag.StringVar(&config.Prefix, "p", "", "only consider variables starting with prefix")
	flag.StringVar(&config.TemplateFile, "t", "", "template file")
	flag.BoolVar(&config.Version, "v", false, "show version")
	flag.StringVar(&config.OutputFile, "o", "", "output file")
	flag.Parse()
	return config
}

func (c Config) Usage() {
	flag.Usage()
}
