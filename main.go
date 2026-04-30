package main

import (
	"fmt"
	"os"

	"github.com/zooplus/tpl/tpl"
)

var BuildVersion = "development" // Fallback if not set during build

// control ALL exits from main
func main() {
	os.Exit(run())
}

func run() int {
	config := tpl.ParseFlags()
	logger := tpl.NewLogger(config.Debug)

	if config.Version {
		fmt.Printf("version %s\n", BuildVersion)
		return 0
	}

	if len(config.TemplateFile) == 0 {
		config.Usage()
		return 1
	}

	if _, err := os.Stat(config.TemplateFile); os.IsNotExist(err) {
		logger.Error("%s not found\n", config.TemplateFile)
		return 2
	}

	processor, err := tpl.NewTemplateProcessor(config, logger)
	if err != nil {
		logger.Error("error creating template processor: %v\n", err)
		return 2
	}

	defer func() {
		if err := processor.Close(); err != nil {
			logger.Error("error closing processor: %v\n", err)
		}
	}()

	err = processor.RenderTemplate()
	if err != nil {
		logger.Error("error rendering template %v: %v\n", config.TemplateFile, err)
		return 2
	}

	return 0
}
