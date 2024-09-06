package connector

import (
	"errors"
	"fmt"
	"io"
	"sync"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes"
)

var ErrUserNotFound = errors.New("user not found on context")

// NOTICE: Ensures the Sessioner interface is implemented.
var _ modes.Sessioner = (*Sessioner)(nil)

// Sessioner implements the Sessioner interface when the server is running in connector mode.
type Sessioner struct {
	// container is the device name.
	//
	// NOTICE: It's a pointer because when the server is created, we don't know the device name yet, that is set later.
	container *string
	docker    dockerclient.APIClient
}

// NewSessioner creates a new instance of Sessioner for the connector mode.
// The container is a pointer to a string because when the server is created, we don't know the device name yet, that
// is set later.
func NewSessioner(container *string, docker dockerclient.APIClient) *Sessioner {
	return &Sessioner{
		container: container,
		docker:    docker,
	}
}

// Shell handles the server's SSH shell session when server is running in connector mode.
func (s *Sessioner) Shell(session gliderssh.Session) error {
	sspty, _, _ := session.Pty()

	// NOTICE(r): To identify what the container the connector should connect to, we use the `deviceName` as the container name
	container := *s.container

	user, ok := session.Context().Value("user").(*osauth.User)
	if !ok {
		return ErrUserNotFound
	}

	resp, id, err := attachShellToContainer(session.Context(), s.docker, container, user, [2]uint{uint(sspty.Window.Height), uint(sspty.Window.Width)})
	if err != nil {
		return err
	}
	defer resp.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			code, err := exitCodeExecFromContainer(s.docker, id)
			if err != nil {
				fmt.Println(err)
			}

			session.Exit(code) //nolint:errcheck
		}()

		if _, err := io.Copy(session, resp.Conn); err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.Close()

		if _, err := io.Copy(resp.Conn, session); err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}()

	wg.Wait()

	return nil
}

// Exec handles the SSH's server exec session when server is running in connector mode.
func (s *Sessioner) Exec(session gliderssh.Session) error {
	sspty, _, isPty := session.Pty()

	// NOTICE(r): To identify what the container the connector should connect to, we use the `deviceName` as the container name
	container := *s.container

	user, ok := session.Context().Value("user").(*osauth.User)
	if !ok {
		return ErrUserNotFound
	}

	resp, id, err := attachExecToContainer(session.Context(), s.docker, container, user, isPty, session.Command(), [2]uint{uint(sspty.Window.Height), uint(sspty.Window.Width)})
	if err != nil {
		return err
	}
	defer resp.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			code, err := exitCodeExecFromContainer(s.docker, id)
			if err != nil {
				fmt.Println(err)
			}

			session.Exit(code) //nolint:errcheck
		}()

		// NOTICE: According to the [Docker] documentation, we can "demultiplex" a command sent to container, but only
		// when the exec started doesn't allocate a TTY. As a result, we check if the exec's is requesting it and do
		// what was recommended by [Docker]'s to get the stdout and stderr separately.
		//
		// [Docker]: https://pkg.go.dev/github.com/docker/docker/client#Client.ContainerAttach
		if isPty {
			if _, err := io.Copy(session, resp.Reader); err != nil && err != io.EOF {
				fmt.Println(err)
			}
		} else {
			if _, err := stdcopy.StdCopy(session, session.Stderr(), resp.Reader); err != nil && err != io.EOF {
				fmt.Println(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(resp.Conn, session); err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}()

	wg.Wait()

	return nil
}

// Heredoc handles the server's SSH heredoc session when server is running in connector mode.
//
// heredoc is special block of code that contains multi-line strings that will be redirected to a stdin of a shell. It
// request a shell, but doesn't allocate a pty.
func (s *Sessioner) Heredoc(session gliderssh.Session) error {
	sspty, _, _ := session.Pty()

	// NOTICE(r): To identify what the container the connector should connect to, we use the `deviceName` as the container name
	container := *s.container

	user, ok := session.Context().Value("user").(*osauth.User)
	if !ok {
		return ErrUserNotFound
	}

	resp, id, err := attachHereDocToContainer(session.Context(), s.docker, container, user, [2]uint{uint(sspty.Window.Height), uint(sspty.Window.Width)})
	if err != nil {
		return err
	}
	defer resp.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			code, err := exitCodeExecFromContainer(s.docker, id)
			if err != nil {
				fmt.Println(err)
			}

			session.Exit(code) //nolint:errcheck
		}()

		if _, err := stdcopy.StdCopy(session, session.Stderr(), resp.Reader); err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(resp.Conn, session); err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}()

	wg.Wait()

	return nil
}

// SFTP handles the SSH's server sftp session when server is running in connector mode.
//
// sftp is a subsystem of SSH that allows file operations over SSH.
func (s *Sessioner) SFTP(_ gliderssh.Session) error {
	return errors.New("SFTP isn't supported to ShellHub Agent in connector mode")
}
