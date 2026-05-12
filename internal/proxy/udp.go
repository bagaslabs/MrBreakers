package proxy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Socks5UDPConn handles UDP packets over SOCKS5.
type Socks5UDPConn struct {
	tcpConn   net.Conn
	udpConn   *net.UDPConn
	relayAddr *net.UDPAddr
	target    string
}

func (c *Socks5UDPConn) Close() error {
	if c.tcpConn != nil {
		c.tcpConn.Close()
	}
	if c.udpConn != nil {
		return c.udpConn.Close()
	}
	return nil
}

func (c *Socks5UDPConn) Write(b []byte) (int, error) {
	// SOCKS5 UDP Header: [RSV 2B][FRAG 1B][ATYP 1B][DST.ADDR VAR][DST.PORT 2B][DATA]
	host, portStr, err := net.SplitHostPort(c.target)
	if err != nil {
		return 0, err
	}
	var port uint16
	fmt.Sscanf(portStr, "%d", &port)

	ip := net.ParseIP(host)
	var header []byte
	if ip == nil {
		// Domain name
		header = make([]byte, 7+len(host))
		header[0], header[1], header[2], header[3] = 0, 0, 0, 3 // RSV, RSV, FRAG, ATYP=Domain
		header[4] = byte(len(host))
		copy(header[5:], host)
		binary.BigEndian.PutUint16(header[5+len(host):], port)
	} else if ip4 := ip.To4(); ip4 != nil {
		// IPv4
		header = make([]byte, 10)
		header[0], header[1], header[2], header[3] = 0, 0, 0, 1 // RSV, RSV, FRAG, ATYP=IPv4
		copy(header[4:], ip4)
		binary.BigEndian.PutUint16(header[8:], port)
	} else {
		// IPv6
		header = make([]byte, 22)
		header[0], header[1], header[2], header[3] = 0, 0, 0, 4 // RSV, RSV, FRAG, ATYP=IPv6
		copy(header[4:], ip.To16())
		binary.BigEndian.PutUint16(header[20:], port)
	}

	packet := append(header, b...)
	return c.udpConn.WriteToUDP(packet, c.relayAddr)
}

func (c *Socks5UDPConn) Read(b []byte) (int, error) {
	buf := make([]byte, 2048)
	n, _, err := c.udpConn.ReadFromUDP(buf)
	if err != nil {
		return 0, err
	}

	if n < 10 {
		return 0, errors.New("packet too short")
	}

	// [RSV 2][FRAG 1][ATYP 1][ADDR VAR][PORT 2][DATA]
	atyp := buf[3]
	offset := 4
	switch atyp {
	case 1: // IPv4
		offset += 4 + 2
	case 3: // Domain
		offset += int(buf[4]) + 1 + 2
	case 4: // IPv6
		offset += 16 + 2
	default:
		return 0, errors.New("unknown ATYP")
	}

	if n < offset {
		return 0, errors.New("packet too short after header")
	}

	copy(b, buf[offset:n])
	return n - offset, nil
}

func (c *Socks5UDPConn) SetReadDeadline(t time.Time) error {
	return c.udpConn.SetReadDeadline(t)
}

func (c *Socks5UDPConn) SetDeadline(t time.Time) error {
	return c.udpConn.SetDeadline(t)
}

func (c *Socks5UDPConn) SetWriteDeadline(t time.Time) error {
	return c.udpConn.SetWriteDeadline(t)
}

func (c *Socks5UDPConn) LocalAddr() net.Addr {
	return c.udpConn.LocalAddr()
}

func (c *Socks5UDPConn) RemoteAddr() net.Addr {
	return c.relayAddr
}

func DialUDP(p Proxy, target string) (net.Conn, error) {
	proxyAddr := fmt.Sprintf("%s:%s", p.Host, p.Port)
	// Use longer timeout for slow proxies
	tcpConn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	// 1. Greeting
	if p.User != "" {
		tcpConn.Write([]byte{0x05, 0x01, 0x02})
	} else {
		tcpConn.Write([]byte{0x05, 0x01, 0x00})
	}

	resp := make([]byte, 2)
	if _, err := io.ReadFull(tcpConn, resp); err != nil {
		tcpConn.Close()
		return nil, err
	}

	if resp[0] != 0x05 {
		tcpConn.Close()
		return nil, errors.New("invalid socks version")
	}

	// 2. Auth
	if resp[1] == 0x02 {
		authMsg := append([]byte{0x01, byte(len(p.User))}, []byte(p.User)...)
		authMsg = append(authMsg, byte(len(p.Password)))
		authMsg = append(authMsg, []byte(p.Password)...)
		tcpConn.Write(authMsg)
		if _, err := io.ReadFull(tcpConn, resp); err != nil {
			tcpConn.Close()
			return nil, err
		}
		if resp[1] != 0x00 {
			tcpConn.Close()
			return nil, errors.New("auth failed")
		}
	} else if resp[1] != 0x00 {
		tcpConn.Close()
		return nil, errors.New("auth method not supported")
	}

	// 3. UDP Associate
	// Send 0.0.0.0:0 as client address
	tcpConn.Write([]byte{0x05, 0x03, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	
	respHeader := make([]byte, 4)
	if _, err := io.ReadFull(tcpConn, respHeader); err != nil {
		tcpConn.Close()
		return nil, err
	}

	if respHeader[1] != 0x00 {
		tcpConn.Close()
		return nil, fmt.Errorf("udp associate failed: %d", respHeader[1])
	}

	// Get relay address
	var relayIP net.IP
	switch respHeader[3] {
	case 1: // IPv4
		addrBuf := make([]byte, 4)
		io.ReadFull(tcpConn, addrBuf)
		relayIP = net.IP(addrBuf)
	case 3: // Domain name
		lenBuf := make([]byte, 1)
		io.ReadFull(tcpConn, lenBuf)
		domainBuf := make([]byte, int(lenBuf[0]))
		io.ReadFull(tcpConn, domainBuf)
		addr, err := net.ResolveIPAddr("ip", string(domainBuf))
		if err == nil {
			relayIP = addr.IP
		}
	case 4: // IPv6
		addrBuf := make([]byte, 16)
		io.ReadFull(tcpConn, addrBuf)
		relayIP = net.IP(addrBuf)
	}

	portBuf := make([]byte, 2)
	io.ReadFull(tcpConn, portBuf)
	relayPort := binary.BigEndian.Uint16(portBuf)

	// CRITICAL: If relay IP is 0.0.0.0 or private, use the proxy's IP address.
	if relayIP.IsUnspecified() || isPrivateIP(relayIP) {
		hostOnly := p.Host
		if strings.Contains(hostOnly, ":") {
			hostOnly, _, _ = net.SplitHostPort(hostOnly)
		}
		addr, err := net.ResolveIPAddr("ip", hostOnly)
		if err == nil {
			relayIP = addr.IP
		} else {
			relayIP = net.ParseIP(hostOnly)
		}
	}

	relayAddr := &net.UDPAddr{
		IP:   relayIP,
		Port: int(relayPort),
	}

	// 4. Setup local UDP port
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		tcpConn.Close()
		return nil, err
	}

	return &Socks5UDPConn{
		tcpConn:   tcpConn,
		udpConn:   udpConn,
		relayAddr: relayAddr,
		target:    target,
	}, nil
}

func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return false
}
