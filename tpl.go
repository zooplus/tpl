package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var BuildVersion string
var environment = make(map[string]any)
var templateFile string // temporary global, until we refactor to methods

// add custom functions
var customFuctions = template.FuncMap{
	"include":     include,
	"mustInclude": mustInclude,
}

var (
	reInsertQuoteAfterComma  = regexp.MustCompile(`,([^[{"])`)
	reInsertQuoteBeforeComma = regexp.MustCompile(`([^]}"]),`)
	reInsertQuoteAfterBrace  = regexp.MustCompile(`([\[{])([^][}{,"])`)
	reInsertQuoteBeforeBrace = regexp.MustCompile(`([^][}{,"])([\]}])`)
	reInsertQuoteAfterColon  = regexp.MustCompile(`([^:]):([^:[{"])`)
	reInsertQuoteBeforeColon = regexp.MustCompile(`([^:"]):([^:])`)
	reReplaceDoubleColon     = regexp.MustCompile(`::`)
)



func inputToObject(inputStr string, debug bool) (result interface{}, err error) {
	if debug {
		fmt.Fprintf(os.Stderr, "----\ninput is: %v\n", inputStr)
	}

	if !looksLikeJSON(inputStr) {
		if debug {
			fmt.Fprintf(os.Stderr, "result is: %v\n----\n", inputStr)
		}
		return inputStr, nil
	}

	// try to parse a plain json first
	jsonStr := inputStr
	err = json.Unmarshal([]byte(jsonStr), &result)

	// now try to enrich unquoted json
	if err != nil {
		// insert " after , if next is none of [ { "
		jsonStr = reInsertQuoteAfterComma.ReplaceAllString(jsonStr, ",\"$1")
		// insert " before , if previous is none of ] } "
		jsonStr = reInsertQuoteBeforeComma.ReplaceAllString(jsonStr, "$1\",")
		// insert " after [ { if next is none of ] [ } { , "
		jsonStr = reInsertQuoteAfterBrace.ReplaceAllString(jsonStr, "$1\"$2")
		// insert " before ] } if previous is none of ] [ } { , "
		jsonStr = reInsertQuoteBeforeBrace.ReplaceAllString(jsonStr, "$1\"$2")
		// insert " after : if next is none of : [ { "
		jsonStr = reInsertQuoteAfterColon.ReplaceAllString(jsonStr, "$1:\"$2")
		// insert " before : if previous is not :
		jsonStr = reInsertQuoteBeforeColon.ReplaceAllString(jsonStr, "$1\":$2")
		// replace :: with : (double colons can be used to escape a colon)
		jsonStr = reReplaceDoubleColon.ReplaceAllString(jsonStr, ":")
	}
	if debug {
		fmt.Fprintf(os.Stderr, "json is: %v\n", jsonStr)
	}

	// try parsing json again, if it fails fall back to the plain input value
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil || result == nil || reflect.TypeOf(result).Kind() == reflect.Float64 {
		result = inputStr
	}

	if debug {
		if err != nil {
			fmt.Fprintf(os.Stderr, "result is: %v, error: %v\n----\n", result, err)
		} else {
			fmt.Fprintf(os.Stderr, "result is: %v\n----\n", result)
		}
	}

	return result, err
}

func renderInclude(templateFile string, fileName string, safeMode bool) string {
	// lookup relative file names in same directory like main template
	lookupDir := ""
	if !strings.HasPrefix(fileName, "/") {
		lookupDir = path.Dir(templateFile)
	}

	// ignore non-existing files
	if safeMode {
		if _, err := os.Stat(path.Join(lookupDir, fileName)); os.IsNotExist(err) {
			return ""
		}
	}

	tpl := template.Must(template.New(path.Base(fileName)).Funcs(sprig.TxtFuncMap()).ParseFiles(path.Join(lookupDir, fileName)))

	var result bytes.Buffer
	tpl.Execute(&result, environment)
	return result.String()
}

func include(fileName string) string {
	return renderInclude(templateFile, fileName, true)
}

func mustInclude(fileName string) string {
	return renderInclude(templateFile, fileName, false)
}


func main() {
	config := parseFlags()
	templateFile = config.TemplateFile
	logger := NewLogger(config.Debug)

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
		logger.Error("%s not found\n", config.TemplateFile)
	}

	// generate environment map
	for _, envVar := range os.Environ() {
		envKey, envValue, ok := strings.Cut(envVar, "=")
		if !ok {
			continue
		}

		if !strings.HasPrefix(envKey, config.Prefix) {
			continue
		}

		data, err := inputToObject(envValue, config.Debug)
		if err != nil {
			environment[envKey] = envValue
		} else {
			environment[envKey] = data
		}
	}

	if config.Debug {
		logger.Debug("environment map is: %v\n", environment)
	}

	outputWriter := os.Stdout
	if len(config.OutputFile) > 0 {
		// Create file and truncate it if it already exists
		out, err := os.OpenFile(config.OutputFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error("Error opening output file: %s\n", err)
			return
		}
		outputWriter = out
	}

	// render template
	tpl := template.Must(template.New(path.Base(config.TemplateFile)).Funcs(sprig.TxtFuncMap()).Funcs(customFuctions).ParseFiles(config.TemplateFile))
	err := tpl.Execute(outputWriter, environment)
	if err != nil {
		logger.Error("error rendering template %v: %v\n", config.TemplateFile, err)
	}
}
