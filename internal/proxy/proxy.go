package proxy

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	xproxy "golang.org/x/net/proxy"
)

type Proxy struct {
	Host     string
	Port     string
	User     string
	Password string
}

func LoadProxies(path string) ([]Proxy, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var proxies []Proxy
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) == 4 {
			proxies = append(proxies, Proxy{
				Host:     parts[0],
				Port:     parts[1],
				User:     parts[2],
				Password: parts[3],
			})
		}
	}

	return proxies, scanner.Err()
}

func GetDialer(p Proxy) (xproxy.Dialer, error) {
	auth := &xproxy.Auth{
		User:     p.User,
		Password: p.Password,
	}
	addr := fmt.Sprintf("%s:%s", p.Host, p.Port)
	return xproxy.SOCKS5("tcp", addr, auth, xproxy.Direct)
}

func GetRandomProxy(proxies []Proxy) *Proxy {
	if len(proxies) == 0 {
		return nil
	}
	return &proxies[rand.Intn(len(proxies))]
}
