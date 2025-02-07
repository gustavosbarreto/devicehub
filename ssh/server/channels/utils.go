package channels

import (
	"io"
	"sync"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Recorder struct {
	queue   chan string
	channel gossh.Channel
}

func NewRecorder(channel gossh.Channel, sess *session.Session, camera *session.Camera, seat int) (io.WriteCloser, error) {
	// NOTE: The queue's size is a random number.
	queue := make(chan string, 100)

	go func() {
		for {
			msg, ok := <-queue
			if !ok {
				log.WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("recorder queue is closed")

				return
			}

			if err := camera.WriteFrame(&models.SessionRecorded{ //nolint:errcheck
				UID:       sess.UID,
				Seat:      seat,
				Namespace: sess.Lookup["domain"],
				Message:   msg,
				Width:     int(sess.Pty.Columns),
				Height:    int(sess.Pty.Rows),
			}); err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("failed to send the session frame to record")

					// NOTE: When a frame isn't sent correctly, we stop the writing loop, only reading from the queue,
					// and discarding the messages to avoid stuck the go routine.
				break
			}
		}

		for {
			// NOTE: Reads the queue and discards the data to avoid stuck the go routine.
			if _, ok := <-queue; !ok {
				log.WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
					Warning("recorder queue is closed")

				return
			}
		}
	}()

	return &Recorder{
		queue:   queue,
		channel: channel,
	}, nil
}

// record enqueues a session frame to be recorded. If the queue is closed, nothing is done.
func (c *Recorder) record(msg string) {
	select {
	case c.queue <- msg:
	default:
		log.Trace("the message couldn't sent to the record queue")
	}
}

func (c *Recorder) Write(data []byte) (int, error) {
	read, err := c.channel.Write(data)
	if err != nil {
		return read, err
	}

	c.record(string(data))

	return read, nil
}

func (c *Recorder) Close() error {
	close(c.queue)

	return c.channel.CloseWrite()
}

// pipe pipes data between client and agent, and vice versa, recording each frame when ShellHub instance are Cloud or
// Enterprise.
func pipe(ctx gliderssh.Context, sess *session.Session, client gossh.Channel, agent gossh.Channel, seat int) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	c := io.MultiReader(client, client.Stderr())
	a := io.MultiReader(agent, agent.Stderr())

	go func() {
		defer wg.Done()

		if envs.IsEnterprise() || envs.IsCloud() {
			recordURL := ctx.Value("RECORD_URL").(string)
			if recordURL == "" {
				log.WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID, "record_url": recordURL}).
					Warning("failed to start session's record because the record URL is empty")

				goto normal
			}

			camera, err := sess.Record(ctx, recordURL, seat)
			if err != nil {
				goto normal
			}

			recorder, err := NewRecorder(client, sess, camera, seat)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID, "record_url": recordURL}).
					Warning("failed to connect to session record endpoint")

				goto normal
			}

			defer recorder.Close() //nolint:errcheck

			if _, err := io.Copy(recorder, a); err != nil && err != io.EOF {
				log.WithError(err).Error("failed on coping data from client to agent")
			}

			return
		}

		// NOTE: "normal" labels indicate the default way of copying data between clients and the agent without recording.
		// Their idea was, if something goes wrong with the recording flow, the session will continue, even without the
		// recording.
	normal:
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, a); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer func() {
			// NOTE: When request is [ExecRequestType] and agent's version is less than v0.9.2, we should close the agent
			// connection to avoid it be hanged after data flow ends.
			if ver, err := semver.NewVersion(sess.Device.Info.Version); ver != nil && err == nil {
				// NOTE: We indicate here v0.9.3, but it is not included due the assertion `less than`.
				if ver.LessThan(semver.MustParse("v0.9.3")) && sess.Type == ExecRequestType {
					agent.Close()
				} else {
					agent.CloseWrite() //nolint:errcheck
				}
			} else {
				agent.CloseWrite() //nolint:errcheck
			}
		}()

		if _, err := io.Copy(agent, c); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}

// hose is a generic version of [pipe] function without the record capability.
func hose(sess *session.Session, agent gossh.Channel, client gossh.Channel) {
	defer log.
		WithFields(log.Fields{"session": sess.UID, "sshid": sess.SSHID}).
		Trace("data pipe between client and agent has done")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	a := io.MultiReader(agent, agent.Stderr())
	c := io.MultiReader(client, client.Stderr())

	go func() {
		defer wg.Done()
		defer agent.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(agent, c); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from client to agent")
		}

		log.Trace("agent channel data copy done")
	}()

	go func() {
		defer wg.Done()
		defer client.CloseWrite() //nolint:errcheck

		if _, err := io.Copy(client, a); err != nil && err != io.EOF {
			log.WithError(err).Error("failed on coping data from agent to client")
		}

		log.Trace("client channel data copy done")
	}()

	wg.Wait()
}
