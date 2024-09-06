package ssh

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes/host/command"
)

type pipe struct {
	in  *os.File
	out *os.File
	err *os.File
}

func (p *pipe) Read(data []byte) (int, error) {
	return p.in.Read(data)
}

func (p *pipe) Write(data []byte) (int, error) {
	return p.out.Write(data)
}

func (p *pipe) Close() error {
	os.Exit(0)

	return nil
}

// NewSFTPServer creates a new SFTP server when a new session is created between the agent and the server.
func NewSFTPServer(mode command.SFTPServerMode) {
	piped := &pipe{os.Stdin, os.Stdout, os.Stderr}

	if mode == command.SFTPServerModeDocker {
		if err := syscall.Chroot("/host"); err != nil {
			fmt.Fprintln(os.Stderr, err)

			return
		}
	}

	home, ok := os.LookupEnv("HOME")
	if !ok {
		fmt.Fprintln(os.Stderr, errors.New("HOME environment variable not set"))

		return
	}

	toInt := func(s string, ok bool) (int, error) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}

		return i, nil
	}

	gid, err := toInt(os.LookupEnv("GID"))
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.New("GID environment variable not set"))

		return
	}

	uid, err := toInt(os.LookupEnv("UID"))
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.New("UID environment variable not set"))

		return
	}

	if err := syscall.Chdir(home); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	if err := syscall.Setgid(gid); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	if err := syscall.Setuid(uid); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	server, err := sftp.NewServer(piped, []sftp.ServerOption{}...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return
	}

	if err := server.Serve(); err != io.EOF {
		fmt.Fprintln(os.Stderr, err)
	}

	server.Close()
}
