package main

import (
	"fmt"
	"os"
	"github.com/tgagor/tpl/tpl"
)

var BuildVersion = "development" // Fallback if not set during build

// control ALL exits from main
func main() {
	config := tpl.ParseFlags()
	logger := tpl.NewLogger(config.Debug)

	if config.Version {
		fmt.Printf("version %s\n", BuildVersion)
		os.Exit(0)
	}

	if len(config.TemplateFile) == 0 {
		config.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(config.TemplateFile); os.IsNotExist(err) {
		logger.Error("%s not found\n", config.TemplateFile)
		os.Exit(2)
	}

	processor, err := tpl.NewTemplateProcessor(config, logger)
	if err != nil {
		logger.Error("error creating template processor: %v\n", err)
		os.Exit(2)
	}

	defer processor.Close() // ensure resources are cleaned up
	err = processor.RenderTemplate()
	if err != nil {
		logger.Error("error rendering template %v: %v\n", config.TemplateFile, err)
		os.Exit(2)
	}
}
