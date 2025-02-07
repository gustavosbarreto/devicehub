package channels

import (
	"strings"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// KeepAliveRequestTypePrefix Through the time, the [KeepAliveRequestType] type sent from agent to server changed its
// name, but always keeping the prefix "keepalive". So, to maintain the retro compatibility, we check if this prefix
// exists and perform the necessary operations.
const KeepAliveRequestTypePrefix string = "keepalive"

const (
	// Once the session has been set up, a program is started at the remote end.  The program can be a shell, an
	// application program, or a subsystem with a host-independent name.  Only one of these requests can succeed per
	// channel
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	ShellRequestType = "shell"
	// This message will request that the server start the execution of the given command.  The 'command' string may
	// contain a path.  Normal precautions MUST be taken to prevent the execution of unauthorized commands.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	ExecRequestType = "exec"
	// This last form executes a predefined subsystem.  It is expected that these will include a general file transfer
	// mechanism, and possibly other features.  Implementations may also allow configuring more such mechanisms.  As
	// the user's shell is usually used to execute the subsystem, it is advisable for the subsystem protocol to have a
	// "magic cookie" at the beginning of the protocol transaction to distinguish it from arbitrary output generated
	// by shell initialization scripts, etc.  This spurious output from the shell may be filtered out either at the
	// server or at the client.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.5
	SubsystemRequestType = "subsystem"
	//  A pseudo-terminal can be allocated for the session by sending the following message.
	//
	// The 'encoded terminal modes' are described in Section 8.  Zero dimension parameters MUST be ignored.  The
	// character/row dimensions override the pixel dimensions (when nonzero).  Pixel dimensions refer to the drawable
	// area of the window.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.2
	PtyRequestType = "pty-req"
	// When the window (terminal) size changes on the client side, it MAY send a message to the other side to inform it
	// of the new dimensions.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.7
	WindowChangeRequestType = "window-change"
	// It is a custom request type that the Agent sends to maintain the session alive, even when no data is sent.
	KeepAliveRequestType = KeepAliveRequestTypePrefix + "@shellhub.io"
	//  When the command running at the other end terminates, the following message can be sent to return the exit
	//  status of the command. Returning the status is RECOMMENDED.
	//
	// https://www.rfc-editor.org/rfc/rfc4254#section-6.10
	ExitStatusRequest = "exit-status"
	//  The remote command may also terminate violently due to a signal. Such a condition can be indicated by the
	//  following message.  A zero 'exit_status' usually means that the command terminated successfully.
	//
	// https://datatracker.ietf.org/doc/html/rfc4254#section-6.10
	ExitSignalRequest = "exit-signal"
)

// A client may request agent forwarding for a previously-opened session using the following channel request. This
// request is sent after the channel has been opened, but before a [ShellRequestType], command or
// [SubsystemRequestType] has been executed.
//
// https://www.ietf.org/archive/id/draft-miller-ssh-agent-11.html#section-4.1
const AuthRequestOpenSSHRequest = "auth-agent-req@openssh.com"

// After a client has requested that a session have agent forwarding enabled, the server later may request a connection
// to the forwarded agent. The server does this by requesting a dedicated channel to communicate with the client's
// agent.
//
// https://www.ietf.org/archive/id/draft-miller-ssh-agent-11.html#section-4.2
const AuthRequestOpenSSHChannel = "auth-agent@openssh.com"

// DefaultSessionHandler is the default handler for session's channel.
//
// A session is a remote execution of a program.  The program may be a shell, an application, a system command, or some
// built-in subsystem. It may or may not have a tty, and may or may not involve X11 forwarding.
//
// https://www.rfc-editor.org/rfc/rfc4254#section-6
func DefaultSessionHandler() gliderssh.ChannelHandler {
	return func(_ *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		sess, _ := session.ObtainSession(ctx)

		go func() {
			// NOTICE: As [gossh.ServerConn] is shared by all channels calls, close it after a channel close block any
			// other channel invocation. To avoid it, we wait for the connection to be closed to finish the session.
			conn.Wait() //nolint:errcheck

			sess.Finish() //nolint:errcheck
		}()

		logger := log.WithFields(
			log.Fields{
				"uid":      sess.UID,
				"sshid":    sess.SSHID,
				"device":   sess.Device.UID,
				"username": sess.Target.Username,
				"ip":       sess.IPAddress,
			})

		reject := func(err error, msg string) {
			logger.WithError(err).Error(msg)

			newChan.Reject(gossh.ConnectionFailed, msg) //nolint:errcheck
		}

		logger.Info("session channel started")
		defer logger.Info("session channel done")

		client, clientReqs, err := newChan.Accept()
		if err != nil {
			reject(err, "failed to accept the channel opening")

			return
		}

		defer client.Close()

		agent, agentReqs, err := sess.AgentClient.OpenChannel(SessionChannel, nil)
		if err != nil {
			reject(err, "failed to open the 'session' channel on agent")

			return
		}

		defer agent.Close()

		seat, err := sess.NewSeat()
		if err != nil {
			reject(err, "failed to create a new set on the SSH session")

			return
		}

		go pipe(ctx, sess, client, agent, seat)

		// TODO: Add middleware to block certain types of requests.
		for {
			select {
			case <-ctx.Done():
				logger.Info("context has done")

				return
			case req, ok := <-sess.AgentGlobalReqs:
				if !ok {
					logger.Trace("global requests is closed")

					return
				}

				logger.Debugf("global request from agent: %s", req.Type)

				switch {
				// NOTICE: The Agent sends "keepalive" requests to the server to avoid the Web Socket being closed due
				// to inactivity. Through the time, the request type sent from agent to server changed its name, but
				// always keeping the prefix "keepalive". So, to maintain the retro compatibility, we check if this
				// prefix exists and perform the necessary operations.
				case strings.HasPrefix(req.Type, KeepAliveRequestTypePrefix):
					if _, err := client.SendRequest(KeepAliveRequestType, req.WantReply, req.Payload); err != nil {
						logger.Error("failed to send the keepalive request received from agent to client")

						return
					}

					if err := sess.KeepAlive(); err != nil {
						logger.WithError(err).Error("failed to send the API request to inform that the session is open")

						return
					}
				default:
					if req.WantReply {
						if err := req.Reply(false, nil); err != nil {
							logger.WithError(err).Error(err)
						}
					}
				}
			case req, ok := <-agentReqs:
				if !ok {
					logger.Trace("agent requests is closed")

					return
				}

				switch req.Type {
				case ExitStatusRequest:
					session.Event[session.Status](sess, req.Type, req.Payload, seat)
				case ExitSignalRequest:
					session.Event[session.Signal](sess, req.Type, req.Payload, seat)
				default:
					sess.Event(req.Type, req.Payload, seat)
				}

				logger.Debugf("request from agent to client: %s", req.Type)

				ok, err := client.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					logger.WithError(err).Error("failed to send the request from agent to client")

					continue
				}

				if req.WantReply {
					if err := req.Reply(ok, nil); err != nil {
						logger.WithError(err).Error(err)
					}
				}
			case req, ok := <-clientReqs:
				if !ok {
					logger.Trace("client requests is closed")

					return
				}

				switch req.Type {
				case ShellRequestType:
					if sess.Pty.Term != "" {
						if err := sess.Announce(client); err != nil {
							logger.WithError(err).Warn("failed to get the namespace announcement")
						}
					}

					sess.Event(req.Type, req.Payload, seat)
				case ExecRequestType, SubsystemRequestType:
					session.Event[session.Command](sess, req.Type, req.Payload, seat)

					sess.Type = ExecRequestType
				case PtyRequestType:
					var pty session.Pty

					if err := gossh.Unmarshal(req.Payload, &pty); err != nil {
						reject(nil, "failed to recover the session dimensions")
					}

					sess.Pty = pty

					sess.Event(req.Type, pty, seat) //nolint:errcheck
				case WindowChangeRequestType:
					var dimensions session.Dimensions

					if err := gossh.Unmarshal(req.Payload, &dimensions); err != nil {
						reject(nil, "failed to recover the session dimensions")
					}

					sess.Pty.Columns = dimensions.Columns
					sess.Pty.Rows = dimensions.Rows

					sess.Event(req.Type, dimensions, seat) //nolint:errcheck
				case AuthRequestOpenSSHRequest:
					gliderssh.SetAgentRequested(ctx)

					sess.Event(req.Type, req.Payload, seat)
					go func() {
						clientConn := ctx.Value(gliderssh.ContextKeyConn).(gossh.Conn)
						agentChannels := sess.AgentClient.HandleChannelOpen(AuthRequestOpenSSHChannel)

						for {
							newAgentChannel, ok := <-agentChannels
							if !ok {
								reject(nil, "channel for agent forwarding done")

								return
							}

							agentChannel, agentReqs, err := newAgentChannel.Accept()
							if err != nil {
								reject(nil, "failed to accept the chanel request from agent on auth request")

								return
							}

							defer agentChannel.Close()
							go gossh.DiscardRequests(agentReqs)

							clientChannel, clientReqs, err := clientConn.OpenChannel(AuthRequestOpenSSHChannel, nil)
							if err != nil {
								reject(nil, "failed to open the auth request channel from agent to client")

								return
							}

							defer clientChannel.Close()
							go gossh.DiscardRequests(clientReqs)

							hose(sess, agentChannel, clientChannel)

							logger.WithError(err).Trace("auth request channel piping done")
						}
					}()
				default:
					sess.Event(req.Type, req.Payload, seat)
				}

				logger.Debugf("request from client to agent: %s", req.Type)

				ok, err := agent.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					logger.WithError(err).Error("failed to send the request from client to agent")

					continue
				}

				if req.WantReply {
					if err := req.Reply(ok, nil); err != nil {
						logger.WithError(err).Error(err)
					}
				}
			}
		}
	}
}
