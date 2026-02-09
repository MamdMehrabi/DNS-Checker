package dns

import (
	"context"
	"fmt"
	"net"
	"time"
	hostname "github.com/MamdMehrabi/DNS-Checker/extract"
)

func CheckDNS(website string, dnsServer string) bool {
	hostname := hostname.ExtractHostname(website)

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, "udp", dnsServer+":53")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		fmt.Printf("خطا در DNS resolution: %v\n", err)
		return false
	}

	if len(ips) > 0 {
		fmt.Printf("آدرس‌های IP پیدا شده برای %s:\n", hostname)
		for _, ip := range ips {
			fmt.Printf("  - %s\n", ip.String())
		}
		return true
	}

	return false
}
