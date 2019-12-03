package script

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	gexec "os/exec"
	"strconv"
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
	File    string   `json:"file"`
	Argv    []string `json:"argv"`
	Timeout string   `json:"timeout"`
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
		timeout = time.Duration(300)
	}

	ctxe, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	cmd := gexec.CommandContext(ctxe, cfg.File, cfg.Argv...)

	exitCode := 0

	var errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	stderr := errbuf.String()

	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*gexec.ExitError); ok {
			exitCode = exitError.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			exitCode = 1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		exitCode = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	signal := cmd.ProcessState.Sys().(syscall.WaitStatus).Signal().String()

	payload := struct {
		ExitCode   string      `json:"exit_code"`
		ExitSignal string      `json:"exit_signal"`
		Stdout     interface{} `json:"stdout"`
		Stderr     interface{} `json:"stderr"`
	}{
		ExitCode:   fmt.Sprint(exitCode),
		ExitSignal: signal,
		Stdout:     out,
		Stderr:     []byte(stderr),
	}

	return payload, cfg, nil
}
