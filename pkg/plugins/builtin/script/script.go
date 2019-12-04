package script

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gexec "os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the script plugin execute scripts
var (
	Plugin = taskplugin.New("script", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// Config is the configuration needed to execute a script
type Config struct {
	File    string   `json:"file,required"`
	Argv    []string `json:"argv,omitempty"`
	Timeout string   `json:"timeout_seconds,omitempty"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.File == "" {
		return errors.New("file is missing")
	}

	if cfg.Timeout != "" {
		if _, err := strconv.ParseUint(cfg.Timeout, 10, 64); err != nil {
			return fmt.Errorf("timeout is wrong %s", err.Error())
		}
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)

	var timeout time.Duration

	if cfg.Timeout != "" {
		t, _ := strconv.ParseInt(cfg.Timeout, 10, 64)
		timeout = time.Duration(t)
	} else {
		// default is 5*60 = 300 seconds or 5 minutes
		timeout = time.Duration(300)
	}

	ctxe, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	cmd := gexec.CommandContext(ctxe, cfg.File, cfg.Argv...)

	exitCode := 0

	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*gexec.ExitError); ok {
			exitCode = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			exitCode = 1
			if string(out) == "" {
				out = []byte(err.Error())
			}
		}
	} else {
		exitCode = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	signal := cmd.ProcessState.Sys().(syscall.WaitStatus).Signal().String()

	outStr := string(out)

	metadata := struct {
		ExitCode   string `json:"exit_code"`
		ExitSignal string `json:"exit_signal"`
		Output     string `json:"output"`
	}{
		ExitCode:   fmt.Sprint(exitCode),
		ExitSignal: signal,
		Output:     outStr,
	}

	lastNL := strings.LastIndexByte(outStr, '{')
	if lastNL == -1 {
		return nil, metadata, nil
	}

	lastLine := out[lastNL:]
	payload := make(map[string]interface{})
	err = json.Unmarshal([]byte(lastLine), &payload)
	if err != nil {
		return nil, nil, err
	}

	return payload, metadata, nil
}
