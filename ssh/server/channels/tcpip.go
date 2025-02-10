package channels

import (
	"io"
	"net"
	"strconv"
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// DefaultDirectTCPIPHandler is the channel's handler for direct-tcpip channels like "local port forwarding" and "dynamic
// application-level port forwarding".
func DefaultDirectTCPIPHandler(server *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	sess, _ := session.ObtainSession(ctx)
	go func() {
		// NOTICE: As [gossh.ServerConn] is shared by all channels calls, close it after a channel close block any
		// other channel involkation. To avoid it, we wait for the connection be closed to finish the sesison.
		conn.Wait() //nolint:errcheck

		sess.Finish() //nolint:errcheck
	}()

	log.WithFields(log.Fields{
		"username": sess.Target.Username,
		"sshid":    sess.Target.Data,
	}).Trace("handling direct-tcpip channel")

	type channelData struct {
		DestAddr   string `json:"dest_addr"`
		DestPort   uint32 `json:"dest_port"`
		OriginAddr string `json:"origin_addr"`
		OriginPort uint32 `json:"origin_port"`
	}

	data := new(channelData)
	if err := gossh.Unmarshal(newChan.ExtraData(), data); err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed to parse forward data: "+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed to parse forward data")

		return
	}

	if server.LocalPortForwardingCallback == nil || !server.LocalPortForwardingCallback(ctx, data.DestAddr, data.DestPort) {
		newChan.Reject(gossh.Prohibited, "port forwarding is disabled") //nolint:errcheck
		log.WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Info("port forwarding is disabled")

		return
	}

	seat, err := sess.NewSeat()
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed to create the seat"+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed to create the seat")

		return
	}

	sess.Event(DirectTCPIPChannel, data, seat) //nolint:errcheck

	dest := net.JoinHostPort(data.DestAddr, strconv.FormatInt(int64(data.DestPort), 10))

	// NOTE: Certain SSH connections may not necessitate a dedicated handler, such as an SSH handler.
	// In such instances, a new connection to the agent is generated and saved in the metadata for
	// subsequent use.
	// An illustrative scenario is when the SSH connection is initiated with the "-N" flag.
	connection := sess.AgentClient

	agent, err := connection.Dial("tcp", dest)
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed dialing the agent to host and port: "+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed dialing the agent to host and port")

		return
	}

	defer agent.Close()

	client, reqs, err := newChan.Accept()
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed accepting the channel: "+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed accepting the channel")

		return
	}

	defer client.Close()

	go gossh.DiscardRequests(reqs)

	log.WithFields(log.Fields{
		"username":    sess.Target.Username,
		"sshid":       sess.Target.Data,
		"origin_port": data.OriginAddr,
		"origin_addr": data.OriginPort,
		"dest_port":   data.DestPort,
		"dest_addr":   data.DestAddr,
	}).Info("piping data between client and agent")

	wg := new(sync.WaitGroup)

	// TODO: control the running state of these goroutines.
	wg.Add(1)
	go func() {
		defer wg.Done()

		log.WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Trace("copying data from client to agent")

		if _, err := io.Copy(client, agent); err != nil && err != io.EOF {
			log.WithError(err).Error("failed to copy data from agent to client")

			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.WithFields(log.Fields{
			"username":    sess.Target.Username,
			"sshid":       sess.Target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Trace("copying data from agent to client")

		if _, err := io.Copy(agent, client); err != nil && err != io.EOF {
			log.WithError(err).Error("failed to copy data from client to agent")

			return
		}
	}()

	wg.Wait()

	log.WithFields(log.Fields{
		"username":    sess.Target.Username,
		"sshid":       sess.Target.Data,
		"origin_port": data.OriginAddr,
		"origin_addr": data.OriginPort,
		"dest_port":   data.DestPort,
		"dest_addr":   data.DestAddr,
	}).Trace("handling direct-tcpip finished")
}
