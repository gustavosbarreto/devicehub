package sshd

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"sync"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type sshConn struct {
	net.Conn
	closeCallback func(string)
	ctx           sshserver.Context
}

func (c *sshConn) Close() error {
	if id, ok := c.ctx.Value(sshserver.ContextKeySessionID).(string); ok {
		c.closeCallback(id)
	}

	return c.Conn.Close()
}

type Server struct {
	sshd               *sshserver.Server
	api                client.Client
	authData           *models.DeviceAuthResponse
	cmds               map[string]*exec.Cmd
	Sessions           map[string]net.Conn
	deviceName         string
	mu                 sync.Mutex
	keepAliveInterval  int
	singleUserPassword string
}

func SFTPSubsystemHandler(sess sshserver.Session) {
	logrus.WithFields(logrus.Fields{
		"user": sess.Context().User(),
	}).Info("SFTP session started")
	defer sess.Close()

	cmd := exec.Command("/proc/self/exe", []string{"sftp"}...)

	looked, err := user.Lookup(sess.User())
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Error("Failed to lookup user")

		return
	}

	home := fmt.Sprintf("HOME=%s", looked.HomeDir)
	gid := fmt.Sprintf("GID=%s", looked.Gid)
	uid := fmt.Sprintf("UID=%s", looked.Uid)

	cmd.Env = append(cmd.Env, home)
	cmd.Env = append(cmd.Env, gid)
	cmd.Env = append(cmd.Env, uid)

	input, err := cmd.StdinPipe()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Error("Failed to get stdin pipe")

		return
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Error("Failed to get stdout pipe")

		return
	}

	erro, err := cmd.StderrPipe()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Error("Failed to get stderr pipe")

		return
	}

	if err := cmd.Start(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Error("Failed to start command")

		return
	}

	go func() {
		logrus.WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Trace("copying input to session")

		if _, err := io.Copy(input, sess); err != nil && err != io.EOF {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user": sess.Context().User(),
			}).Error("Failed to copy stdin to command")

			return
		}

		logrus.WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Trace("closing input to session ends")

		input.Close()
	}()

	go func() {
		logrus.WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Trace("copying output to session")

		if _, err := io.Copy(sess, output); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user": sess.Context().User(),
			}).Error("Failed to copy stdout to session")

			return
		}

		logrus.WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Trace("closing output to session ends")
	}()

	go func() {
		logrus.WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Trace("copying error to session")

		if _, err := io.Copy(sess, erro); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user": sess.Context().User(),
			}).Error("Failed to copy stderr to session")

			return
		}

		logrus.WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Trace("closing error to session ends")
	}()

	if err = cmd.Wait(); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user": sess.Context().User(),
		}).Error("Failed to wait command")

		return
	}

	logrus.WithFields(logrus.Fields{
		"user": sess.Context().User(),
	}).Info("SFTP session closed")
}

func NewServer(api client.Client, authData *models.DeviceAuthResponse, privateKey string, keepAliveInterval int, singleUserPassword string) *Server {
	server := &Server{
		api:               api,
		authData:          authData,
		cmds:              make(map[string]*exec.Cmd),
		Sessions:          make(map[string]net.Conn),
		keepAliveInterval: keepAliveInterval,
	}

	server.sshd = &sshserver.Server{
		PasswordHandler:        server.passwordHandler,
		PublicKeyHandler:       server.publicKeyHandler,
		Handler:                server.sessionHandler,
		SessionRequestCallback: server.sessionRequestCallback,
		RequestHandlers:        sshserver.DefaultRequestHandlers,
		ChannelHandlers:        sshserver.DefaultChannelHandlers,
		ConnCallback: func(ctx sshserver.Context, conn net.Conn) net.Conn {
			closeCallback := func(id string) {
				server.mu.Lock()
				defer server.mu.Unlock()

				if v, ok := server.cmds[id]; ok {
					v.Process.Kill() // nolint:errcheck
					delete(server.cmds, id)
				}
			}

			return &sshConn{conn, closeCallback, ctx}
		},
		SubsystemHandlers: map[string]sshserver.SubsystemHandler{
			"sftp": SFTPSubsystemHandler,
		},
	}

	err := server.sshd.SetOption(sshserver.HostKeyFile(privateKey))
	if err != nil {
		logrus.Warn(err)
	}

	return server
}

func (s *Server) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}

func (s *Server) HandleConn(conn net.Conn) {
	s.sshd.HandleConn(conn)
}

func (s *Server) SetDeviceName(name string) {
	s.deviceName = name
}

func (s *Server) sessionHandler(session sshserver.Session) {
	sspty, winCh, isPty := session.Pty()

	log := logrus.WithFields(logrus.Fields{
		"user": session.User(),
		"pty":  isPty,
	})

	log.Info("New session request")

	go StartKeepAliveLoop(time.Second*time.Duration(s.keepAliveInterval), session)
	requestType := session.Context().Value("request_type").(string)

	switch {
	case isPty:
		scmd := newShellCmd(s, session.User(), sspty.Term)

		pts, err := startPty(scmd, session, winCh)
		if err != nil {
			logrus.Warn(err)
		}

		u := osauth.LookupUser(session.User())

		err = os.Chown(pts.Name(), int(u.UID), -1)
		if err != nil {
			logrus.Warn(err)
		}

		remoteAddr := session.RemoteAddr()

		logrus.WithFields(logrus.Fields{
			"user":       session.User(),
			"pty":        pts.Name(),
			"remoteaddr": remoteAddr,
			"localaddr":  session.LocalAddr(),
		}).Info("Session started")

		ut := utmpStartSession(
			pts.Name(),
			session.User(),
			remoteAddr.String(),
		)

		s.mu.Lock()
		s.cmds[session.Context().Value(sshserver.ContextKeySessionID).(string)] = scmd
		s.mu.Unlock()

		if err := scmd.Wait(); err != nil {
			logrus.Warn(err)
		}

		logrus.WithFields(logrus.Fields{
			"user":       session.User(),
			"pty":        pts.Name(),
			"remoteaddr": remoteAddr,
			"localaddr":  session.LocalAddr(),
		}).Info("Session ended")

		utmpEndSession(ut)
	case !isPty && requestType == "shell":
		cmd := newShellCmd(s, session.User(), "")

		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()
		stderr, _ := cmd.StderrPipe()

		serverConn, ok := session.Context().Value(sshserver.ContextKeyConn).(*ssh.ServerConn)
		if !ok {
			return
		}

		go func() {
			serverConn.Wait()  // nolint:errcheck
			cmd.Process.Kill() // nolint:errcheck
		}()

		logrus.WithFields(logrus.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command started")

		err := cmd.Start()
		if err != nil {
			logrus.Warn(err)
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
			logrus.Warn(err)
		}

		session.Exit(cmd.ProcessState.ExitCode()) //nolint:errcheck

		logrus.WithFields(logrus.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command ended")
	default:
		u := osauth.LookupUser(session.User())
		if len(session.Command()) == 0 {
			logrus.WithFields(logrus.Fields{
				"user":      session.User(),
				"localaddr": session.LocalAddr(),
			}).Errorln("None command was received")
			logrus.Infoln("Session ended")
			_ = session.Exit(1)

			return
		}

		cmd := newCmd(u, "", "", s.deviceName, session.Command()...)

		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()
		stderr, _ := cmd.StderrPipe()

		serverConn, ok := session.Context().Value(sshserver.ContextKeyConn).(*ssh.ServerConn)
		if !ok {
			return
		}

		go func() {
			serverConn.Wait()  // nolint:errcheck
			cmd.Process.Kill() // nolint:errcheck
		}()

		logrus.WithFields(logrus.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command started")

		err := cmd.Start()
		if err != nil {
			logrus.Warn(err)
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
			logrus.Warn(err)
		}

		session.Exit(cmd.ProcessState.ExitCode()) //nolint:errcheck

		logrus.WithFields(logrus.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command ended")
	}
}

func (s *Server) passwordHandler(ctx sshserver.Context, pass string) bool {
	log := logrus.WithFields(logrus.Fields{
		"user": ctx.User(),
	})
	var ok bool

	if s.singleUserPassword == "" {
		ok = osauth.AuthUser(ctx.User(), pass)
	} else {
		ok = osauth.VerifyPasswordHash(s.singleUserPassword, pass)
	}

	if ok {
		log.Info("Accepted password")
	} else {
		log.Info("Failed password")
	}

	return ok
}

func (s *Server) publicKeyHandler(ctx sshserver.Context, key sshserver.PublicKey) bool {
	if osauth.LookupUser(ctx.User()) == nil {
		return false
	}

	type Signature struct {
		Username  string
		Namespace string
	}

	sig := &Signature{
		Username:  ctx.User(),
		Namespace: s.deviceName,
	}

	sigBytes, err := json.Marshal(sig)
	if err != nil {
		return false
	}

	sigHash := sha256.Sum256(sigBytes)

	res, err := s.api.AuthPublicKey(&models.PublicKeyAuthRequest{
		Fingerprint: ssh.FingerprintLegacyMD5(key),
		Data:        string(sigBytes),
	}, s.authData.Token)
	if err != nil {
		return false
	}

	digest, err := base64.StdEncoding.DecodeString(res.Signature)
	if err != nil {
		return false
	}

	cryptoKey, ok := key.(ssh.CryptoPublicKey)
	if !ok {
		return false
	}

	pubCrypto := cryptoKey.CryptoPublicKey()

	pubKey, ok := pubCrypto.(*rsa.PublicKey)
	if !ok {
		return false
	}

	if err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sigHash[:], digest); err != nil {
		return false
	}

	return true
}

func (s *Server) CloseSession(id string) {
	if session, ok := s.Sessions[id]; ok {
		session.Close()
		delete(s.Sessions, id)
	}
}

func (s *Server) sessionRequestCallback(session sshserver.Session, requestType string) bool {
	session.Context().SetValue("request_type", requestType)

	return true
}

func newShellCmd(s *Server, username, term string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	u := osauth.LookupUser(username)

	if shell == "" {
		shell = u.Shell
	}

	if term == "" {
		term = "xterm"
	}

	cmd := newCmd(u, shell, term, s.deviceName, shell, "--login")

	return cmd
}
