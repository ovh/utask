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

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the script plugin execute scripts
var (
	Plugin = taskplugin.New("script", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
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
	File    string   `json:"file"`
	Argv    []string `json:"argv,omitempty"`
	Timeout string   `json:"timeout,omitempty"`
	Stdin   string   `json:"stdin,omitempty"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.File == "" {
		return errors.New("file is missing")
	}

	scriptPath := filepath.Join(utask.FScriptsFolder, cfg.File)

	f, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s not found in FS: %s", cfg.File, err.Error())
	}

	if f.Mode()&0111 == 0 {
		if err := os.Chmod(scriptPath, f.Mode()|0111); err != nil {
			return fmt.Errorf("can't set exec permissions on %s: %s", cfg.File, err.Error())
		}
	}

	if cfg.Timeout != "" {
		if cfg.Timeout[0] == '-' {
			return errors.New("timeout must be positive")
		}
		if _, err := time.ParseDuration(cfg.Timeout); err != nil {
			return fmt.Errorf("timeout is wrong %s", err.Error())
		}
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
		timeout = time.Duration(2) * time.Minute
	}

	ctxe, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := gexec.CommandContext(ctxe, filepath.Join(utask.FScriptsFolder, cfg.File), cfg.Argv...)
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

	metadata := Metadata{
		ExitCode:      fmt.Sprint(exitCode),
		ProcessState:  pState,
		Output:        outStr,
		ExecutionTime: execTime.String(),
		Error:         metaError,
	}

	outputArray := strings.Split(outStr, "\n")
	lastLine := ""

	for i := len(outputArray) - 1; i >= 0; i-- {
		if len(outputArray[i]) > 0 {
			lastLine = outputArray[i]
			break
		}
	}

	if !(strings.Contains(lastLine, "{") && strings.Contains(lastLine, "}")) {
		return nil, metadata, nil
	}

	payload := make(map[string]interface{})
	err = json.Unmarshal([]byte(lastLine), &payload)
	if err != nil {
		return nil, metadata, err
	}

	return payload, metadata, nil
}
