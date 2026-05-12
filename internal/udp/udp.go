package udp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/mrbreakers/loadtester/internal/config"
	"github.com/mrbreakers/loadtester/internal/metrics"
	"github.com/mrbreakers/loadtester/internal/proxy"
)

func StartWorker(ctx context.Context, cfg *config.Config, stats *metrics.Stats, proxies []proxy.Proxy) {
	target := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	payload := []byte(cfg.Payload)
	interval := time.Duration(cfg.IntervalMs) * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		default:
			dialAndSend(ctx, target, payload, interval, stats, proxies)
		}
	}
}

func dialAndSend(ctx context.Context, target string, payload []byte, interval time.Duration, stats *metrics.Stats, proxies []proxy.Proxy) {
	var conn net.Conn
	var err error

	p := proxy.GetRandomProxy(proxies)
	if p != nil {
		conn, err = proxy.DialUDP(*p, target)
	} else {
		// Only use direct if the proxy list is empty
		var addr *net.UDPAddr
		addr, err = net.ResolveUDPAddr("udp", target)
		if err == nil {
			conn, err = net.DialUDP("udp", nil, addr)
		}
	}

	if err != nil {
		stats.AddFail()
		time.Sleep(1 * time.Second)
		return
	}
	defer conn.Close()

	stats.IncConn()
	
	// Sukses Handshake
	stats.AddSuccess()
	
	// Diam sejenak sesuai interval sebelum diputuskan (reconnect)
	time.Sleep(interval)
	
	stats.DecConn()
	return
}
