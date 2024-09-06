package utmp

/*	At session start, a utmp record is constructed containing the PID,
	tty name, ID, username, remote host, source IP and time.
	Type is set to UserProcess and the record is written to UtmpxFile.
	If a record with the same ID exists, that record is overwritten;
	otherwise the record is appended to the file.  The same record is
	appended to WtmpxFile.

	At session end, Type is set to DeadProcess, the User and Host fields
	are cleared, time is updated and the record is wrtten to UtmpxFile,
	overwriting the session start record. The same record with ID and
	source IP address cleared is appended to WtmpxFile.
*/

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"os"
	"strings"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type ExitStatus struct {
	ETermination int16 // Process temination status - not used
	EExit        int16 // Process exit status - not used
}

const (
	UtmpxFile   = "/var/run/utmp"
	WtmpxFile   = "/var/log/wtmp"
	UserProcess = 0x7 // Normal process
	DeadProcess = 0x8 // Terminated process
)

// UtmpStartSession This function updates the utmp and wtmp files at the start of a user session.
func UtmpStartSession(line, user, remoteAddr string) Utmpx { //nolint:revive
	var u Utmpx

	u.Type = UserProcess
	u.Pid = int32(os.Getpid())

	// There are two versions of the utmpSetTime function
	// defined in utmp_timeval_time??.go, one for systems
	// that write the time fields as 32-bit values and one
	// for systems that write time fields as 64-bit values
	u = utmpSetTime(u)

	// remoteAddr has the form <IPv4 address>:<port>
	// or [<IPv6 address>]:<port>
	// Remove the port suffix and also the square brackets
	// if IPv6, leaving the bare IPv4 or IPv6 address
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ip": remoteAddr,
		}).Warn("wrong remoteAddr format")
	} else {
		// Parse IP address to a standard 16-byte representation
		ip := net.ParseIP(host)
		// Check whether IPv4 or IPv6
		if ip4 := ip.To4(); ip4 != nil {
			// This is a 32-bit IPv4 address to be
			// stored in the first element of u.AddrV6
			u.AddrV6[0] = binary.LittleEndian.Uint32(ip4)
		} else {
			// This is a 128-bit IPv6 address. Each 4 bytes
			// of the address is stored as a 32-bit int in
			// successive elements of u.AddrV6
			u.AddrV6[0] = binary.LittleEndian.Uint32(ip[0:4])
			u.AddrV6[1] = binary.LittleEndian.Uint32(ip[4:8])
			u.AddrV6[2] = binary.LittleEndian.Uint32(ip[8:12])
			u.AddrV6[3] = binary.LittleEndian.Uint32(ip[12:16])
		}
	}

	line = strings.TrimPrefix(line, "/dev/")
	// The index to the utmp record is the last 4 chars of line
	id := line[len(line)-4:]

	_ = copy(u.ID[:], id)
	_ = copy(u.Line[:], line)
	_ = copy(u.User[:], user)
	_ = copy(u.Host[:], host)

	updUtmp(u, id)
	updWtmp(u)

	return u
}

// UtmpEndSession this function updates the utmp and wtmp files at the end of a user session.
func UtmpEndSession(u Utmpx) { //nolint:revive
	u.Type = DeadProcess
	u.User = [32]byte{}
	u.Host = [256]byte{}
	u = utmpSetTime(u)

	updUtmp(u, string(u.ID[:]))

	u.ID = [4]byte{}
	u.AddrV6 = [4]uint32{}

	updWtmp(u)
}

// This function updates the utmp file by overwriting the record with index
// id if present; otherwise by appending the new record to the file.
func updUtmp(u Utmpx, id string) {
	file, err := os.OpenFile(
		UtmpxFile,
		os.O_RDWR|os.O_CREATE,
		0o644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": UtmpxFile,
			"err":  err,
		}).Warn("Open failed")

		return
	}

	defer file.Close()

	// Lock the file
	lk := unix.Flock_t{
		Type: int16(unix.F_WRLCK),
		Pid:  int32(os.Getpid()),
	}

	err = unix.FcntlFlock(file.Fd(), unix.F_SETLKW, &lk)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": UtmpxFile,
			"err":  err,
		}).Warn("Lock failed")

		return
	}

	var ut Utmpx

	// Read through the utmp file looking for a record with index id
	for {
		offset, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"file": UtmpxFile,
				"err":  err,
			}).Warn("Null seek failed")

			return
		}

		err = binary.Read(file, binary.LittleEndian, &ut)
		if err != nil {
			break // EOF found: no record with index id
		}

		utID := string(bytes.Trim(ut.ID[:], "\x00"))

		if utID == id {
			// Required record found, rewind to overwrite it
			if _, err = file.Seek(offset, io.SeekStart); err != nil {
				logrus.WithFields(logrus.Fields{
					"file": UtmpxFile,
					"err":  err,
				}).Warn("Back seek failed")

				return
			}

			break
		}
	}

	err = binary.Write(file, binary.LittleEndian, &u)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": UtmpxFile,
			"err":  err,
		}).Warn("Write failed")
	}
}

// This function updates the wtmp file by appending the record to the file.
func updWtmp(u Utmpx) {
	file, err := os.OpenFile(
		WtmpxFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Open failed")

		return
	}

	defer file.Close()

	lk := unix.Flock_t{
		Type: int16(unix.F_WRLCK),
		Pid:  int32(os.Getpid()),
	}

	err = unix.FcntlFlock(file.Fd(), unix.F_SETLKW, &lk)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Lock failed")
	}

	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Seek to end failed")

		return
	}

	// Check that the file is a multiple of the record size
	rem := fileSize % int64(unsafe.Sizeof(Utmpx{}))
	if rem != 0 {
		fileSize -= rem
		logrus.WithFields(logrus.Fields{
			"file":     WtmpxFile,
			"filesize": fileSize,
		}).Warn("Database size invalid, truncating")

		if err := file.Truncate(fileSize); err != nil {
			logrus.WithFields(logrus.Fields{
				"file": WtmpxFile,
				"err":  err,
			}).Warn("Database truncate failed")
		}
	}

	err = binary.Write(file, binary.LittleEndian, &u)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": WtmpxFile,
			"err":  err,
		}).Warn("Write failed")

		if err := file.Truncate(fileSize); err != nil {
			logrus.WithFields(logrus.Fields{
				"file": WtmpxFile,
				"err":  err,
			}).Warn("Database truncate failed")
		}
	}
}
