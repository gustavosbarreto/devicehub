package host

import (
	"io"
	"os"
	"os/exec"
	"syscall"

	creackpty "github.com/creack/pty"
	glidderssh "github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

func openPty(c *exec.Cmd) (*os.File, *os.File, error) {
	ptmx, tty, err := creackpty.Open()
	if err != nil {
		return nil, nil, err
	}
	defer tty.Close()

	if c.Stdout == nil {
		c.Stdout = tty
	}
	if c.Stderr == nil {
		c.Stderr = tty
	}
	if c.Stdin == nil {
		c.Stdin = tty
	}

	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}

	c.SysProcAttr.Setsid = true
	c.SysProcAttr.Setctty = true

	if err := c.Start(); err != nil {
		_ = ptmx.Close()

		return nil, nil, err
	}

	return ptmx, tty, err
}

func startPty(c *exec.Cmd, out io.ReadWriter, winCh <-chan glidderssh.Window) (*os.File, error) {
	f, tty, err := openPty(c)
	if err != nil {
		return nil, err
	}

	go func() {
		for win := range winCh {
			_ = creackpty.Setsize(f, &creackpty.Winsize{uint16(win.Height), uint16(win.Width), 0, 0})
		}
	}()

	go func() {
		_, err := io.Copy(out, f)
		if err != nil {
			log.Warn(err)
		}
	}()

	go func() {
		_, err := io.Copy(f, out)
		if err != nil {
			log.Warn(err)
		}
	}()

	return tty, nil
}

// initPty initializes and configures a new pseudo-terminal (PTY) for the provided command. Returns a pty and its corresponding tty.
func initPty(c *exec.Cmd, sess io.ReadWriter, winCh <-chan glidderssh.Window) (*os.File, *os.File, error) {
	pty, tty, err := creackpty.Open()
	if err != nil {
		return nil, nil, err
	}

	if c.Stdout == nil {
		c.Stdout = tty
	}
	if c.Stderr == nil {
		c.Stderr = tty
	}
	if c.Stdin == nil {
		c.Stdin = tty
	}

	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}

	c.SysProcAttr.Setsid = true
	c.SysProcAttr.Setctty = true

	// listen for window size changes from the SSH client and update the PTY's dimensions.
	go func() {
		for win := range winCh {
			_ = creackpty.Setsize(pty, &creackpty.Winsize{uint16(win.Height), uint16(win.Width), 0, 0})
		}
	}()

	// forward the command's output to the SSH session
	go func() {
		_, err := io.Copy(sess, pty)
		if err != nil {
			log.Warn(err)
		}
	}()

	// forward the input from the SSH session to the command
	go func() {
		_, err := io.Copy(pty, sess)
		if err != nil {
			log.Warn(err)
		}
	}()

	return pty, tty, nil
}
