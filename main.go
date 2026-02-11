package main

import (
	"fmt"
	"os"
)

var BuildVersion string

// control ALL exits from main
func main() {
	config := ParseFlags()
	logger := NewLogger(config.Debug)

	if config.Version {
		if BuildVersion == "" {
			BuildVersion = "development" // Fallback if not set during build
		}
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

	processor, err := NewTemplateProcessor(config, logger)
	if err != nil {
		logger.Error("error creating template processor: %v\n", err)
		os.Exit(2)
	}

	err = processor.RenderTemplate()
	if err != nil {
		logger.Error("error rendering template %v: %v\n", config.TemplateFile, err)
		os.Exit(2)
	}
}
