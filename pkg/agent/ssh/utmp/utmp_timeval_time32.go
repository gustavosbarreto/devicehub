//go:build !arm64
// +build !arm64

package utmp

/*	The Utmpx struct is derived from the Linux definition (man 5 utmpx).

	On all Linux systems (both 32-bit and 64-bit) except for
	arm64/aarch64, a utmp record is 384 bytes in length and the
	Session, Tv.Sec and Tv.Usec fields are all 32 bits in length.

	On arm64/aarch64, a utmp record is 400 bytes in length and the
	Session, Tv.Sec and Tv.Usec fields are all 64 bits in length.

	There are two versions of this file, one for arm64/aarch64 and
	one for all other architectures.
*/

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type TimeVal struct {
	Sec  int32 // Seconds since epoch
	Usec int32 // Microseconds
}

type Utmpx struct {
	Type     int16      // UserProcess or DeadProcess
	Padding  [2]byte    // Padding to align rest of struct
	Pid      int32      // PID of the ShellHub agent
	Line     [32]byte   // tty associated with the process
	ID       [4]byte    // Index, last 4 characters of Line
	User     [32]byte   // Username
	Host     [256]byte  // Source IP address
	Exit     ExitStatus // Exit status - not used
	Session  int32      // Session ID - not used
	Tv       TimeVal    // Time entry was made
	AddrV6   [4]uint32  // Source IP address. IPv4 in AddrV6[0]
	Reserved [20]byte   // Not used
}

// This function writes the current time into the utmp record.
func utmpSetTime(u Utmpx) Utmpx {
	a := unix.Timeval{}
	if err := unix.Gettimeofday(&a); err != nil {
		logrus.Warn(err)
	}

	u.Tv.Sec, u.Tv.Usec = int32(a.Sec), int32(a.Usec)

	return u
}
