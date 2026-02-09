package http

import (
	"fmt"
	"net"
	"context"
	"net/http"
	"time"
)

func CheckHTTP(website string, dnsServer string) bool {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			resolver := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					return dialer.DialContext(ctx, "udp", dnsServer+":53")
				},
			}

			ips, err := resolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("no IPs found for %s", host)
			}

			return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", website, nil)
	if err != nil {
		fmt.Printf("خطا در ایجاد درخواست: %v\n", err)
		return false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("خطا در اتصال HTTP: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	fmt.Printf("وضعیت HTTP: %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	if resp.StatusCode < 400 {
		return true
	}

	return false
}