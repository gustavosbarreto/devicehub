package packets

import (
	"encoding/binary"
	"net"
)

const (
	Multicast = "multicast"
	Unicast   = "unicast"
)

// PacketMaxSize defines the max size of each IP packet received. This value is related to the interface MTU.
const PacketMaxSize = 1024

// IPv4 packet.
//
// https://en.wikipedia.org/wiki/IPv4#Header
type Packet struct {
	Version        uint8   // Version of the IP protocol (typically 4 for IPv4)
	IHL            uint8   // Internet Header Length (in 32-bit words)
	TOS            uint8   // Type of Service
	TotalLength    uint16  // Total length of the packet (header + data)
	Identification uint16  // Identification field for packet fragmentation
	Flags          uint8   // Flags for fragmentation control
	FragmentOffset uint16  // Fragment offset
	TTL            uint8   // Time To Live
	Protocol       uint8   // Protocol (e.g., TCP, UDP)
	HeaderChecksum uint16  // Checksum of the header
	Source         [4]byte // Source IP address
	Destination    [4]byte // Destination IP address
	Options        []byte  // Optional fields (if any)
	Payload        []byte  // Payload (data)
}

// Protocols stores a map of [byte] for string matching the protocol's name.
//
// https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
var Protocols = map[byte]string{
	0x00: "HOPOPT",
	0x01: "ICMP",
	0x02: "IGMP",
	0x03: "GGP",
	0x04: "IP-in-IP",
	0x05: "ST",
	0x06: "TCP",
	0x07: "CBT",
	0x08: "EGP",
	0x09: "IGP",
	0x0A: "BBN-RCC-MON",
	0x0B: "NVP-II",
	0x0C: "PUP",
	0x0D: "ARGUS",
	0x0E: "EMCON",
	0x0F: "XNET",
	0x10: "CHAOS",
	0x11: "UDP",
	0x12: "MUX",
	0x13: "DCN-MEAS",
	0x14: "HMP",
	0x15: "PRM",
	0x16: "XNS-IDP",
	0x17: "TRUNK-1",
	0x18: "TRUNK-2",
	0x19: "LEAF-1",
	0x1A: "LEAF-2",
	0x1B: "RDP",
	0x1C: "IRTP",
	0x1D: "ISO-TP4",
	0x1E: "NETBLT",
	0x1F: "MFE-NSP",
	0x20: "MERIT-INP",
	0x21: "DCCP",
	0x22: "3PC",
	0x23: "IDPR",
	0x24: "XTP",
	0x25: "DDP",
	0x26: "IDPR-CMTP",
	0x27: "TP++",
	0x28: "IL",
	0x29: "IPv6",
	0x2A: "SDRP",
	0x2B: "IPv6-Route",
	0x2C: "IPv6-Frag",
	0x2D: "IDRP",
	0x2E: "RSVP",
	0x2F: "GRE",
	0x30: "DSR",
	0x31: "BNA",
	0x32: "ESP",
	0x33: "AH",
	0x34: "I-NLSP",
	0x35: "SwIPe",
	0x36: "NARP",
	0x37: "MOBILE",
	0x38: "TLSP",
	0x39: "SKIP",
	0x3A: "IPv6-ICMP",
	0x3B: "IPv6-NoNxt",
	0x3C: "IPv6-Opts",
	0x3D: "Any host internal protocol",
	0x3E: "CFTP",
	0x3F: "Any local network",
	0x40: "SAT-EXPAK",
	0x41: "KRYPTOLAN",
	0x42: "RVD",
	0x43: "IPPC",
	0x44: "Any distributed file system",
	0x45: "SAT-MON",
	0x46: "VISA",
	0x47: "IPCU",
	0x48: "CPNX",
	0x49: "CPHB",
	0x4A: "WSN",
	0x4B: "PVP",
	0x4C: "BR-SAT-MON",
	0x4D: "SUN-ND",
	0x4E: "WB-MON",
	0x4F: "WB-EXPAK",
	0x50: "ISO-IP",
	0x51: "VMTP",
	0x52: "SECURE-VMTP",
	0x53: "VINES",
	0x54: "TTP/IPTM",
	0x55: "NSFNET-IGP",
	0x56: "DGP",
	0x57: "TCF",
	0x58: "EIGRP",
	0x59: "OSPF",
	0x5A: "Sprite-RPC",
	0x5B: "LARP",
	0x5C: "MTP",
	0x5D: "AX.25",
	0x5E: "OS",
	0x5F: "MICP",
	0x60: "SCC-SP",
	0x61: "ETHERIP",
	0x62: "ENCAP",
	0x63: "Any private encryption scheme",
	0x64: "GMTP",
	0x65: "IFMP",
	0x66: "PNNI",
	0x67: "PIM",
	0x68: "ARIS",
	0x69: "SCPS",
	0x6A: "QNX",
	0x6B: "A/N",
	0x6C: "IPComp",
	0x6D: "SNP",
	0x6E: "Compaq-Peer",
	0x6F: "IPX-in-IP",
	0x70: "VRRP",
	0x71: "PGM",
	0x72: "Any 0-hop protocol",
	0x73: "L2TP",
	0x74: "DDX",
	0x75: "IATP",
	0x76: "STP",
	0x77: "SRP",
	0x78: "UTI",
	0x79: "SMP",
	0x7A: "SM",
	0x7B: "PTP",
	0x7C: "IS-IS over IPv4",
	0x7D: "FIRE",
	0x7E: "CRTP",
	0x7F: "CRUDP",
	0x80: "SSCOPMCE",
	0x81: "IPLT",
	0x82: "SPS",
	0x83: "PIPE",
	0x84: "SCTP",
	0x85: "FC",
	0x86: "RSVP-E2E-IGNORE",
	0x87: "Mobility Header",
	0x88: "UDPLite",
	0x89: "MPLS-in-IP",
	0x8A: "manet",
	0x8B: "HIP",
	0x8C: "Shim6",
	0x8D: "WESP",
	0x8E: "ROHC",
	0x8F: "Ethernet",
	0x90: "AGGFRAG",
	0x91: "NSH",
	0x92: "Unassigned",
	0xFD: "Use for experimentation and testing",
	0xFE: "Use for experimentation and testing",
	0xFF: "Reserved",
}

func Protocol(buffer []byte) string {
	if len(buffer) < 9 {
		return "Invalid"
	}

	if protocol, ok := Protocols[buffer[9]]; ok {
		return protocol
	}

	return "Unassigned"
}

func Source(buffer []byte) [4]byte {
	return [4]byte{buffer[12], buffer[13], buffer[14], buffer[15]}
}

func Destination(buffer []byte) [4]byte {
	return [4]byte{buffer[16], buffer[17], buffer[18], buffer[19]}
}

func Length(buffer []byte) int {
	return int(binary.BigEndian.Uint16(buffer[2:4]))
}

func TimeToLive(buffer []byte) int {
	return int(binary.BigEndian.Uint16(buffer[8:9]))
}

// TODO: try don't use the [net.IPv4] or use it everytime.
func IsMulticast(buffer []byte) bool {
	return net.IPv4(buffer[16], buffer[17], buffer[18], buffer[19]).IsMulticast()
}

// TODO: try don't use the [net.IPv4] or use it everytime.
func IsUnicast(buffer []byte) bool {
	return net.IPv4(buffer[16], buffer[17], buffer[18], buffer[19]).IsGlobalUnicast()
}
