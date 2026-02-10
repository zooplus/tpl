package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func run(t *testing.T, env map[string]string, args []string) (string, error) {
	cmd := shell.Command{
		Command: "go",
		Args:    args,
		Env:     env,
		Logger:  logger.Discard,
	}

	return shell.RunCommandAndGetOutputE(t, cmd)
}

func tpl(t *testing.T, env map[string]string, args ...string) (string, error) {
	defaultArgs := []string{"run", "."}
	return run(t, env, append(defaultArgs, args...))
}

func mergedEnv(overrides map[string]string) map[string]string {
	env := map[string]string{}
	for _, pair := range os.Environ() {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		env[key] = value
	}

	for key, value := range overrides {
		env[key] = value
	}

	return env
}

func baseEnv() map[string]string {
	return map[string]string{
		"foo":       "bar",
		"bar":       "[foo,bar]",
		"foobar":    "{foo:bar,bar:foo}",
		"foobaz":    "{foo:[bar,baz]}",
		"baz":       "1.0-123",
		"number":    "59614658972",
		"null":      "null",
		"empty":     "",
		"money":     "500\u20ac",
		"special":   "?&>=:/",
		"woot":      "[]",
		"whoa":      "{}",
		"backslash": "\\.\\/",
		"urls":      "{google:[https:://google.com,http:://google.de],github:https:://github.com}",
		"json":      `{"abc":123,"def":["a","b","c"],"ghi":"[{,!?!,}]"}`,
	}
}

func readFileAsString(t *testing.T, filePath string) string {
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return string(content)
}

func TestTplPrintVersion(t *testing.T) {
	t.Parallel()

	out, err := tpl(t, map[string]string{}, "-v")
	// should print version
	assert.Regexp(t, "^version (v?[0-9]{1,}\\.[0-9]{1,}\\.[0-9]{1,})|(development)", out)
	assert.Nil(t, err)
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestTplRendersExpectedOutput(t *testing.T) {
	env := mergedEnv(baseEnv())

	output, err := tpl(t, env, "-t", "test/test.tpl")
	require.NoError(t, err)

	expected := readFileAsString(t, "test/test.txt")
	// trimming because diff ignores that, but tests don't
	require.Equal(t, strings.TrimRight(expected, "\n"), strings.TrimRight(output, "\n"))
}

func TestTplPrefixFilter(t *testing.T) {
	t.Parallel()

	env := mergedEnv(map[string]string{
		"APP_VAR":     "app_value",
		"APP_VERSION": "1.0.0",
		"OTHER_VAR":   "other_value",
		"GLOBAL_VAR":  "global_value",
	})

	output, err := tpl(t, env, "-p", "APP_", "-t", "test/test_prefix.tpl")
	require.NoError(t, err)

	// Only APP_* variables should be defined, others should be empty
	assert.Contains(t, output, "APP_VAR: app_value")
	assert.Contains(t, output, "APP_VERSION: 1.0.0")
	assert.Contains(t, output, "OTHER_VAR: <no value>")
	assert.Contains(t, output, "GLOBAL_VAR: <no value>")
}

func TestTplOutputFile(t *testing.T) {
	t.Parallel()

	env := mergedEnv(baseEnv())
	outputFile := t.TempDir() + "/output.txt"

	_, err := tpl(t, env, "-t", "test/test.tpl", "-o", outputFile)
	require.NoError(t, err)

	// Verify output file was created and contains expected content
	content := readFileAsString(t, outputFile)
	expected := readFileAsString(t, "test/test.txt")

	require.Equal(t, strings.TrimRight(expected, "\n"), strings.TrimRight(content, "\n"))
}

func TestTplLargeEnvCounts(t *testing.T) {
	t.Parallel()
	expected := readFileAsString(t, "test/test.txt")
	envCounts := []int{700, 1000, 10000}

	for _, count := range envCounts {
		t.Run(fmt.Sprintf("count_%d", count), func(t *testing.T) {
			envOverrides := baseEnv()
			for i := 1; i <= count; i++ {
				envOverrides[fmt.Sprintf("ENV_VAR_%d", i)] = fmt.Sprintf("this is env var number %d", i)
			}

			env := mergedEnv(envOverrides)
			output, err := tpl(t, env, "-t", "test/test.tpl")
			require.NoError(t, err)
			require.Equal(t, strings.TrimRight(expected, "\n"), strings.TrimRight(output, "\n"))
		})
	}
}
