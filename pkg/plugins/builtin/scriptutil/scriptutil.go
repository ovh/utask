package scriptutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errors"
)

const (
	OutputModeAutoResult       = "auto-result"
	OutputModeDisabled         = "disabled"
	OutputModeManualDelimiters = "manual-delimiters"
	OutputModeManualLastLine   = "manual-lastline"
)

var (
	exitCodesUnrecoverableRegex = regexp.MustCompile(`^(\d+)(?:-(\d+))?$`)
)

func ValidateExitCodesUnreachable(exitCodes []string) error {
	for _, value := range exitCodes {
		matches := exitCodesUnrecoverableRegex.FindStringSubmatch(value)
		if len(matches) == 0 {
			return fmt.Errorf("invalid value %q for exit_codes_unrecoverable, should be an integer, or a range of integer (e.g: 123, or 120-130)", value)
		}
		exitCodeStartStr, exitCodeEndStr := matches[1], matches[2]
		var exitCodeStart, exitCodeEnd int64
		if exitCodeEndStr != "" {
			var err error
			exitCodeStart, err = strconv.ParseInt(exitCodeStartStr, 10, 64)
			if err != nil {
				return err
			}
			exitCodeEnd, err = strconv.ParseInt(exitCodeEndStr, 10, 64)
			if err != nil {
				return err
			}

			if exitCodeEnd <= exitCodeStart {
				return fmt.Errorf("exit_codes_unrecoverable value %q should have end exit_code superior to start exit_code", value)
			}

			if exitCodeStart == 0 {
				return fmt.Errorf("exit_codes_unrecoverable cannot map exit code 0 (non-sense)")
			}
		}
	}
	return nil
}

func FormatErrorExitCode(exitCode int, exitCodes []string, err error) error {
	pluginError := fmt.Errorf("exit code: %d", exitCode)

	for _, value := range exitCodes {
		matches := exitCodesUnrecoverableRegex.FindStringSubmatch(value)
		if len(matches) == 0 {
			return fmt.Errorf("exit_codes_unrecoverable value doesnt match regex, fatal error")
		}

		exitCodeStartStr, exitCodeEndStr := matches[1], matches[2]
		if exitCodeEndStr == "" && exitCodeStartStr != fmt.Sprint(exitCode) {
			continue
		}

		if exitCodeStartStr == fmt.Sprint(exitCode) {
			pluginError = errors.NewBadRequest(err, fmt.Sprintf("Client error: exit code: %d", exitCode))
			break
		}

		var exitCodeStart, exitCodeEnd int64
		exitCodeStart, err = strconv.ParseInt(exitCodeStartStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid starting exit_code value %q in exit_codes_unrecoverable: %s", exitCodeStartStr, err)
		}
		exitCodeEnd, err = strconv.ParseInt(exitCodeEndStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid ending exit_code value %q in exit_codes_unrecoverable: %s", exitCodeEndStr, err)
		}

		if exitCodeStart <= int64(exitCode) && int64(exitCode) <= exitCodeEnd {
			pluginError = errors.NewBadRequest(err, fmt.Sprintf("Client error: exit code: %d", exitCode))
			break
		}
	}
	return pluginError
}

func GenerateOutputDelimitersRegexp(start, end string) (*regexp.Regexp, error) {
	return regexp.Compile("(?s)^.*" + start + "(.*)" + end + ".*$")
}

func ParseOutput(outStr, outputMode string, outputManualDelimiters []string) (resultLine string, err error) {
	switch outputMode {
	case OutputModeManualDelimiters:
		if rexp, err := GenerateOutputDelimitersRegexp(outputManualDelimiters[0], outputManualDelimiters[1]); err != nil {
			return "", fmt.Errorf("unable to compile output_manual_delimiters regexp: %s", err)
		} else if matches := rexp.FindStringSubmatch(outStr); len(matches) > 0 {
			resultLine = matches[1]
		}

	case OutputModeAutoResult, OutputModeManualLastLine:
		var lastIndex int
		resultLine, lastIndex = retrieveLastLine(outStr)
		if resultLine == "" && lastIndex != -1 {
			// a lot of programs are returning a new line at the end of output, we need to strip it if exists
			resultLine, lastIndex = retrieveLastLine(outStr[0:lastIndex])
		}
	}

	return
}

func retrieveLastLine(outStr string) (resultLine string, lastIndex int) {
	resultLine = outStr
	lastIndex = strings.LastIndexByte(outStr, '\n')
	if lastIndex != -1 {
		resultLine = strings.TrimSpace(outStr[lastIndex:])
	}
	return
}
