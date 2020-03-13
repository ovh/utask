package script

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	gexec "os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ovh/utask"

	"github.com/ovh/utask/pkg/plugins/builtin/scriptutil"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the script plugin execute scripts
var (
	Plugin = taskplugin.New("script", "0.2", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

const (
	exitCodeMetadataKey      string = "exit_code"
	processStateMetadataKey  string = "process_state"
	outputMetadataKey        string = "output"
	executionTimeMetadataKey string = "execution_time"
	errorMetadataKey         string = "error"
)

// Metadata represents the metadata of script execution
type Metadata struct {
	ExitCode      string `json:"exit_code"`
	ProcessState  string `json:"process_state"`
	Output        string `json:"output"`
	ExecutionTime string `json:"execution_time"`
	Error         string `json:"error"`
}

// Config is the configuration needed to execute a script
type Config struct {
	File                   string   `json:"file_path"`
	Argv                   []string `json:"argv,omitempty"`
	Timeout                string   `json:"timeout,omitempty"`
	Stdin                  string   `json:"stdin,omitempty"`
	OutputMode             string   `json:"output_mode"`
	OutputManualDelimiters []string `json:"output_manual_delimiters"`
	ExitCodesUnrecoverable []string `json:"exit_codes_unrecoverable"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.File == "" {
		return errors.New("file is missing")
	}

	scriptPath := filepath.Join(utask.FScriptsFolder, cfg.File)

	f, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("can't stat %q: %s", scriptPath, err.Error())
	}

	if f.Mode()&0111 == 0 {
		return fmt.Errorf("%q is not executable", scriptPath)
	}

	if cfg.Timeout != "" {
		if cfg.Timeout[0] == '-' {
			return errors.New("timeout must be positive")
		}
		if _, err := time.ParseDuration(cfg.Timeout); err != nil {
			return fmt.Errorf("can't parse timeout field %q: %s", cfg.Timeout, err.Error())
		}
	}

	switch cfg.OutputMode {
	case "":
		// default will have to be reset in exec as config modification will not be persisted
		cfg.OutputMode = scriptutil.OutputModeManualLastLine
	case scriptutil.OutputModeDisabled, scriptutil.OutputModeManualDelimiters, scriptutil.OutputModeManualLastLine:
	default:
		return fmt.Errorf("invalid value %q for output_mode, allowed values are: %s", cfg.OutputMode, strings.Join([]string{scriptutil.OutputModeDisabled, scriptutil.OutputModeManualDelimiters, scriptutil.OutputModeManualLastLine}, ", "))
	}

	if cfg.OutputManualDelimiters != nil && cfg.OutputMode != scriptutil.OutputModeManualDelimiters {
		return fmt.Errorf("invalid parameter \"output_manual_delimiters\", output_mode is configured to %q", cfg.OutputMode)
	}

	if cfg.OutputMode == scriptutil.OutputModeManualDelimiters && (cfg.OutputManualDelimiters == nil || len(cfg.OutputManualDelimiters) != 2) {
		length := 0
		if cfg.OutputManualDelimiters != nil {
			length = len(cfg.OutputManualDelimiters)
		}
		return fmt.Errorf("wrong number of output_manual_delimiters, 2 expected, found %d", length)
	}

	if cfg.OutputManualDelimiters != nil {
		if _, err := scriptutil.GenerateOutputDelimitersRegexp(cfg.OutputManualDelimiters[0], cfg.OutputManualDelimiters[1]); err != nil {
			return fmt.Errorf("unable to compile output_manual_delimiters regexp: %s", err)
		}
	}

	if err := scriptutil.ValidateExitCodesUnreachable(cfg.ExitCodesUnrecoverable); err != nil {
		return err
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)

	var timeout time.Duration

	if cfg.Timeout != "" {
		timeout, _ = time.ParseDuration(cfg.Timeout)
	} else {
		// default is 2 * 1 minute = 2 minutes
		timeout = 2 * time.Minute
	}

	if cfg.OutputMode == "" {
		cfg.OutputMode = scriptutil.OutputModeManualLastLine
	}

	ctxe, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := gexec.CommandContext(ctxe, fmt.Sprintf("./%s", cfg.File), cfg.Argv...)
	cmd.Dir = utask.FScriptsFolder
	cmd.Stdin = strings.NewReader(cfg.Stdin)

	exitCode := 0
	metaError := ""

	// start exec time timer
	timer := time.Now()
	// execute script
	out, err := cmd.CombinedOutput()
	// evaluate exec time
	execTime := time.Since(timer)

	if err != nil {
		if exitError, ok := err.(*gexec.ExitError); ok {
			exitCode = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			exitCode = 1
		}
		metaError = err.Error()
	} else {
		exitCode = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	pState := cmd.ProcessState.String()

	outStr := string(out)

	metadata := map[string]interface{}{
		exitCodeMetadataKey:      fmt.Sprint(exitCode),
		processStateMetadataKey:  pState,
		outputMetadataKey:        outStr,
		executionTimeMetadataKey: execTime.String(),
		errorMetadataKey:         metaError,
	}

	payload := make(map[string]interface{})

	if resultLine, err := scriptutil.ParseOutput(outStr, cfg.OutputMode, cfg.OutputManualDelimiters); err != nil {
		return nil, metadata, err
	} else if resultLine != "" {
		err = json.Unmarshal([]byte(resultLine), &payload)
		if err != nil && exitCode == 0 {
			return nil, metadata, err
		}
	}

	if exitCode != 0 {
		return payload, metadata, scriptutil.FormatErrorExitCode(exitCode, cfg.ExitCodesUnrecoverable, err)
	}

	return payload, metadata, nil
}
