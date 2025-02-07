package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/host"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Command struct {
	Command string `json:"command"`
}

type Subsystem struct {
	Subsystem string `json:"subsystem"`
}

type Status struct {
	Status uint32 `json:"status"`
}

type Signal struct {
	Name    uint32 `json:"status"`
	Dumped  bool   `json:"dumped"`
	Message string `json:"message"`
	Lang    string `json:"lang"`
}

type Dimensions struct {
	Columns uint32 `json:"columns"`
	Rows    uint32 `json:"rows"`
	Width   uint32 `json:"width"`
	Height  uint32 `json:"height"`
}

// NOTICE: [Pty] cannot use [Dimensions] inside itself due [ssh.Unmarshal] issues.
type Pty struct {
	Term     string `json:"term"`
	Columns  uint32 `json:"columns" `
	Rows     uint32 `json:"rows"`
	Width    uint32 `json:"width"`
	Height   uint32 `json:"height"`
	Modelist []byte `json:"modelist"`
}

type Data struct {
	Target *target.Target
	// SSHID is the combination of device's name and namespace name.
	SSHID string
	// Device is the device connected.
	Device    *models.Device
	IPAddress string
	// Type is the connection type.
	Type string
	// Term is the terminal used for the client.
	Term string
	// TODO:
	Lookup map[string]string
	// Pty is the PTY dimension.
	Pty Pty
	// Handled check if the session is already handling a "shell", "exec" or a "subsystem".
	Handled bool
}

// TODO: implement [io.Read] and [io.Write] on session to simplify the data piping.
type Session struct {
	// UID is the session's UID.
	UID string

	// AgentConn is the connection between the Server and Agent.
	AgentConn net.Conn
	// AgentClient is a [gossh.Client] connected and authenticated to the agent, waiting for a open sesssion request.
	AgentClient *gossh.Client
	// AgentGlobalReqs is the channel to handle global request like "keepalive".
	AgentGlobalReqs <-chan *gossh.Request

	api    internalclient.Client
	tunnel *httptunnel.Tunnel

	once *sync.Once

	// Seat is a counter of how many passengers a session has. It's used on the record session feature.
	//
	// A passenger is, in a multiplexed SSH session, the subsequent SSH sessions that connect to the same server using
	// the already established master connection.
	Seat *atomic.Int32

	Data
}

// NewSession creates a new Session but differs from [New] as it only creates
// the session without registering, connecting to the agent and etc.
//
// It's designed to be used within New.
func NewSession(ctx gliderssh.Context, tunnel *httptunnel.Tunnel, cache cache.Cache) (*Session, error) {
	snap := getSnapshot(ctx)

	api, err := internalclient.NewClient()
	if err != nil {
		return nil, err
	}

	sshid := ctx.User()

	hos, err := host.NewHost(ctx.RemoteAddr().String())
	if err != nil {
		log.WithError(err).
			Error("failed to create a new host")

		return nil, ErrHost
	}

	target, err := target.NewTarget(sshid)
	if err != nil {
		return nil, err
	}

	var namespace, hostname string
	if target.IsSSHID() {
		namespace, hostname, err = target.SplitSSHID()
		if err != nil {
			return nil, err
		}
	} else {
		if hos.IsLocalhost() {
			var data string

			if err := cache.Get(ctx, "web-ip/"+sshid, &data); err != nil {
				log.WithError(err).
					Error("failed to get the ip from web session")

				return nil, err
			}

			if err := cache.Delete(ctx, "web-ip/"+sshid); err != nil {
				log.WithError(err).
					Error("failed to delete the web session ip from cache")

				return nil, err
			}

			parts := strings.Split(data, ":")
			target.Data = parts[0]
			hos.Host = parts[1]
		}

		device, err := api.GetDevice(target.Data)
		if err != nil {
			return nil, err
		}

		namespace = device.Namespace
		hostname = device.Name
	}

	lookup := map[string]string{
		"domain": namespace,
		"name":   hostname,
	}

	device, errs := api.DeviceLookup(lookup)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	session := &Session{
		UID:    ctx.SessionID(),
		api:    api,
		tunnel: tunnel,
		Data: Data{
			IPAddress: hos.Host,
			Target:    target,
			Device:    device,
			Lookup:    lookup,
			SSHID:     fmt.Sprintf("%s@%s.%s", target.Username, namespace, hostname),
		},
		once: new(sync.Once),
		Seat: new(atomic.Int32),
	}

	session.Data.Lookup["username"] = target.Username
	session.Data.Lookup["ip_address"] = hos.Host

	snap.save(session, StateCreated)

	return session, nil
}

func (s *Session) checkFirewall() (bool, error) {
	if err := s.api.FirewallEvaluate(s.Data.Lookup); err != nil {
		defer log.WithError(err).WithFields(log.Fields{
			"uid":   s.UID,
			"sshid": s.SSHID,
		}).Info("an error or a firewall rule block this connection")

		switch {
		case errors.Is(err, internalclient.ErrFirewallConnection):
			return false, ErrFirewallConnection
		case errors.Is(err, internalclient.ErrFirewallBlock):
			return false, ErrFirewallBlock
		default:
			return false, ErrFirewallUnknown
		}
	}

	return true, nil
}

func (s *Session) checkBilling() (bool, error) {
	device, err := s.api.GetDevice(s.Device.UID)
	if err != nil {
		defer log.WithError(err).WithFields(log.Fields{
			"uid":   s.UID,
			"sshid": s.SSHID,
		}).Info("failed to get the device on billing evaluation")

		return false, ErrFindDevice
	}

	if evaluatation, status, _ := s.api.BillingEvaluate(device.TenantID); status != 402 && !evaluatation.CanConnect {
		defer log.WithError(err).WithFields(log.Fields{
			"uid":   s.UID,
			"sshid": s.SSHID,
		}).Info("an error or a billing rule blocked this connection")

		return false, ErrBillingBlock
	}

	return true, nil
}

// registerAPISession registers a new session on the API.
func (s *Session) register() error {
	err := s.api.SessionCreate(requests.SessionCreate{
		UID:       s.UID,
		DeviceUID: s.Device.UID,
		Username:  s.Target.Username,
		IPAddress: s.IPAddress,
		Type:      "none",
		Term:      "none",
	})
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("Error when trying to register the client on API")

		return err
	}

	return nil
}

// Authenticate marks the session as authenticated on the API.
//
// It returns an error if authentication fails.
func (s *Session) authenticate() error {
	value := true

	return s.api.UpdateSession(s.UID, &models.SessionUpdate{
		Authenticated: &value,
	})
}

// connect connects the session's client to the session's agent.
func (s *Session) connect(ctx gliderssh.Context, authOpt authFunc) error {
	config := &gossh.ClientConfig{
		User:            s.Target.Username,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint: gosec
	}

	if err := authOpt(s, config); err != nil {
		return errors.New("fail to generate the authentication information")
	}

	const Addr = "tcp"

	// NOTICE: When the agent connection is closed, we should redial this connection before try to authenticate.
	if s.AgentConn == nil {
		if err := s.Dial(ctx); err != nil {
			return err
		}
	}

	if config.Timeout > 0 {
		if err := s.AgentConn.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline")

			return err
		}
	}

	conn, chans, reqs, err := gossh.NewClientConn(s.AgentConn, Addr, config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID}).
			Error("Error when trying to create the client's connection")

			// NOTICE: To help identifing when the Agent's connection is closed, we set it to nil when a authentication
			// error happens.
		s.AgentConn = nil

		return err
	}

	if config.Timeout > 0 {
		if err := s.AgentConn.SetReadDeadline(time.Time{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline with Time{}")

			return err
		}
	}

	ch := make(chan *gossh.Request)
	close(ch)

	s.AgentClient = gossh.NewClient(conn, chans, ch)
	s.AgentGlobalReqs = reqs

	return nil
}

func (s *Session) Dial(ctx gliderssh.Context) error {
	var err error

	ctx.Lock()
	s.AgentConn, err = s.tunnel.Dial(ctx, s.Device.TenantID+":"+s.Device.UID)
	if err != nil {
		return errors.Join(ErrDial, err)
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", s.UID), nil)
	if err = req.Write(s.AgentConn); err != nil {
		return err
	}
	ctx.Unlock()

	return nil
}

func (s *Session) Evaluate(ctx gliderssh.Context) error {
	snap := getSnapshot(ctx)

	if envs.IsCloud() || envs.IsEnterprise() {
		if ok, err := s.checkFirewall(); err != nil || !ok {
			return err
		}

		if envs.HasBilling() {
			if ok, err := s.checkBilling(); err != nil || !ok {
				return err
			}
		}
	}

	snap.save(s, StateEvaluated)

	return nil
}

// Auth authenticate a [Session] based on the provided context.
//
// As a client may try to create N sessions with the same context, a [snapshot] is used
// to save/retrieve the current session state. To illustrate a practical use of this
// pattern you can imagine a client that wants to connect to a specified device. It first
// calls the `PublicKeyHandler` with a specified context. At this stage, there are no
// sessions associated with the provided context, and a new one will be created. If it
// fails, the same client (and consequently the same context) will call the
// `PasswordHandler`, which also calls `session.New`. Since we have already created a
// session in the previous authentication attempt, instead of repeating all operations,
// we can safely retrieve the same session again but attempt authentication with a
// password this time.
//
// Next steps can use the context's snapshot to retrieve the created session. An error is
// returned if any occurs.
func (s *Session) Auth(ctx gliderssh.Context, auth Auth) error {
	snap := getSnapshot(ctx)

	// The following code is structured to be read from top to bottom, disregarding the
	// switch and case statements. These statements serve as a "cache" for handling
	// different states efficiently.
	sess, state := snap.retrieve()
	switch state {
	case StateEvaluated:
		if err := auth.Evaluate(sess); err != nil {
			return err
		}

		if err := sess.register(); err != nil {
			return err
		}

		snap.save(sess, StateRegistered)

		fallthrough
	case StateRegistered:
		if err := sess.connect(ctx, auth.Auth()); err != nil {
			return err
		}

		if err := sess.authenticate(); err != nil {
			return err
		}
	default:
		// The default arm is intended to avoid [StateNil] and [StateCreated], what are used before the authentication.
		return errors.New("invalid session state")
	}

	snap.save(sess, StateFinished)

	return nil
}

func (s *Session) Record(ctx context.Context, url string, seat int) (*Camera, error) {
	conn, err := s.api.RecordSession(ctx, s.UID, seat, url)
	if err != nil {
		log.WithError(err).Error("failed to start the record session process")

		return nil, err
	}

	return NewCamera(conn), nil
}

func (s *Session) NewSeat() (int, error) {
	seat := int(s.Seat.Load())
	defer s.Seat.Add(1)

	return seat, nil
}

// Events register an event to the session.
func (s *Session) Event(t string, data any, seat int) {
	go s.api.EventSession(s.UID, &models.SessionEvent{ //nolint:errcheck
		Type:      t,
		Timestamp: clock.Now(),
		Data:      data,
		Seat:      seat,
	})
}

func Event[D any](sess *Session, t string, data []byte, seat int) {
	d := new(D)
	if err := gossh.Unmarshal(data, d); err != nil {
		return
	}

	go sess.api.EventSession(sess.UID, &models.SessionEvent{ //nolint:errcheck
		Type:      t,
		Timestamp: clock.Now(),
		Data:      d,
		Seat:      seat,
	})
}

func (s *Session) KeepAlive() error {
	if errs := s.api.KeepAliveSession(s.UID); len(errs) > 0 {
		log.Error(errs[0])

		return errs[0]
	}

	return nil
}

// Announce is a custom message provided by the end user that can be printed when a new connection within the namespace
// is established.
//
// Returns the announcement or an error, if any. If no announcement is set, it returns an empty string.
func (s *Session) Announce(client gossh.Channel) error {
	if _, err := client.Write([]byte(
		"Connected to " + s.SSHID + " via ShellHub.\n\r",
	)); err != nil {
		return err
	}

	namespace, errs := s.api.
		NamespaceLookup(s.Device.TenantID)
	if len(errs) > 0 {
		log.WithError(errs[0]).Warn("unable to retrieve the namespace's connection announcement")

		return errs[0]
	}

	announcement := namespace.Settings.ConnectionAnnouncement

	if announcement == "" {
		return nil
	}

	// Remove whitespaces and new lines at end
	announcement = strings.TrimRightFunc(announcement, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t'
	})

	if _, err := client.Write([]byte(strings.ReplaceAll(announcement, "\n", "\n\r") + "\n\r")); err != nil {
		return err
	}

	return nil
}

// Finish terminate the session between Agent and Client, sending a request to Agent to closes it.
func (s *Session) Finish() (err error) {
	s.once.Do(func() {
		if s.AgentConn != nil {
			request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", s.UID), nil)

			if err = request.Write(s.AgentConn); err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
					Warning("Error when trying write the request to /ssh/close")
			}
		}

		if errs := s.api.FinishSession(s.UID); len(errs) > 0 {
			log.WithError(errs[0]).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to finish the session")

			err = errs[0]
		}

		log.WithFields(
			log.Fields{
				"uid":      s.UID,
				"device":   s.Device.UID,
				"username": s.Target.Username,
				"ip":       s.IPAddress,
			}).Info("session finished")
	})

	return nil
}
