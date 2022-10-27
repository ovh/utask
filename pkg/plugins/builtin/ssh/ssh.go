package pluginssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/ovh/utask/pkg/plugins/builtin/scriptutil"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// connection configuration values
const (
	MaxHops           = 10
	ConnTimeout       = 10 * time.Second
	DefaultCmdTimeout = 5 * time.Minute
)

// ssh plugin opens an ssh connection and runs commands on target machine
var (
	Plugin = taskplugin.New("ssh", "0.2", execssh,
		taskplugin.WithConfig(configssh, ConfigSSH{}),
		taskplugin.WithResources(resourcesssh),
	)
	ErrSessionTimeout = errors.New("ssh session has not terminated before timeout")
)

// ConfigSSH is the data needed to perform an SSH action
type ConfigSSH struct {
	User                   string            `json:"user"`
	Target                 string            `json:"target"`
	Hops                   []string          `json:"hops"`
	Script                 string            `json:"script"`
	OutputMode             string            `json:"output_mode"`
	Result                 map[string]string `json:"result"`
	OutputManualDelimiters []string          `json:"output_manual_delimiters"`
	Key                    string            `json:"ssh_key"`
	KeyPassphrase          string            `json:"ssh_key_passphrase"`
	ExitCodesUnrecoverable []string          `json:"exit_codes_unrecoverable"`
	Timeout                string            `json:"timeout,omitempty"`
}

func resourcesssh(i interface{}) []string {
	cfg := i.(*ConfigSSH)

	resources := []string{
		"socket",
		"url:" + cfg.Target,
	}
	for _, hop := range cfg.Hops {
		resources = append(resources, "url:"+hop)
	}

	return resources
}

func configssh(i interface{}) error {
	cfg := i.(*ConfigSSH)

	if cfg.User == "" {
		return errors.New("missing ssh username")
	}

	if cfg.Target == "" {
		return errors.New("missing ssh target")
	}

	if cfg.Key == "" {
		return errors.New("missing ssh key")
	}

	if len(cfg.Hops) > MaxHops {
		return fmt.Errorf("ssh too many hops (max %d)", MaxHops)
	}

	if cfg.Timeout != "" {
		dur, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return fmt.Errorf("can't parse timeout field %q: %s", cfg.Timeout, err.Error())
		}
		if dur < 0 {
			return errors.New("timeout must be positive")
		}
	}

	switch cfg.OutputMode {
	case "":
		// default will have to be reset in execssh as config modification will not be persisted
		cfg.OutputMode = scriptutil.OutputModeAutoResult
	case scriptutil.OutputModeAutoResult, scriptutil.OutputModeDisabled, scriptutil.OutputModeManualDelimiters, scriptutil.OutputModeManualLastLine:
	default:
		return fmt.Errorf("invalid value %q for output_mode, allowed values are: %s", cfg.OutputMode, strings.Join([]string{scriptutil.OutputModeAutoResult, scriptutil.OutputModeDisabled, scriptutil.OutputModeManualDelimiters, scriptutil.OutputModeManualLastLine}, ", "))
	}

	if cfg.OutputManualDelimiters != nil && cfg.OutputMode != scriptutil.OutputModeManualDelimiters {
		return fmt.Errorf("invalid parameter \"output_manual_delimiters\", output_mode is configured to %q", cfg.OutputMode)
	}

	if len(cfg.Result) > 0 && cfg.OutputMode != scriptutil.OutputModeAutoResult {
		return fmt.Errorf("invalid parameter \"result\", output_mode is configured to %q", cfg.OutputMode)
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

func execssh(stepName string, i interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := i.(*ConfigSSH)

	if cfg.OutputMode == "" {
		cfg.OutputMode = scriptutil.OutputModeAutoResult
	}

	var signer ssh.Signer
	var err error

	if cfg.KeyPassphrase == "" {
		signer, err = ssh.ParsePrivateKey([]byte(cfg.Key))
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.Key), []byte(cfg.KeyPassphrase))
	}
	if err != nil {
		return nil, nil, errors.NewBadRequest(err, "ssh plugin: private key")
	}

	var executionTimeout time.Duration

	if cfg.Timeout != "" {
		// Can skip error, value already validated.
		executionTimeout, _ = time.ParseDuration(cfg.Timeout)
	} else {
		executionTimeout = DefaultCmdTimeout
	}

	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         ConnTimeout,
	}

	var target string
	hops := []string{}
	if len(cfg.Hops) > 0 {
		// Start with first hop
		target = cfg.Hops[0]
		hops = append(hops, cfg.Hops[1:]...)
		hops = append(hops, cfg.Target)
	} else {
		target = cfg.Target
	}

	var firstErr error
	for {
		_, _, err := net.SplitHostPort(target)
		if err != nil {
			// port may be missing, append it and retry
			if firstErr != nil {
				return nil, nil, errors.NewBadRequest(firstErr, "ssh plugin: host port")
			}
			target = net.JoinHostPort(target, "22")
			firstErr = err
		} else {
			break
		}
	}

	client, err := ssh.Dial("tcp", target, config)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	execStr := cfg.Script

	// resulting JSON, able to compute commands like:
	// {
	//     "pwd": $(pwd)
	// }
	injectPL := `'{'`
	idx := 0
	for k, v := range cfg.Result {
		if idx > 0 {
			injectPL += `,`
		}
		injectPL += fmt.Sprintf(`'"%s":"'"%s"'"'`, strings.Replace(k, "\"", "", -1), strings.Replace(v, "\"", "", -1))
		idx++
	}
	injectPL += `'}'`

	if cfg.OutputMode == scriptutil.OutputModeAutoResult {
		execStr = fmt.Sprintf(`
function printResultJSON {
echo -n %s | sed --posix -z 's/\n/\\n/g'
}
trap printResultJSON EXIT
`, injectPL) + execStr
	}

	in := bytes.NewBuffer([]byte(execStr))
	session.Stdin = in

	extraCmd := ""
	for i, hop := range hops {
		if i > 0 {
			extraCmd += " -- "
		}
		extraCmd += hop
	}

	exitCode := 0
	exitSignal := ""
	exitMessage := ""

	// Directly execute the command
	if len(cfg.Hops) == 0 {
		extraCmd = execStr
	}

	exit := make(chan struct{}, 1)
	timer := time.NewTimer(executionTimeout)

	go func() {
		select {
		case <-timer.C:
			err := session.Signal(ssh.SIGKILL)
			if err != nil {
				logrus.Warnf("session signal error: %s", err)
			}
		case <-exit:
		}
	}()
	cmdOutput, cmdErr := session.CombinedOutput(extraCmd)
	if !timer.Stop() {
		logrus.Debugf("session run error: %s", cmdErr)
		cmdErr = ErrSessionTimeout
	}
	close(exit)

	if cmdErr != nil {
		exitErr, ok := cmdErr.(*ssh.ExitError)
		if ok {
			exitCode = exitErr.Waitmsg.ExitStatus()
			exitSignal = exitErr.Waitmsg.Signal()
			exitMessage = exitErr.Waitmsg.Msg()
		} else {
			return nil, nil, cmdErr
		}
	}
	outStr := string(cmdOutput)

	metadata := map[string]interface{}{
		"output":      outStr,
		"exit_code":   strconv.Itoa(exitCode),
		"exit_signal": exitSignal,
		"exit_msg":    exitMessage,
	}
	output := make(map[string]interface{})

	if resultLine, err := scriptutil.ParseOutput(outStr, cfg.OutputMode, cfg.OutputManualDelimiters); err != nil {
		return nil, metadata, err
	} else if resultLine != "" {
		err = json.Unmarshal([]byte(resultLine), &output)
		if err != nil && exitCode == 0 {
			return nil, metadata, err
		}
	}
	if exitCode != 0 {
		return output, metadata, scriptutil.FormatErrorExitCode(exitCode, cfg.ExitCodesUnrecoverable, cmdErr)
	}
	return output, metadata, nil
}
