package server

import (
	"net"
	"os"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/server/auth"
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	Addr           string
	Broker         string
	ConnectTimeout time.Duration
}

type Server struct {
	sshd   *gliderssh.Server
	opts   *Options
	tunnel *httptunnel.Tunnel
}

// NewServer create a new ShellHub's Connect server.
func NewServer(opts *Options, tunnel *httptunnel.Tunnel) *Server {
	server := &Server{ // nolint: exhaustruct
		opts:   opts,
		tunnel: tunnel,
	}

	server.sshd = &gliderssh.Server{ // nolint: exhaustruct
		Addr:             opts.Addr,
		PasswordHandler:  auth.PasswordHandler,
		PublicKeyHandler: auth.PublicKeyHandler,
		SessionRequestCallback: func(client gliderssh.Session, request string) bool {
			metadata.StoreRequest(client.Context(), request)

			return true
		},
		Handler: handler.SSHHandler(tunnel),
		SubsystemHandlers: map[string]gliderssh.SubsystemHandler{
			handler.SFTPSubsystem: handler.SFTPSubsystemHandler(tunnel),
		},
	}

	if _, err := os.Stat(os.Getenv("PRIVATE_KEY")); os.IsNotExist(err) {
		log.WithError(err).Fatal("private key not found!")
	}

	if err := server.sshd.SetOption(gliderssh.HostKeyFile(os.Getenv("PRIVATE_KEY"))); err != nil {
		log.WithError(err).Fatal("host key not found!")
	}

	return server
}

func (s *Server) ListenAndServe() error {
	log.WithFields(log.Fields{
		"addr": s.opts.Addr,
	}).Info("ssh server listening")

	list, err := net.Listen("tcp", s.opts.Addr)
	if err != nil {
		log.WithError(err).Error("failed to listen an serve the TCP server")

		return err
	}

	proxy := &proxyproto.Listener{Listener: list} // nolint: exhaustruct
	defer proxy.Close()

	return s.sshd.Serve(proxy)
}
