package session

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/host"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Session struct {
	Client gliderssh.Session
	// Username is the user that is trying to connect to the device; user on device.
	Username string `json:"username"`
	Device   string `json:"device_uid"` // nolint: tagliatelle
	// UID is the device's UID.
	UID           string `json:"uid"`
	IPAddress     string `json:"ip_address"` // nolint: tagliatelle
	Type          string `json:"type"`
	Term          string `json:"term"`
	Authenticated bool   `json:"authenticated"`
	Lookup        map[string]string
	Pty           bool
	Dialed        net.Conn
}

// checkFirewall evaluates if there are firewall rules that block the connection.
func (s *Session) checkFirewall(ctx gliderssh.Context) (bool, error) {
	api := metadata.RestoreAPI(ctx)
	lookup := metadata.RestoreLookup(ctx)

	if envs.IsCloud() || envs.IsEnterprise() {
		if err := api.FirewallEvaluate(lookup); err != nil {
			switch {
			case errors.Is(err, internalclient.ErrFirewallConnection):
				return false, errors.Join(ErrFirewallConnection, err)
			case errors.Is(err, internalclient.ErrFirewallBlock):
				return false, errors.Join(ErrFirewallBlock, err)
			default:
				return false, errors.Join(ErrFirewallUnknown, err)
			}
		}
	}

	return true, nil
}

// checkBilling evaluates if the device's namespace has pending payment questions.
func (s *Session) checkBilling(ctx gliderssh.Context, device string) (bool, error) {
	api := metadata.RestoreAPI(ctx)

	if envs.IsCloud() && envs.HasBilling() {
		device, err := api.GetDevice(device)
		if err != nil {
			return false, errors.Join(ErrFindDevice, err)
		}

		if evaluatation, status, _ := api.BillingEvaluate(device.TenantID); status != 402 && !evaluatation.CanConnect {
			return false, errors.Join(ErrBillingBlock, err)
		}
	}

	return true, nil
}

// dial dials the a connection between SSH server and the device agent.
func (s *Session) dial(ctx gliderssh.Context, tunnel *httptunnel.Tunnel, device string, session string) (net.Conn, error) {
	dialed, err := tunnel.Dial(ctx, device)
	if err != nil {
		return nil, errors.Join(ErrDial, err)
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", session), nil)
	if err = req.Write(dialed); err != nil {
		return nil, err
	}

	return dialed, nil
}

// NewSession creates a new Client from a client to agent, validating data, instance and payment.
func NewSession(client gliderssh.Session, tunnel *httptunnel.Tunnel) (*Session, error) {
	ctx := client.Context()

	hos, err := host.NewHost(client.RemoteAddr().String())
	if err != nil {
		return nil, ErrHost
	}

	if hos.IsLocalhost() {
		env := loadEnv(client.Environ())
		if value, ok := env["IP_ADDRESS"]; ok {
			hos.Host = value
		}
	}

	uid := ctx.Value(gliderssh.ContextKeySessionID).(string) //nolint:forcetypeassert

	device := metadata.RestoreDevice(ctx)
	target := metadata.RestoreTarget(ctx)
	lookup := metadata.RestoreLookup(ctx)

	lookup["username"] = target.Username
	lookup["ip_address"] = hos.Host

	session := new(Session)
	if ok, err := session.checkFirewall(ctx); err != nil || !ok {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("Error when trying to evaluate firewall rules")

		return nil, err
	}

	if ok, err := session.checkBilling(ctx, device.UID); err != nil || !ok {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("Error when trying to evaluate billing")

		return nil, err
	}

	dialed, err := session.dial(ctx, tunnel, device.UID, uid)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("Error when trying to dial")

		return nil, ErrDial
	}

	session.Client = client
	session.UID = uid
	session.Username = target.Username
	session.IPAddress = hos.Host
	session.Device = device.UID
	session.Lookup = lookup
	session.Dialed = dialed

	session.setPty()
	session.setType()

	session.Register(client) // nolint:errcheck

	return session, nil
}

// NewSessionWithoutClient creates a new session to connect the agent, validating data, instance and payment.
//
// This function is used to create a new session when the client is not available, what is true when the SSH client
// indicate that the request type is `none` or in the case of a port forwarding
func NewSessionWithoutClient(ctx gliderssh.Context, tunnel *httptunnel.Tunnel) (*Session, error) {
	uid := ctx.Value(gliderssh.ContextKeySessionID).(string) //nolint:forcetypeassert

	hos, err := host.NewHost(ctx.RemoteAddr().String())
	if err != nil {
		return nil, ErrHost
	}

	device := metadata.RestoreDevice(ctx)
	target := metadata.RestoreTarget(ctx)
	lookup := metadata.RestoreLookup(ctx)

	lookup["username"] = target.Username
	lookup["ip_address"] = hos.Host

	session := new(Session)
	if ok, err := session.checkFirewall(ctx); err != nil || !ok {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": target.Username}).
			Error("Error when trying to evaluate firewall rules")

		return nil, err
	}

	if ok, err := session.checkBilling(ctx, device.UID); err != nil || !ok {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": target.Username}).
			Error("Error when trying to evaluate billing")

		return nil, err
	}

	dialed, err := session.dial(ctx, tunnel, device.UID, uid)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": target.Username}).
			Error("Error when trying to dial")

		return nil, ErrDial
	}

	session.Client = nil
	session.UID = uid
	session.Username = target.Username
	session.IPAddress = hos.Host
	session.Device = device.UID
	session.Lookup = lookup
	session.Dialed = dialed

	return session, nil
}

func (s *Session) GetType() string {
	return s.Type
}

// NewClientConnWithDeadline creates a new connection to the agent.
func (s *Session) NewClientConnWithDeadline(config *gossh.ClientConfig) (*gossh.Client, <-chan *gossh.Request, error) {
	const Addr = "tcp"

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
				Error("Error when trying to set dial deadline")

			return nil, nil, err
		}
	}

	cli, chans, reqs, err := gossh.NewClientConn(s.Dialed, Addr, config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID}).
			Error("Error when trying to create the client's connection")

		return nil, nil, err
	}

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(time.Time{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
				Error("Error when trying to set dial deadline with Time{}")

			return nil, nil, err
		}
	}

	ch := make(chan *gossh.Request)
	close(ch)

	return gossh.NewClient(cli, chans, ch), reqs, nil
}

// Register registers a new Client at the api.
func (s *Session) Register(_ gliderssh.Session) error {
	err := internalclient.
		NewClient().
		SessionCreate(requests.SessionCreate{
			UID:       s.UID,
			DeviceUID: s.Device,
			Username:  s.Username,
			IPAddress: s.IPAddress,
			Type:      s.Type,
			Term:      s.Term,
		})
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
			Error("Error when trying to register the client on API")

		return err
	}

	return nil
}

// ConnectionAnnouncement retrieves the connection announcement of the device's namespace.
// A connection announcement is a custom message provided by the end user that can be printed
// when a new connection within the namespace is established.
//
// Returns the announcement or an error, if any. If no announcement is set, it returns an empty string.
func (s *Session) ConnectionAnnouncement() (string, error) {
	device := metadata.RestoreDevice(s.Client.Context())
	if device == nil {
		return "", nil
	}

	namespace, errs := internalclient.
		NewClient().
		NamespaceLookup(device.TenantID)
	if len(errs) > 0 {
		return "", errs[0]
	}

	return namespace.Settings.ConnectionAnnouncement, nil
}

func (s *Session) Finish() error {
	if s.Dialed != nil {
		request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", s.UID), nil)

		if err := request.Write(s.Dialed); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
				Warning("Error when trying write the request to /ssh/close")
		}
	}

	if errs := internalclient.NewClient().FinishSession(s.UID); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
			Error("Error when trying to finish the session")

		return errs[0]
	}

	return nil
}

// NewClientConfiguration creates a [gossh.ClientConfig] with the default configuration required by ShellHub
// to connect to the device agent that are inside the [gliderssh.Context].
func NewClientConfiguration(ctx gliderssh.Context) (*gossh.ClientConfig, error) {
	target := metadata.RestoreTarget(ctx)
	if target == nil {
		return nil, errors.New("failed to get the target from context")
	}

	config := &gossh.ClientConfig{
		User:            target.Username,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint: gosec
	}

	api := metadata.RestoreAPI(ctx)
	if api == nil {
		return nil, errors.New("failed to get the API from context")
	}

	switch metadata.RestoreAuthenticationMethod(ctx) {
	case metadata.PublicKeyAuthenticationMethod:
		privateKey, err := api.CreatePrivateKey()
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(privateKey.Data)

		parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		signer, err := gossh.NewSignerFromKey(parsed)
		if err != nil {
			return nil, err
		}

		config.Auth = []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		}
	case metadata.PasswordAuthenticationMethod:
		password := metadata.RestorePassword(ctx)

		config.Auth = []gossh.AuthMethod{
			gossh.Password(password),
		}
	}

	return config, nil
}
