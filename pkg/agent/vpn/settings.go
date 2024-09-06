package vpn

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Settings struct {
	Address [4]byte `json:"address"`
	Mask    byte    `json:"mask"`
}

// ParseSettings read and parses the [Settings] structure from an [io.Reader].
func ParseSettings(data io.Reader) (*Settings, error) {
	body, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	settings := Settings{}
	if err = json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// String converts a [Settings] to a string representation on the format $IP/$Mask.
func (s *Settings) String() string {
	ip := net.IPv4(s.Address[0], s.Address[1], s.Address[2], s.Address[3])

	return fmt.Sprintf("%s/%d", ip.String(), s.Mask)
}
