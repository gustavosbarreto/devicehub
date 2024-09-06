//go:build freebsd

package host

import (
	"os"
	"os/exec"

	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes/host/command"
)

func newShellCmd(deviceName string, username string, term string, envs []string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	user, err := osauth.LookupUser(username)
	if err != nil {
		return nil
	}

	if shell == "" {
		shell = user.Shell
	}

	if term == "" {
		term = "xterm"
	}

	cmd := command.NewCmd(user, shell, term, deviceName, envs, shell, "-")

	return cmd
}
