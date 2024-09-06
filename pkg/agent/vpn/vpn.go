package vpn

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/shellhub-io/shellhub/pkg/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/pkg/agent/vpn/pkg/ifce"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	log "github.com/sirupsen/logrus"
)

type VPN struct {
	// tunnel is the reverse WebSocket connection between Agent and ShellHub's server.
	tunnel *tunnel.Tunnel
	// httpc is the HTTP client for the ShellHub's server.
	httpc client.Client
	// token is the JWT token used to operate on ShellHub's server.
	// TODO: insert the token into the HTTP client.
	token string
	// done is a channel used to indicate to the connection handler that the connection was closed.
	done chan struct{}
}

// ConnectEndpoint is used by ShellHub's server to start a new VPN connection with the Agent.
const ConnectEndpoint string = "/vpn/connect"

// CloseEndpoint is used by ShellHub's server to close a VPN connection.
const CloseEndpoint string = "/vpn/close"

// NewVPN creates a new instance of VPN client.
func NewVPN(cli client.Client, token string) *VPN {
	return &VPN{
		tunnel: tunnel.NewCustomTunnel(ConnectEndpoint, CloseEndpoint),
		httpc:  cli,
		token:  token,
		done:   make(chan struct{}),
	}
}

// Handler handles the connection established between the ShellHub's server to Agent, starting the packet transmission.
func (s *VPN) Handler(conn net.Conn, settings *Settings) error {
	log.Debug("vpn connection accepted")
	defer log.Debug("vpn connection closed")

	log.WithFields(log.Fields{
		"addrss": settings.Address,
		"mask":   settings.Mask,
	}).Debug("interface data")

	face, err := ifce.NewInterface(settings.String())
	if err != nil {
		log.WithError(err).Error("failed to create or configure the interface")

		return err
	}

	defer face.Close()

	log.WithFields(log.Fields{
		"interface": face.Name(),
	}).Debug("interface create")

	if err := face.Up(); err != nil {
		log.WithError(err).Error("failed to get up the interface")

		return err
	}

	log.WithFields(log.Fields{
		"interface": face.Name(),
	}).Debug("interface up")

	wg := new(sync.WaitGroup)

	// done closes the connection between the ShellHub's server and the network interface on the Agent.
	done := sync.OnceFunc(func() {
		log.Trace("conn and ifce connections closed")

		conn.Close()
		face.Close()

		s.tunnel.Close()
	})

	go func() {
		<-s.done

		log.Trace("message on done channel received")

		done()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Trace("reading from interface done")
		defer done()

		buffer := make([]byte, ifce.MaximumTransmissionUnit)

		for {
			read, err := conn.Read(buffer)
			if err != nil {
				log.WithError(err).Error("failed to read from connection to interface")

				return
			}

			if read == 0 {
				continue
			}

			if _, err := face.Write(buffer[:read]); err != nil {
				log.WithError(err).Error("failed to write to interface")

				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Trace("reading from conn done")
		defer done()

		buffer := make([]byte, ifce.MaximumTransmissionUnit)

		for {
			read, err := face.Read(buffer)
			if err != nil {
				log.WithError(err).Error("failed to read from interface to connection")

				return
			}

			if read == 0 {
				continue
			}

			if _, err := conn.Write(buffer[:read]); err != nil {
				log.WithError(err).Error("failed to write to connection")
			}
		}
	}()

	log.WithFields(log.Fields{
		"address":   settings.String(),
		"interface": face.Name(),
	}).Info("VPN connection started")

	wg.Wait()
	return nil
}

// Close closes the ShellHub Agent's listening, stoping it from receive new connection requests.
func (s *VPN) Close() error {
	// NOTE: It sends a close message to the handler.
	s.done <- struct{}{}

	return s.tunnel.Close()
}

var (
	ErrConnectListen     = errors.New("listen closed on vpn connection")
	ErrConnectionReverse = errors.New("reverse connection lost")
)

func (s *VPN) Connect(ctx context.Context) error {
	s.tunnel.ConnHandler = handler(s.Handler)
	s.tunnel.CloseHandler = closeHandler(s.Close)

	listener, err := s.httpc.NewReverseListener(ctx, s.token, "/vpn/connection")
	if err != nil {
		return errors.Join(ErrConnectionReverse, err)
	}

	defer listener.Close()

	go func() {
		<-ctx.Done()

		log.Trace("message on ctx channel received")

		s.Close()
	}()

	if err := s.tunnel.Listen(listener); err != nil {
		return errors.Join(ErrConnectListen, err)
	}

	return nil
}
