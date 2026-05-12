package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/mrbreakers/loadtester/internal/config"
	"github.com/mrbreakers/loadtester/internal/metrics"
	"github.com/mrbreakers/loadtester/internal/proxy"
	xproxy "golang.org/x/net/proxy"
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
		var dialer xproxy.Dialer
		dialer, err = proxy.GetDialer(*p)
		if err != nil {
			stats.AddFail()
			return
		}
		conn, err = dialer.Dial("tcp", target)
	} else {
		conn, err = net.DialTimeout("tcp", target, 5*time.Second)
	}

	if err != nil {
		stats.AddFail()
		time.Sleep(1 * time.Second) // Wait before retry
		return
	}
	defer conn.Close()

	stats.AddSuccess()
	stats.IncConn()

	// Diam sejenak sesuai interval sebelum diputuskan (reconnect)
	time.Sleep(interval)
	
	stats.DecConn()
	return
}
