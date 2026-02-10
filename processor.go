package main

import (
	"regexp"
	"strings"
	"path"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"bytes"
)

type TemplateProcessor struct {
	config Config
	logger Logger
	environment map[string]any
	quotingRegexes struct {
		afterComma  *regexp.Regexp
		beforeComma *regexp.Regexp
		afterBrace  *regexp.Regexp
		beforeBrace *regexp.Regexp
		afterColon  *regexp.Regexp
		beforeColon *regexp.Regexp
		doubleColon *regexp.Regexp
	}
}

func (tp *TemplateProcessor) NewTemplateProcessor(config Config, logger Logger, environment map[string]any) *TemplateProcessor {
	return &TemplateProcessor{
		config: config,
		logger: logger,
		environment: environment,
		quotingRegexes: struct {
			afterComma  *regexp.Regexp
			beforeComma *regexp.Regexp
			afterBrace  *regexp.Regexp
			beforeBrace *regexp.Regexp
			afterColon  *regexp.Regexp
			beforeColon *regexp.Regexp
			doubleColon *regexp.Regexp
		}{
			afterComma:  regexp.MustCompile(`,([^[{"])`),
			beforeComma: regexp.MustCompile(`([^]}"]),`),
			afterBrace:  regexp.MustCompile(`([\[{])([^][}{,"])`),
			beforeBrace: regexp.MustCompile(`([^][}{,"])([\]}])`),
			afterColon:  regexp.MustCompile(`([^:]):([^:[{"])`),
			beforeColon: regexp.MustCompile(`([^:"]):([^:])`),
			doubleColon: regexp.MustCompile(`::`),
		},
	}
}

func (tp *TemplateProcessor) renderInclude(fileName string, safeMode bool) (string, error) {
	// lookup relative file names in same directory like main template
	lookupDir := ""
	if !strings.HasPrefix(fileName, "/") {
		lookupDir = path.Dir(tp.config.TemplateFile)
	}

	// ignore non-existing files
	if safeMode {
		if _, err := os.Stat(path.Join(lookupDir, fileName)); os.IsNotExist(err) {
			return "", nil
		}
	}

	tpl := template.Must(template.New(path.Base(fileName)).Funcs(sprig.TxtFuncMap()).ParseFiles(path.Join(lookupDir, fileName)))

	var result bytes.Buffer
	err := tpl.Execute(&result, tp.environment)
	return result.String(), err
}
