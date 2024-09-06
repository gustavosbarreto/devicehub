package ifce

import (
	"errors"
	"fmt"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// MaximumTransmissionUnit (MTU) is the size of the largest protocol data unit (PDU) that can be communicated in a
// single network layer transaction.
const MaximumTransmissionUnit = 1024

const InterfaceName = "shb"

var ErrGenerateInterface = errors.New("failed to generate an interface")

func generateInterfaceName() (string, error) {
	const attempts = 1000

	for i := 0; i < attempts; i++ {
		name := fmt.Sprintf("%s%d", InterfaceName, i)

		if _, err := netlink.LinkByName(name); err == nil {
			continue
		}

		return name, nil
	}

	return "", ErrGenerateInterface
}

type Interface struct {
	face *water.Interface
	link netlink.Link
}

var (
	ErrInterfaceCreate        = errors.New("failed to create the interface")
	ErrInterfaceUp            = errors.New("failed to get up the interface")
	ErrInterfaceConfiguration = errors.New("failed to configure the interface")
	ErrInterfaceMTU           = errors.New("failed to configure the MTU")
)

func NewInterface(addrr string) (*Interface, error) {
	name, err := generateInterfaceName()
	if err != nil {
		return nil, err
	}

	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name:       name,
			MultiQueue: true,
		},
	})
	if err != nil {
		return nil, errors.Join(ErrInterfaceCreate, err)
	}

	addr, err := netlink.ParseAddr(addrr)
	if err != nil {
		return nil, errors.Join(ErrInterfaceConfiguration, err)
	}

	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, errors.Join(ErrInterfaceConfiguration, err)
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return nil, errors.Join(ErrInterfaceConfiguration, err)
	}

	if err := netlink.LinkSetMTU(link, MaximumTransmissionUnit); err != nil {
		return nil, errors.Join(ErrInterfaceMTU, err)
	}

	return &Interface{
		face: iface,
		link: link,
	}, nil
}

func (i *Interface) Up() error {
	if err := netlink.LinkSetUp(i.link); err != nil {
		return errors.Join(ErrInterfaceUp, err)
	}

	return nil
}

func (i *Interface) Name() string {
	return i.face.Name()
}

func (i *Interface) Close() error {
	return i.face.Close()
}

func (i *Interface) Read(buffer []byte) (int, error) {
	return i.face.Read(buffer)
}

func (i *Interface) Write(buffer []byte) (int, error) {
	return i.face.Write(buffer)
}
