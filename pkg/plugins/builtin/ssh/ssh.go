package pluginssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/juju/errors"
	"golang.org/x/crypto/ssh"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// connection configuration values
const (
	MaxHops                    = 10
	ConnTimeout                = 10 * time.Second
	OutputModeAutoResult       = "auto-result"
	OutputModeDisabled         = "disabled"
	OutputModeManualDelimiters = "manual-delimiters"
	OutputModeManualLastLine   = "manual-lastline"
)

// ssh plugin opens an ssh connection and runs commands on target machine
var (
	Plugin = taskplugin.New("ssh", "0.2", execssh,
		taskplugin.WithConfig(configssh, ConfigSSH{}),
	)
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
	AllowExitNonZero       bool              `json:"allow_exit_non_zero"`
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

	switch cfg.OutputMode {
	case "":
		// default will have to be reset in execssh as config modification will not be persisted
		cfg.OutputMode = OutputModeAutoResult
	case OutputModeAutoResult, OutputModeDisabled, OutputModeManualDelimiters, OutputModeManualLastLine:
	default:
		return fmt.Errorf("invalid value %q for output_mode, allowed values are: %s", cfg.OutputMode, strings.Join([]string{OutputModeAutoResult, OutputModeDisabled, OutputModeManualDelimiters, OutputModeManualLastLine}, ", "))
	}

	if cfg.OutputManualDelimiters != nil && cfg.OutputMode != OutputModeManualDelimiters {
		return fmt.Errorf("invalid parameter \"output_manual_delimiters\", output_mode is configured to %q", cfg.OutputMode)
	}

	if len(cfg.Result) > 0 && cfg.OutputMode != OutputModeAutoResult {
		return fmt.Errorf("invalid parameter \"result\", output_mode is configured to %q", cfg.OutputMode)
	}

	if cfg.OutputMode == OutputModeManualDelimiters && (cfg.OutputManualDelimiters == nil || len(cfg.OutputManualDelimiters) != 2) {
		length := 0
		if cfg.OutputManualDelimiters != nil {
			length = len(cfg.OutputManualDelimiters)
		}
		return fmt.Errorf("wrong number of output_manual_delimiters, 2 expected, found %d", length)
	}

	if cfg.OutputManualDelimiters != nil {
		if _, err := generateOutputDelimitersRegexp(cfg.OutputManualDelimiters[0], cfg.OutputManualDelimiters[1]); err != nil {
			return fmt.Errorf("unable to compile output_manual_delimiters regexp: %s", err)
		}
	}

	return nil
}

func generateOutputDelimitersRegexp(start, end string) (*regexp.Regexp, error) {
	return regexp.Compile("(?s)^.*" + start + "(.*)" + end + ".*$")
}

func execssh(stepName string, i interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := i.(*ConfigSSH)

	if cfg.OutputMode == "" {
		cfg.OutputMode = OutputModeAutoResult
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

	if cfg.OutputMode == OutputModeAutoResult {
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

	exitStatus := 0
	exitSignal := ""
	exitMessage := ""

	// Directly execute the command
	if len(cfg.Hops) == 0 {
		extraCmd = execStr
	}

	output, err := session.CombinedOutput(extraCmd)
	if err != nil {
		exitErr, ok := err.(*ssh.ExitError)
		if ok {
			exitStatus = exitErr.Waitmsg.ExitStatus()
			exitSignal = exitErr.Waitmsg.Signal()
			exitMessage = exitErr.Waitmsg.Msg()
		} else {
			return nil, nil, err
		}
	}

	outStr := string(output)
	metadata := map[string]interface{}{
		"output":      outStr,
		"exit_status": fmt.Sprint(exitStatus),
		"exit_signal": exitSignal,
		"exit_msg":    exitMessage,
	}

	payload := make(map[string]interface{})

	var resultLine string
	switch cfg.OutputMode {
	case OutputModeManualDelimiters:
		if rexp, err := generateOutputDelimitersRegexp(cfg.OutputManualDelimiters[0], cfg.OutputManualDelimiters[1]); err != nil {
			return nil, nil, fmt.Errorf("unable to compile output_manual_delimiters regexp: %s", err)
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

	if resultLine != "" {
		err = json.Unmarshal([]byte(resultLine), &payload)
		if err != nil {
			return nil, metadata, err
		}
	}

	if exitStatus != 0 && !cfg.AllowExitNonZero {
		return payload, metadata, fmt.Errorf("exit status code: %d", exitStatus)
	}

	return payload, metadata, nil
}

func retrieveLastLine(outStr string) (resultLine string, lastIndex int) {
	lastIndex = strings.LastIndexByte(outStr, '\n')
	if lastIndex != -1 {
		resultLine = strings.TrimSpace(outStr[lastIndex:])
	}
	return
}
