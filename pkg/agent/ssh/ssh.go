package ssh

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	log "github.com/sirupsen/logrus"
)

// SSHPingDefaultInterval is the default time interval between ping on agent.
const SSHPingDefaultInterval = 10 * time.Minute

type SSH struct {
	Server    *Server
	tunnel    *tunnel.Tunnel
	cli       client.Client
	token     string
	listening chan bool
}

// NewSSH creates a new instance of SSH server.
func NewSSH(cli client.Client, token string) *SSH {
	return &SSH{
		cli:       cli,
		token:     token,
		listening: make(chan bool),
	}
}

func connSSHHandler(serv *Server) func(c echo.Context) error {
	return func(c echo.Context) error {
		hj, ok := c.Response().Writer.(http.Hijacker)
		if !ok {
			return c.String(http.StatusInternalServerError, "webserver doesn't support hijacking")
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			return c.String(http.StatusInternalServerError, "failed to hijack connection")
		}

		id := c.Param("id")
		httpConn := c.Request().Context().Value("http-conn").(net.Conn)
		serv.Sessions.Store(id, httpConn)
		serv.HandleConn(httpConn)

		conn.Close()

		return nil
	}
}

func httpHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		replyError := func(err error, msg string, code int) error {
			/*log.WithError(err).WithFields(log.Fields{
				"remote":    c.Request().RemoteAddr,
				"namespace": c.Request().Header.Get("X-Namespace"),
				"path":      c.Request().Header.Get("X-Path"),
				"version":   AgentVersion,
			}).Error(msg)*/

			return c.String(code, msg)
		}

		in, err := net.Dial("tcp", ":80")
		if err != nil {
			return replyError(err, "failed to connect to HTTP server on device", http.StatusInternalServerError)
		}

		defer in.Close()

		url, err := url.Parse(c.Request().Header.Get("X-Path"))
		if err != nil {
			return replyError(err, "failed to parse URL", http.StatusInternalServerError)
		}

		c.Request().URL.Scheme = "http"
		c.Request().URL = url

		if err := c.Request().Write(in); err != nil {
			return replyError(err, "failed to write request to the server on device", http.StatusInternalServerError)
		}

		out, _, err := c.Response().Hijack()
		if err != nil {
			return replyError(err, "failed to hijack connection", http.StatusInternalServerError)
		}

		defer out.Close() // nolint:errcheck

		if _, err := io.Copy(out, in); err != nil {
			return replyError(err, "failed to copy response from device service to client", http.StatusInternalServerError)
		}

		return nil
	}
}

func closeSSHHandler(serv *Server) func(c echo.Context) error {
	return func(c echo.Context) error {
		id := c.Param("id")
		serv.CloseSession(id)

		/*log.WithFields(
			log.Fields{
				"id":             id,
				"version":        AgentVersion,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
			},
		).Info("A tunnel connection was closed")*/

		return nil
	}
}

// Close closes the ShellHub Agent's listening, stoping it from receive new connection requests.
func (s *SSH) Close() error {
	return s.tunnel.Close()
}

func (s *SSH) Listen(ctx context.Context) error {
	s.tunnel = tunnel.NewBuilder().
		WithConnHandler(connSSHHandler(s.Server)).
		WithCloseHandler(closeSSHHandler(s.Server)).
		WithHTTPHandler(httpHandler()).
		Build()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			/*if a.isClosed() {
				log.WithFields(log.Fields{
					"version":        AgentVersion,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
				}).Info("Stopped listening for connections")

				cancel()

				return
			}*/

			listener, err := s.cli.NewReverseListener(ctx, s.token, "/ssh/connection")
			if err != nil {
				log.Error("Failed to connect to SSH server through reverse tunnel. Retry in 10 seconds")

				time.Sleep(time.Second * 10)

				continue
			}

			log.Info("SSH server connection established")

			{
				// NOTE: Tunnel'll only realize that it lost its connection to the ShellHub SSH when the next
				// "keep-alive" connection fails. As a result, it will take this interval to reconnect to its server.
				err := s.tunnel.Listen(listener)

				log.WithError(err).Warn("Tunnel listener closed")

				listener.Close() // nolint:errcheck
			}
		}
	}()

	<-ctx.Done()

	return s.Close()
}
