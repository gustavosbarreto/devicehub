package handler

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// Errors returned by handlers to client.
var (
	ErrRequestShell            = fmt.Errorf("failed to open a shell in the device")
	ErrRequestExec             = fmt.Errorf("failed to exec the command in the device")
	ErrRequestHeredoc          = fmt.Errorf("failed to exec the command as heredoc in the device")
	ErrRequestUnsupported      = fmt.Errorf("failed to get the request type")
	ErrPublicKey               = fmt.Errorf("failed to get the parsed public key")
	ErrPrivateKey              = fmt.Errorf("failed to get a key data from the server")
	ErrSigner                  = fmt.Errorf("failed to create a signer from the private key")
	ErrConnect                 = fmt.Errorf("failed to connect to device")
	ErrSession                 = fmt.Errorf("failed to create a session between the server to the agent")
	ErrGetAuth                 = fmt.Errorf("failed to get auth data from key")
	ErrWebData                 = fmt.Errorf("failed to get the data to connect to device")
	ErrFindDevice              = fmt.Errorf("failed to find the device")
	ErrFindPublicKey           = fmt.Errorf("failed to get the public key from the server")
	ErrEvaluatePublicKey       = fmt.Errorf("failed to evaluate the public key in the server")
	ErrForbiddenPublicKey      = fmt.Errorf("failed to use the public key for this action")
	ErrDataPublicKey           = fmt.Errorf("failed to parse the public key data")
	ErrSignaturePublicKey      = fmt.Errorf("failed to decode the public key signature")
	ErrVerifyPublicKey         = fmt.Errorf("failed to verify the public key")
	ErrSignerPublicKey         = fmt.Errorf("failed to signer the public key")
	ErrDialSSH                 = fmt.Errorf("failed to dial to connect to server")
	ErrEnvIPAddress            = fmt.Errorf("failed to set the env virable of ip address from client")
	ErrEnvWS                   = fmt.Errorf("failed to set the env virable of web socket from client")
	ErrPipe                    = fmt.Errorf("failed to pipe client data to agent")
	ErrPty                     = fmt.Errorf("failed to request the pty to agent")
	ErrShell                   = fmt.Errorf("failed to get the shell to agent")
	ErrTarget                  = fmt.Errorf("failed to get client target")
	ErrAuthentication          = fmt.Errorf("failed to authenticate to device")
	ErrEnvs                    = fmt.Errorf("failed to parse server envs")
	ErrConfiguration           = fmt.Errorf("failed to create communication configuration")
	ErrInvalidVersion          = fmt.Errorf("failed to parse device version")
	ErrUnsuportedPublicKeyAuth = fmt.Errorf("connections using public keys are not permitted when the agent version is 0.5.x or earlier")
)

type ConfigOptions struct {
	RecordURL string `env:"RECORD_URL"`

	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

func parseConfig() (*ConfigOptions, error) {
	return envs.Parse[ConfigOptions]()
}

// SSHHandler handlers a "normal" SSH connection.
func SSHHandler(_ *httptunnel.Tunnel) gliderssh.Handler {
	return func(client gliderssh.Session) {
		log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection started")
		defer log.WithFields(log.Fields{"sshid": client.User()}).Info("SSH connection closed")

		defer client.Close()

		// TODO:
		sess := client.Context().Value("session").(*session.Session)
		sess.SetClientSession(client)

		opts, err := parseConfig()
		if err != nil {
			echo(sess.UID, client, err, "Error while parsing envs")

			return
		}

		// When the Shellhub instance dennies connections with
		// potentially broken agents, we need to evaluate the connection's context
		// and identify potential bugs. The server must reject the connection
		// if there's a possibility of issues; otherwise, proceeds.
		if err := evaluateContext(client, opts); err != nil {
			echo(sess.UID, client, err, "Error while evaluating context")

			return
		}

		agent, reqs, err := sess.NewAgentSession()
		if err != nil {
			echo(sess.UID, client, err, "Error when trying to start the agent's session")

			return
		}
		defer agent.Close()

		if err := connectSSH(sess, reqs, *opts); err != nil {
			echo(sess.UID, client, err, "Error during SSH connection")

			return
		}
	}
}

func connectSSH(sess *session.Session, reqs <-chan *gossh.Request, opts ConfigOptions) error {
	api := metadata.RestoreAPI(sess.Client.Context())

	go session.HandleRequests(sess.Client.Context(), reqs, api, sess.Client.Context().Done())

	switch sess.GetType() {
	case session.Term, session.Web:
		if err := shell(sess, sess.Client, sess.Agent, api, opts); err != nil {
			return ErrRequestShell
		}
	case session.HereDoc:
		err := heredoc(sess, sess.Client, sess.Agent, api)
		if err != nil {
			return ErrRequestHeredoc
		}
	case session.Exec, session.SCP:
		device := metadata.RestoreDevice(sess.Client.Context())

		if err := exec(sess, sess.Client, sess.Agent, api, device); err != nil {
			return ErrRequestExec
		}
	default:
		if err := sess.Client.Exit(255); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": sess.UID, "sshid": sess.Client.User()}).
				Warning("exiting client returned an error")
		}

		return ErrRequestUnsupported
	}

	return nil
}

func shell(sess *session.Session, client gliderssh.Session, agent *gossh.Session, api internalclient.Client, opts ConfigOptions) error {
	uid := sess.UID

	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	pty, winCh, _ := client.Pty()

	log.WithFields(log.Fields{"session": uid, "sshid": client.User()}).
		Debug("requesting a PTY for session")

	if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to request a PTY")

		return err
	}

	go resizeWindow(uid, agent, winCh)

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to create a flow of data from agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, done)

	go func() {
		buffer := make([]byte, 1024)
		for {
			read, err := flw.Stdout.Read(buffer)
			// The occurrence of io.EOF is expected when the connection ends.
			// This indicates that we have reached the end of the input stream, and we need
			// to break out of the loop to handle the termination of the connection
			if err == io.EOF {
				break
			}
			// Unlike io.EOF, when 'err' is simply not nil, it signifies an unexpected error,
			// and we need to log to handle it appropriately.
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": uid, "sshid": client.User()}).
					Warning("failed to read from stdout in pty client")

				break
			}

			if _, err = io.Copy(client, bytes.NewReader(buffer[:read])); err != nil && err != io.EOF {
				log.WithError(err).
					WithFields(log.Fields{"session": uid, "sshid": client.User()}).
					Warning("failed to copy from stdout in pty client")

				break
			}

			if envs.IsEnterprise() || envs.IsCloud() {
				message := string(buffer[:read])

				api.RecordSession(&models.SessionRecorded{
					UID:       uid,
					Namespace: sess.Lookup["domain"],
					Message:   message,
					Width:     pty.Window.Height,
					Height:    pty.Window.Width,
				}, opts.RecordURL)
			}
		}
	}()

	go flw.PipeErr(client.Stderr(), nil)

	go func() {
		// When agent stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Shell(); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to start a new shell")

		return err
	}

	// Writes the welcome message. The message consists of a static string about the device
	// and an optional custom string provided by the user.
	if target := metadata.RestoreTarget(client.Context()); target != nil {
		sess.Client.Write([]byte( // nolint: errcheck
			"Connected to " + target.Username + "@" + target.Data + " via ShellHub.\n",
		))
	}

	announcement, err := sess.ConnectionAnnouncement()
	if err != nil {
		log.WithError(err).Warn("unable to retrieve the namespace's connection announcement")
	} else if announcement != "" {
		sess.Client.Write([]byte("Announcement:\n")) // nolint: errcheck

		// Remove whitespaces and new lines at end
		announcement = strings.TrimRightFunc(announcement, func(r rune) bool {
			return r == ' ' || r == '\n' || r == '\t'
		})

		sess.Client.Write([]byte("    " + strings.ReplaceAll(announcement, "\n", "\n    ") + "\n")) // nolint: errcheck
	}

	if err := agent.Wait(); isUnknownExitError(err) {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Warning("client remote shell returned an error")
	}

	// We can safely ignore EOF errors on exit
	if err := client.Exit(0); err != nil && err != io.EOF {
		log.WithError(err).
			WithFields(log.Fields{"session": sess.UID, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}

func heredoc(sess *session.Session, client gliderssh.Session, agent *gossh.Session, api internalclient.Client) error {
	uid := sess.UID

	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to create a flow of data from agent")

		return err
	}

	done := make(chan bool)

	go flw.PipeIn(client, nil)
	go flw.PipeOut(client, done)
	go flw.PipeErr(client.Stderr(), nil)

	go func() {
		// When agent stop to send data, it means that the command has finished and the process should be closed.
		<-done

		agent.Close()
	}()

	if err := agent.Shell(); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to start a new shell")

		return err
	}

	if err := agent.Wait(); isUnknownExitError(err) {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Warning("command on agent returned an error")
	}

	if err := client.Exit(exitCodeFromError(err)); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}

func exec(sess *session.Session, client gliderssh.Session, agent *gossh.Session, api internalclient.Client, device *models.Device) error {
	uid := sess.UID

	if errs := api.SessionAsAuthenticated(uid); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to authenticate the session")

		return errs[0]
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("failed to create a flow of data from agent to agent")

		return err
	}

	// request a new pty when isPty is true
	pty, winCh, isPty := client.Pty()
	if isPty {
		log.WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Debug("requesting a PTY for session")

		if err := agent.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
				Error("failed to request a PTY")

			return err
		}
	}

	if isPty {
		go resizeWindow(uid, agent, winCh)
	}

	waitPipeIn := make(chan bool)
	waitPipeOut := make(chan bool)

	go flw.PipeIn(client, waitPipeIn)
	go flw.PipeOut(client, waitPipeOut)
	go flw.PipeErr(client.Stderr(), nil)

	if err := agent.Start(client.RawCommand()); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User(), "command": client.RawCommand()}).
			Error("failed to start a command on agent")

		return err
	}

	if device.Info.Version != "latest" {
		ver, err := semver.NewVersion(device.Info.Version)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
				Error("failed to parse device version")

			return err
		}

		// version less 0.9.3 does not support the exec command, what will make some commands to hang forever.
		if ver.LessThan(semver.MustParse("0.9.3")) {
			go func() {
				// When agent stop to send data, it means that the command has finished and the process should be closed.
				<-waitPipeIn
				agent.Close()
			}()
		}
	}

	// When agent stop to send data, it means that the command has finished and the process should be closed.
	<-waitPipeOut
	agent.Close()

	if err = agent.Wait(); isUnknownExitError(err) {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User(), "command": client.RawCommand()}).
			Warning("command on agent returned an error")
	}

	if err := client.Exit(exitCodeFromError(err)); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Warning("exiting client returned an error")
	}

	return nil
}
