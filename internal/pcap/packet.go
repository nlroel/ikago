package pcap

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"ikago/internal/addr"
	"net"
)

// ConnPacket describes a packet and its connection
type ConnPacket struct {
	// Packet is a packet
	Packet gopacket.Packet
	// Conn is the connection of the packet
	Conn *Conn
}

// NATGuide describes simplified information about a NAT
type NATGuide struct {
	// Src is the source in NAT
	Src string
	// Proto is the protocol in NAT
	Proto gopacket.LayerType
}

// NATIndicator indicates the NAT information about a packet
type NATIndicator struct {
	src          net.Addr
	dst          net.Addr
	embSrc       net.Addr
	conn         *Conn
	hardwareAddr net.HardwareAddr
}

// EmbSrcIP returns the embedded source IP
func (indicator *NATIndicator) EmbSrcIP() net.IP {
	switch t := indicator.embSrc.(type) {
	case *net.IPAddr:
		return indicator.embSrc.(*net.IPAddr).IP
	case *net.TCPAddr:
		return indicator.embSrc.(*net.TCPAddr).IP
	case *net.UDPAddr:
		return indicator.embSrc.(*net.UDPAddr).IP
	case *addr.ICMPQueryAddr:
		return indicator.embSrc.(*addr.ICMPQueryAddr).IP
	default:
		panic(fmt.Errorf("type %T not support", t))
	}
}

// PacketIndicator indicates a packet
type PacketIndicator struct {
	linkLayer        gopacket.Layer
	networkLayer     gopacket.Layer
	transportLayer   gopacket.Layer
	icmpv4Indicator  *ICMPv4Indicator
	applicationLayer gopacket.ApplicationLayer
}

func (indicator *PacketIndicator) ipv4Layer() *layers.IPv4 {
	if indicator.NetworkLayerType() == layers.LayerTypeIPv4 {
		return indicator.networkLayer.(*layers.IPv4)
	}

	return nil
}

func (indicator *PacketIndicator) ipv6Layer() *layers.IPv6 {
	if indicator.NetworkLayerType() == layers.LayerTypeIPv6 {
		return indicator.networkLayer.(*layers.IPv6)
	}

	return nil
}

func (indicator *PacketIndicator) arpLayer() *layers.ARP {
	if indicator.NetworkLayerType() == layers.LayerTypeARP {
		return indicator.networkLayer.(*layers.ARP)
	}

	return nil
}

// LinkLayer returns the link layer
func (indicator *PacketIndicator) LinkLayer() gopacket.Layer {
	return indicator.linkLayer
}

// LinkLayerType returns the type of the link layer
func (indicator *PacketIndicator) LinkLayerType() gopacket.LayerType {
	return indicator.linkLayer.LayerType()
}

// SrcHardwareAddr returns the source hardware address
func (indicator *PacketIndicator) SrcHardwareAddr() net.HardwareAddr {
	t := indicator.LinkLayerType()
	switch t {
	case layers.LayerTypeLoopback:
		return nil
	case layers.LayerTypeEthernet:
		return indicator.linkLayer.(*layers.Ethernet).SrcMAC
	default:
		panic(fmt.Errorf("link layer type %s not support", t))
	}
}

// DstHardwareAddr returns the destination hardware address
func (indicator *PacketIndicator) DstHardwareAddr() net.HardwareAddr {
	t := indicator.LinkLayerType()
	switch t {
	case layers.LayerTypeLoopback:
		return nil
	case layers.LayerTypeEthernet:
		return indicator.linkLayer.(*layers.Ethernet).DstMAC
	default:
		panic(fmt.Errorf("link layer type %s not support", t))
	}
}

// NetworkLayer returns the network layer
func (indicator *PacketIndicator) NetworkLayer() gopacket.Layer {
	return indicator.networkLayer
}

// NetworkLayerType returns the type of the network layer
func (indicator *PacketIndicator) NetworkLayerType() gopacket.LayerType {
	return indicator.networkLayer.LayerType()
}

// TransportLayer returns the transport layer
func (indicator *PacketIndicator) TransportLayer() gopacket.Layer {
	return indicator.transportLayer
}

// TransportLayerType returns the type of the transport layer
func (indicator *PacketIndicator) TransportLayerType() gopacket.LayerType {
	return indicator.transportLayer.LayerType()
}

// TCPLayer returns the TCP layer
func (indicator *PacketIndicator) TCPLayer() *layers.TCP {
	if indicator.TransportLayerType() == layers.LayerTypeTCP {
		return indicator.transportLayer.(*layers.TCP)
	}

	return nil
}

// UDPLayer returns the UDP layer
func (indicator *PacketIndicator) UDPLayer() *layers.UDP {
	if indicator.TransportLayerType() == layers.LayerTypeUDP {
		return indicator.transportLayer.(*layers.UDP)
	}

	return nil
}

// SrcIP returns the source IP
func (indicator *PacketIndicator) SrcIP() net.IP {
	t := indicator.NetworkLayerType()
	switch t {
	case layers.LayerTypeIPv4:
		return indicator.ipv4Layer().SrcIP
	case layers.LayerTypeIPv6:
		return indicator.ipv6Layer().SrcIP
	case layers.LayerTypeARP:
		return indicator.arpLayer().SourceProtAddress
	default:
		panic(fmt.Errorf("network layer type %s not support", t))
	}
}

// DstIP returns the destination IP
func (indicator *PacketIndicator) DstIP() net.IP {
	t := indicator.NetworkLayerType()
	switch t {
	case layers.LayerTypeIPv4:
		return indicator.ipv4Layer().DstIP
	case layers.LayerTypeIPv6:
		return indicator.ipv6Layer().DstIP
	case layers.LayerTypeARP:
		return indicator.arpLayer().DstProtAddress
	default:
		panic(fmt.Errorf("network layer type %s not support", t))
	}
}

// Hop returns the TTL in IPv4 layer or hop limit in IPv6 layer
func (indicator *PacketIndicator) Hop() uint8 {
	t := indicator.NetworkLayerType()
	switch t {
	case layers.LayerTypeIPv4:
		return indicator.ipv4Layer().TTL
	case layers.LayerTypeIPv6:
		return indicator.ipv6Layer().HopLimit
	default:
		panic(fmt.Errorf("network layer type %s not support", t))
	}
}

// SrcPort returns the source port
func (indicator *PacketIndicator) SrcPort() uint16 {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP:
		return uint16(indicator.TCPLayer().SrcPort)
	case layers.LayerTypeUDP:
		return uint16(indicator.UDPLayer().SrcPort)
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// DstPort returns the destination port
func (indicator *PacketIndicator) DstPort() uint16 {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP:
		return uint16(indicator.TCPLayer().DstPort)
	case layers.LayerTypeUDP:
		return uint16(indicator.UDPLayer().DstPort)
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// NATSrc returns the source used in NAT
func (indicator *PacketIndicator) NATSrc() net.Addr {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP:
		return &net.TCPAddr{
			IP:   indicator.SrcIP(),
			Port: int(indicator.SrcPort()),
		}
	case layers.LayerTypeUDP:
		return &net.UDPAddr{
			IP:   indicator.SrcIP(),
			Port: int(indicator.SrcPort()),
		}
	case layers.LayerTypeICMPv4:
		if indicator.icmpv4Indicator.IsQuery() {
			return &addr.ICMPQueryAddr{
				IP: indicator.SrcIP(),
				Id: indicator.icmpv4Indicator.Id(),
			}
		} else {
			return indicator.icmpv4Indicator.NATSrc()
		}
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// NATDst returns the destination used in NAT
func (indicator *PacketIndicator) NATDst() net.Addr {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP:
		return &net.TCPAddr{
			IP:   indicator.DstIP(),
			Port: int(indicator.DstPort()),
		}
	case layers.LayerTypeUDP:
		return &net.UDPAddr{
			IP:   indicator.DstIP(),
			Port: int(indicator.DstPort()),
		}
	case layers.LayerTypeICMPv4:
		if indicator.icmpv4Indicator.IsQuery() {
			return &addr.ICMPQueryAddr{
				IP: indicator.DstIP(),
				Id: indicator.icmpv4Indicator.Id(),
			}
		} else {
			return indicator.icmpv4Indicator.NATDst()
		}
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// NATProto returns the protocol used in NAT
func (indicator *PacketIndicator) NATProto() gopacket.LayerType {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP, layers.LayerTypeUDP:
		return t
	case layers.LayerTypeICMPv4:
		if indicator.icmpv4Indicator.IsQuery() {
			return t
		} else {
			return indicator.icmpv4Indicator.embTransportLayerType
		}
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// Src returns the source
func (indicator *PacketIndicator) Src() net.Addr {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP:
		return &net.TCPAddr{
			IP:   indicator.SrcIP(),
			Port: int(indicator.SrcPort()),
		}
	case layers.LayerTypeUDP:
		return &net.UDPAddr{
			IP:   indicator.SrcIP(),
			Port: int(indicator.SrcPort()),
		}
	case layers.LayerTypeICMPv4:
		if indicator.icmpv4Indicator.IsQuery() {
			return &addr.ICMPQueryAddr{
				IP: indicator.SrcIP(),
				Id: indicator.icmpv4Indicator.Id(),
			}
		} else {
			return &net.IPAddr{
				IP: indicator.SrcIP(),
			}
		}
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// Dst returns the destination
func (indicator *PacketIndicator) Dst() net.Addr {
	t := indicator.TransportLayerType()
	switch t {
	case layers.LayerTypeTCP:
		return &net.TCPAddr{
			IP:   indicator.DstIP(),
			Port: int(indicator.DstPort()),
		}
	case layers.LayerTypeUDP:
		return &net.UDPAddr{
			IP:   indicator.DstIP(),
			Port: int(indicator.DstPort()),
		}
	case layers.LayerTypeICMPv4:
		if indicator.icmpv4Indicator.IsQuery() {
			return &addr.ICMPQueryAddr{
				IP: indicator.DstIP(),
				Id: indicator.icmpv4Indicator.Id(),
			}
		} else {
			return &net.IPAddr{
				IP: indicator.DstIP(),
			}
		}
	default:
		panic(fmt.Errorf("transport layer type %s not support", t))
	}
}

// Payload returns the payload, or layer contents in application layer
func (indicator *PacketIndicator) Payload() []byte {
	if indicator.applicationLayer == nil {
		return nil
	}
	return indicator.applicationLayer.LayerContents()
}

// ParsePacket parses a packet and returns a packet indicator
func ParsePacket(packet gopacket.Packet) (*PacketIndicator, error) {
	var (
		linkLayer          gopacket.Layer
		networkLayer       gopacket.Layer
		networkLayerType   gopacket.LayerType
		transportLayer     gopacket.Layer
		transportLayerType gopacket.LayerType
		icmpv4Indicator    *ICMPv4Indicator
		applicationLayer   gopacket.ApplicationLayer
	)

	// Parse packet
	linkLayer = packet.LinkLayer()
	if linkLayer == nil {
		// Guess loopback
		linkLayer = packet.Layer(layers.LayerTypeLoopback)
	}
	networkLayer = packet.NetworkLayer()
	if networkLayer == nil {
		// Guess ARP
		networkLayer = packet.Layer(layers.LayerTypeARP)
		if networkLayer == nil {
			return nil, errors.New("missing network layer")
		}

		return &PacketIndicator{
			networkLayer:     networkLayer,
			transportLayer:   nil,
			icmpv4Indicator:  nil,
			applicationLayer: nil,
		}, nil
	}
	networkLayerType = networkLayer.LayerType()
	transportLayer = packet.TransportLayer()
	if transportLayer == nil {
		// Guess ICMPv4
		transportLayer = packet.Layer(layers.LayerTypeICMPv4)
		if transportLayer == nil {
			return nil, errors.New("missing transport layer")
		}
	}
	transportLayerType = transportLayer.LayerType()
	applicationLayer = packet.ApplicationLayer()

	// Parse network layer
	switch networkLayerType {
	case layers.LayerTypeIPv4, layers.LayerTypeIPv6:
		break
	default:
		return nil, fmt.Errorf("network layer type %s not support", networkLayerType)
	}

	// Parse transport layer
	switch transportLayerType {
	case layers.LayerTypeTCP, layers.LayerTypeUDP, layers.LayerTypeARP:
		break
	case layers.LayerTypeICMPv4:
		var err error
		icmpv4Indicator, err = ParseICMPv4Layer(transportLayer.(*layers.ICMPv4))
		if err != nil {
			return nil, fmt.Errorf("parse icmpv4 layer: %w", err)
		}
	default:
		return nil, fmt.Errorf("transport layer type %s not support", transportLayerType)
	}

	return &PacketIndicator{
		linkLayer:        linkLayer,
		networkLayer:     networkLayer,
		transportLayer:   transportLayer,
		icmpv4Indicator:  icmpv4Indicator,
		applicationLayer: applicationLayer,
	}, nil
}

// ParseEmbPacket parses an embedded packet used in transferring between client and server without link layer
func ParseEmbPacket(contents []byte) (*PacketIndicator, error) {
	// Guess network layer type
	packet := gopacket.NewPacket(contents, layers.LayerTypeIPv4, gopacket.Default)
	networkLayer := packet.NetworkLayer()
	if networkLayer == nil {
		return nil, errors.New("missing network layer")
	}
	if networkLayer.LayerType() != layers.LayerTypeIPv4 {
		return nil, errors.New("network layer type not support")
	}
	ipVersion := networkLayer.(*layers.IPv4).Version
	switch ipVersion {
	case 4:
		break
	case 6:
		// Not IPv4, but IPv6
		embPacket := gopacket.NewPacket(contents, layers.LayerTypeIPv6, gopacket.Default)
		networkLayer = embPacket.NetworkLayer()
		if networkLayer == nil {
			return nil, errors.New("missing network layer")
		}
		if networkLayer.LayerType() != layers.LayerTypeIPv6 {
			return nil, errors.New("network layer type not support")
		}
	default:
		return nil, fmt.Errorf("ip version %d not support", ipVersion)
	}

	// Parse packet
	indicator, err := ParsePacket(packet)
	if err != nil {
		return nil, err
	}
	return indicator, nil
}

// ParseRawPacket parses an array of byte as a packet and returns a packet indicator
func ParseRawPacket(contents []byte) (*gopacket.Packet, error) {
	// Guess link layer type, and here we regard Ethernet layer as a link layer
	packet := gopacket.NewPacket(contents, layers.LayerTypeEthernet, gopacket.Default)
	if len(packet.Layers()) < 0 {
		return nil, errors.New("missing link layer")
	}

	linkLayer := packet.LinkLayer()
	if linkLayer == nil {
		// Guess loopback
		packet = gopacket.NewPacket(contents, layers.LayerTypeLoopback, gopacket.Default)

		linkLayer := packet.Layer(layers.LayerTypeLoopback)
		if linkLayer == nil {
			return nil, errors.New("missing link layer")
		}

		return &packet, nil
	}

	linkLayerType := linkLayer.LayerType()
	if linkLayerType != layers.LayerTypeEthernet {
		return nil, fmt.Errorf("link layer type %s not support", linkLayerType)
	}

	return &packet, nil
}

// SendTCPPacket opens a temporary TCP connection and sends a packet
func SendTCPPacket(addr string, data []byte) error {
	// Create connection
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	// Write data
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// SendUDPPacket opens a temporary UDP connection and sends a packet
func SendUDPPacket(addr string, data []byte) error {
	// Create connection
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	// Write data
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
