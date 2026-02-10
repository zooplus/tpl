package main

import (
	"flag"
	"fmt"
	"os"
)

var BuildVersion string


func main() {
	config := parseFlags()
	logger := NewLogger(config.Debug)
	processor := NewTemplateProcessor(config, logger)

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

	if _, err := os.Stat(config.TemplateFile); os.IsNotExist(err) {
		logger.Fatal("%s not found\n", config.TemplateFile)
	}

	processor.buildEnvironment()
	logger.Debug("environment map is: %v\n", processor.environment)

	err := processor.setWriter()
	if err != nil {
		logger.Fatal("error setting writer: %v\n", err)
	}

	err = processor.renderTemplate()
	if err != nil {
		logger.Fatal("error rendering template %v: %v\n", config.TemplateFile, err)
	}
}
