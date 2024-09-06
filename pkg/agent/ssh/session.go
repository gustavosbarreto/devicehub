package ssh

import (
	"fmt"

	gliderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

// Type is the type of SSH session.
type Type string

const (
	// SessionTypeShell is the session's type returned when the SSH client requests a shell.
	SessionTypeShell Type = "shell"
	// SessionTypeHeredoc is the session's type returned when the SSH client requests a command execution with a heredoc.
	// "heredoc" is a format that does not require a TTY, but attaches the client input to the command's stdin.
	// It is used to execute a sequence of commands in a single SSH connection without the need to open a shell.
	SessionTypeHeredoc Type = "heredoc"
	// SessionTypeExec is the session's type returned when the SSH client requests a command execution.
	SessionTypeExec Type = "exec"
	// SessionTypeSubsystem is the session's type returned when the SSH client requests a subsystem.
	SessionTypeSubsystem Type = "subsystem"
	// SessionTypeUnknown is the session's type returned when the SSH client requests an unknown session type.
	SessionTypeUnknown Type = "unknown"
)

// GetSessionType returns the session's type based on the SSH client session.
func GetSessionType(session gliderssh.Session) (Type, error) {
	_, _, isPty := session.Pty()

	requestType, ok := session.Context().Value("request_type").(string)
	if !ok {
		return SessionTypeUnknown, fmt.Errorf("failed to get request type from session context")
	}

	switch {
	case isPty && requestType == RequestTypeShell:
		return SessionTypeShell, nil
	case !isPty && requestType == RequestTypeShell:
		return SessionTypeHeredoc, nil
	case requestType == RequestTypeExec:
		return SessionTypeExec, nil
	case requestType == RequestTypeSubsystem:
		return SessionTypeSubsystem, nil
	default:
		return SessionTypeUnknown, nil
	}
}

func (s *Server) sessionHandler(session gliderssh.Session) {
	log.Info("New session request")

	go s.startKeepAliveLoop(session)

	sessionType, err := GetSessionType(session)
	if err != nil {
		log.Error(err)

		return
	}

	log.WithField("type", sessionType).Info("Request type got")

	switch sessionType {
	case SessionTypeShell:
		s.mode.Shell(session) //nolint:errcheck
	case SessionTypeHeredoc:
		s.mode.Heredoc(session) //nolint:errcheck
	default:
		s.mode.Exec(session) //nolint:errcheck
	}

	log.Info("Session ended")
}
