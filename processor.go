package main

import (
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/Masterminds/sprig/v3"
)

type QuotingRegexes struct {
	afterComma  *regexp.Regexp
	beforeComma *regexp.Regexp
	afterBrace  *regexp.Regexp
	beforeBrace *regexp.Regexp
	afterColon  *regexp.Regexp
	beforeColon *regexp.Regexp
	doubleColon *regexp.Regexp
}

type TemplateProcessor struct {
	config         Config
	logger         Logger
	environment    map[string]any
	writer         io.Writer
	quotingRegexes QuotingRegexes
}

func NewTemplateProcessor(config Config, logger Logger) (*TemplateProcessor, error) {
	tp := &TemplateProcessor{
		config:      config,
		logger:      logger,
		writer:      os.Stdout,
		environment: make(map[string]any),
		// compile once and reuse later
		quotingRegexes: QuotingRegexes{
			afterComma:  regexp.MustCompile(`,([^[{"])`),
			beforeComma: regexp.MustCompile(`([^]}"]),`),
			afterBrace:  regexp.MustCompile(`([\[{])([^][}{,"])`),
			beforeBrace: regexp.MustCompile(`([^][}{,"])([\]}])`),
			afterColon:  regexp.MustCompile(`([^:]):([^:[{"])`),
			beforeColon: regexp.MustCompile(`([^:"]):([^:])`),
			doubleColon: regexp.MustCompile(`::`),
		},
	}

	tp.buildEnvironment()
	logger.Debug("environment map is: %v\n", tp.environment)

	err := tp.setWriter()
	if err != nil {
		return nil, fmt.Errorf("error setting writer: %v", err)
	}

	return tp, nil
}

func (tp *TemplateProcessor) parseTemplate(filePath string) (*template.Template, error) {
	return template.New(path.Base(filePath)).
		Funcs(sprig.TxtFuncMap()).
		Funcs(tp.funcMap()).
		ParseFiles(filePath)
}

func (tp *TemplateProcessor) renderInclude(fileName string, safeMode bool) (string, error) {
	// lookup relative file names in same directory like main template
	lookupDir := ""
	if !strings.HasPrefix(fileName, "/") {
		lookupDir = path.Dir(tp.config.TemplateFile)
	}
	fullPath := path.Join(lookupDir, fileName)

	// ignore non-existing files
	if safeMode {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return "", nil
		}
	}

	tpl, err := tp.parseTemplate(fullPath)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	if err := tpl.Execute(&result, tp.environment); err != nil {
		return "", err
	}
	return result.String(), nil
}

func looksLikeJSON(inputStr string) bool {
	trimmed := strings.TrimSpace(inputStr)
	if trimmed == "" {
		return false
	}

	switch trimmed[0] {
	case '{', '[', '"':
		return true
	default:
		return false
	}
}

func (tp *TemplateProcessor) parseInput(inputStr string) (result interface{}, err error) {
	tp.logger.Debug("----\ninput is: %v\n", inputStr)

	if !looksLikeJSON(inputStr) {
		tp.logger.Debug("result is: %v\n----\n", inputStr)
		return inputStr, nil
	}

	// try to parse a plain json first
	jsonStr := inputStr
	err = json.Unmarshal([]byte(jsonStr), &result)

	// now try to enrich unquoted json
	if err != nil {
		// insert " after , if next is none of [ { "
		jsonStr = tp.quotingRegexes.afterComma.ReplaceAllString(jsonStr, ",\"$1")
		// insert " before , if previous is none of ] } "
		jsonStr = tp.quotingRegexes.beforeComma.ReplaceAllString(jsonStr, "$1\",")
		// insert " after [ { if next is none of ] [ } { , "
		jsonStr = tp.quotingRegexes.afterBrace.ReplaceAllString(jsonStr, "$1\"$2")
		// insert " before ] } if previous is none of ] [ } { , "
		jsonStr = tp.quotingRegexes.beforeBrace.ReplaceAllString(jsonStr, "$1\"$2")
		// insert " after : if next is none of : [ { "
		jsonStr = tp.quotingRegexes.afterColon.ReplaceAllString(jsonStr, "$1:\"$2")
		// insert " before : if previous is not :
		jsonStr = tp.quotingRegexes.beforeColon.ReplaceAllString(jsonStr, "$1\":$2")
		// replace :: with : (double colons can be used to escape a colon)
		jsonStr = tp.quotingRegexes.doubleColon.ReplaceAllString(jsonStr, ":")
	}
	tp.logger.Debug("json is: %v\n", jsonStr)

	// try parsing json again, if it fails fall back to the plain input value
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil || result == nil || reflect.TypeOf(result).Kind() == reflect.Float64 {
		result = inputStr
	}

	if err != nil {
		tp.logger.Debug("result is: %v, error: %v\n----\n", result, err)
	} else {
		tp.logger.Debug("result is: %v\n----\n", result)
	}

	return result, err
}

// generate environment map
func (tp *TemplateProcessor) buildEnvironment() {
	for _, envVar := range os.Environ() {
		key, value, ok := strings.Cut(envVar, "=")
		if !ok {
			continue
		}

		if !strings.HasPrefix(key, tp.config.Prefix) {
			continue
		}

		data, err := tp.parseInput(value)
		if err != nil {
			tp.environment[key] = value
		} else {
			tp.environment[key] = data
		}
	}
}

func (tp *TemplateProcessor) setWriter() error {
	if len(tp.config.OutputFile) > 0 {
		// Create file and truncate it if it already exists
		out, err := os.OpenFile(tp.config.OutputFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening output file: %s", err)
		}
		tp.writer = out
	}
	return nil
}

func (tp *TemplateProcessor) RenderTemplate() error {
	tpl, err := tp.parseTemplate(tp.config.TemplateFile)
	if err != nil {
		return err
	}
	return tpl.Execute(tp.writer, tp.environment)
}
