package main

import (
	"os"
)

var BuildVersion string


func main() {
	config := parseFlags()
	logger := NewLogger(config.Debug)
	processor := NewTemplateProcessor(config, logger)

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
