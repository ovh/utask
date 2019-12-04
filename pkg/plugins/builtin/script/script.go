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

	"github.com/ovh/utask"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the script plugin execute scripts
var (
	Plugin = taskplugin.New("script", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

type Metadata struct {
	ExitCode     string `json:"exit_code"`
	ProcessState string `json:"process_state"`
	Output       string `json:"output"`
}

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
		// default is 2*60 = 120 seconds or 2 minutes
		timeout = time.Duration(120)
	}

	timeout = timeout * time.Second

	ctxe, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := gexec.CommandContext(ctxe, utask.FScriptsFolder+cfg.File, cfg.Argv...)

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

	pState := cmd.ProcessState.String()

	outStr := string(out)

	metadata := Metadata{
		ExitCode:     fmt.Sprint(exitCode),
		ProcessState: pState,
		Output:       outStr,
	}

	lastNL := strings.LastIndexByte(outStr, '{')
	if lastNL == -1 {
		return nil, metadata, nil
	}

	lastLine := out[lastNL:]
	payload := make(map[string]interface{})
	err = json.Unmarshal([]byte(lastLine), &payload)
	if err != nil {
		return nil, metadata, nil
	}

	return payload, metadata, nil
}
