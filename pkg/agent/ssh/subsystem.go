package ssh

import (
	gliderssh "github.com/gliderlabs/ssh"
)

// sftpSubsystemHandler handles the SFTP subsystem session.
func (s *Server) sftpSubsystemHandler(session gliderssh.Session) {
	go s.startKeepAliveLoop(session)

	s.mode.SFTP(session) //nolint:errcheck
}
