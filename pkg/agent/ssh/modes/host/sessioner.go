package host

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes/host/command"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/utmp"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// NOTICE: Ensures the Sessioner interface is implemented.
var _ modes.Sessioner = (*Sessioner)(nil)

// Sessioner implements the Sessioner interface when the server is running in host mode.
type Sessioner struct {
	mu   sync.Mutex
	cmds map[string]*exec.Cmd
	// deviceName is the device name.
	//
	// NOTICE: It's a pointer because when the server is created, we don't know the device name yet, that is set later.
	deviceName *string
}

func (s *Sessioner) SetCmds(cmds map[string]*exec.Cmd) {
	s.cmds = cmds
}

// NewSessioner creates a new instance of Sessioner for the host mode.
// The device name is a pointer to a string because when the server is created, we don't know the device name yet, that
// is set later.
func NewSessioner(deviceName *string, cmds map[string]*exec.Cmd) *Sessioner {
	return &Sessioner{
		deviceName: deviceName,
		cmds:       cmds,
	}
}

// Shell manages the SSH shell session of the server when operating in host mode.
func (s *Sessioner) Shell(session gliderssh.Session) error {
	sspty, winCh, isPty := session.Pty()

	scmd := newShellCmd(*s.deviceName, session.User(), sspty.Term, session.Environ())

	pts, err := startPty(scmd, session, winCh)
	if err != nil {
		log.Warn(err)
	}

	u, err := osauth.LookupUser(session.User())
	if err != nil {
		return err
	}

	err = os.Chown(pts.Name(), int(u.UID), -1)
	if err != nil {
		log.Warn(err)
	}

	remoteAddr := session.RemoteAddr()

	log.WithFields(log.Fields{
		"user":       session.User(),
		"pty":        pts.Name(),
		"ispty":      isPty,
		"remoteaddr": remoteAddr,
		"localaddr":  session.LocalAddr(),
	}).Info("Session started")

	ut := utmp.UtmpStartSession(
		pts.Name(),
		session.User(),
		remoteAddr.String(),
	)

	s.mu.Lock()
	s.cmds[session.Context().Value(gliderssh.ContextKeySessionID).(string)] = scmd
	s.mu.Unlock()

	if err := scmd.Wait(); err != nil {
		log.Warn(err)
	}

	log.WithFields(log.Fields{
		"user":       session.User(),
		"pty":        pts.Name(),
		"remoteaddr": remoteAddr,
		"localaddr":  session.LocalAddr(),
	}).Info("Session ended")

	utmp.UtmpEndSession(ut)

	return nil
}

// Heredoc handles the server's SSH heredoc session when server is running in host mode.
//
// heredoc is special block of code that contains multi-line strings that will be redirected to a stdin of a shell. It
// request a shell, but doesn't allocate a pty.
func (s *Sessioner) Heredoc(session gliderssh.Session) error {
	_, _, isPty := session.Pty()

	cmd := newShellCmd(*s.deviceName, session.User(), "", session.Environ())

	stdout, _ := cmd.StdoutPipe()
	stdin, _ := cmd.StdinPipe()
	stderr, _ := cmd.StderrPipe()

	serverConn, ok := session.Context().Value(gliderssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		return fmt.Errorf("failed to get server connection from session context")
	}

	go func() {
		serverConn.Wait()  // nolint:errcheck
		cmd.Process.Kill() // nolint:errcheck
	}()

	log.WithFields(log.Fields{
		"user":        session.User(),
		"ispty":       isPty,
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command started")

	err := cmd.Start()
	if err != nil {
		log.Warn(err)
	}

	go func() {
		if _, err := io.Copy(stdin, session); err != nil {
			fmt.Println(err) //nolint:forbidigo
		}

		stdin.Close()
	}()

	go func() {
		combinedOutput := io.MultiReader(stdout, stderr)
		if _, err := io.Copy(session, combinedOutput); err != nil {
			fmt.Println(err) //nolint:forbidigo
		}
	}()

	err = cmd.Wait()
	if err != nil {
		log.Warn(err)
	}

	session.Exit(cmd.ProcessState.ExitCode()) //nolint:errcheck

	log.WithFields(log.Fields{
		"user":        session.User(),
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command ended")

	return nil
}

// Exec handles the SSH's server exec session when server is running in host mode.
func (s *Sessioner) Exec(session gliderssh.Session) error {
	if len(session.Command()) == 0 {
		log.WithFields(log.Fields{
			"user":      session.User(),
			"localaddr": session.LocalAddr(),
		}).Error("None command was received")

		log.Info("Session ended")
		_ = session.Exit(1)

		return nil
	}

	user, err := osauth.LookupUser(session.User())
	if err != nil {
		return err
	}

	sPty, sWinCh, sIsPty := session.Pty()

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = user.Shell
	}

	term := sPty.Term
	if sIsPty && term == "" {
		term = "xterm"
	}

	cmd := command.NewCmd(user, shell, term, *s.deviceName, session.Environ(), shell, "-c", session.RawCommand())

	wg := &sync.WaitGroup{}
	if sIsPty {
		pty, tty, err := initPty(cmd, session, sWinCh)
		if err != nil {
			log.Warn(err)
		}

		defer tty.Close()
		defer pty.Close()

		if err := os.Chown(tty.Name(), int(user.UID), -1); err != nil {
			log.Warn(err)
		}
	} else {
		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()
		stderr, _ := cmd.StderrPipe()

		// relay input from the SSH session to the command.
		go func() {
			if _, err := io.Copy(stdin, session); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}

			stdin.Close()
		}()

		wg.Add(1)

		// relay the command's combined output and error streams back to the SSH session.
		go func() {
			defer wg.Done()
			combinedOutput := io.MultiReader(stdout, stderr)
			if _, err := io.Copy(session, combinedOutput); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}
		}()
	}

	log.WithFields(log.Fields{
		"user":        session.User(),
		"ispty":       sIsPty,
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command started")

	if err := cmd.Start(); err != nil {
		return err
	}

	if !sIsPty {
		wg.Wait()
	}

	serverConn, ok := session.Context().Value(gliderssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		return fmt.Errorf("failed to get server connection from session context")
	}

	// kill the process if the SSH connection is interrupted
	go func() {
		serverConn.Wait()  // nolint:errcheck
		cmd.Process.Kill() // nolint:errcheck
	}()

	if err := cmd.Wait(); err != nil {
		log.Warn(err)
	}

	log.WithFields(log.Fields{
		"user":        session.User(),
		"ispty":       sIsPty,
		"remoteaddr":  session.RemoteAddr(),
		"localaddr":   session.LocalAddr(),
		"Raw command": session.RawCommand(),
	}).Info("Command ended")

	if err := session.Exit(cmd.ProcessState.ExitCode()); err != nil { // nolint:errcheck
		log.Warn(err)
	}

	return nil
}

// SFTP handles the SSH's server sftp session when server is running in host mode.
//
// sftp is a subsystem of SSH that allows file operations over SSH.
func (s *Sessioner) SFTP(session gliderssh.Session) error {
	log.WithFields(log.Fields{
		"user": session.Context().User(),
	}).Info("SFTP session started")
	defer session.Close()

	cmd := command.SFTPServerCommand()

	looked, err := user.Lookup(session.User())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to lookup user")

		return errors.New("failed to lookup user")
	}

	home := fmt.Sprintf("HOME=%s", looked.HomeDir)
	gid := fmt.Sprintf("GID=%s", looked.Gid)
	uid := fmt.Sprintf("UID=%s", looked.Uid)

	cmd.Env = append(cmd.Env, home)
	cmd.Env = append(cmd.Env, gid)
	cmd.Env = append(cmd.Env, uid)

	input, err := cmd.StdinPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stdin pipe")

		return errors.New("failed to get stdin pipe")
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stdout pipe")

		return errors.New("failed to get stdout pipe")
	}

	erro, err := cmd.StderrPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stderr pipe")

		return errors.New("failed to get stderr pipe")
	}

	if err := cmd.Start(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to start command")

		return errors.New("failed to start command")
	}

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying input to session")

		if _, err := io.Copy(input, session); err != nil && err != io.EOF {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stdin to command")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing input to session ends")

		input.Close()
	}()

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying output to session")

		if _, err := io.Copy(session, output); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stdout to session")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing output to session ends")
	}()

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying error to session")

		msgs := bufio.NewScanner(erro)
		msgs.Split(bufio.ScanLines)
		for msgs.Scan() {
			if err := msgs.Err(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user": session.Context().User(),
				}).Error("failed when reading the error output from sftp process")

				return
			}

			log.WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error(msgs.Text())
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing error to session ends")
	}()

	if err = cmd.Wait(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to wait command")

		return errors.New("failed to wait command")
	}

	log.WithFields(log.Fields{
		"user": session.Context().User(),
	}).Info("SFTP session closed")

	return nil
}
