package main

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/MamdMehrabi/DNS-Checker/dns"
	hostname "github.com/MamdMehrabi/DNS-Checker/extract"
	"github.com/MamdMehrabi/DNS-Checker/http"
	"github.com/MamdMehrabi/DNS-Checker/resolveDns"
)

// DNS list
var commonDNSServers = []string{
	"8.8.8.8",         // Google DNS
	"8.8.4.4",         // Google DNS
	"1.1.1.1",         // Cloudflare DNS
	"1.0.0.1",         // Cloudflare DNS
	"78.157.42.100",    // Electro DNS
	"78.157.42.101",   // Electro DNS
	"10.202.10.10",    // RadarGame DNS
	"10.202.10.11",    // RadarGame DNS
}

func main() {
	var website, mode string

	fmt.Print("Enter Target Site: (example google.com): ")
	fmt.Scanln(&website)

	fmt.Print("Do you want to test from the DNS list above? (y/n): ")
	fmt.Scanln(&mode)

	if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
		website = "https://" + website
	}

	if strings.ToLower(mode) == "y" || strings.ToLower(mode) == "yes" {
		fmt.Printf("\n🔍 در حال بررسی سایت %s با لیست DNS سرورها...\n\n", website)
		checkDNSList(website, commonDNSServers)
	} else {
		var dnsServer string
		fmt.Print("لطفا DNS سرور را وارد کنید (مثال: 8.8.8.8 یا 1.1.1.1): ")
		fmt.Scanln(&dnsServer)

		fmt.Printf("\n🔍 در حال بررسی سایت %s با DNS %s...\n\n", website, dnsServer)

		if dns.CheckDNS(website, dnsServer) {
			fmt.Println("✅ DNS resolution موفقیت‌آمیز بود")

			if http.CheckHTTP(website, dnsServer) {
				fmt.Println("✅ اتصال HTTP موفقیت‌آمیز بود")
				fmt.Println("\n🎉 سایت با این DNS قابل دسترسی است!")
			} else {
				fmt.Println("⚠️ سایت از طریق DNS قابل resolve است اما اتصال HTTP برقرار نشد")
				fmt.Println("\nℹ️ ممکن است سایت فیلتر شده باشد یا مشکل دیگری وجود داشته باشد")
			}
		} else {
			fmt.Println("❌ DNS resolution ناموفق بود")
			fmt.Println("\n⚠️ این DNS نمی‌تواند این سایت را resolve کند")
		}
	}
}

func checkDNSList(website string, dnsServers []string) {
	hostname := hostname.ExtractHostname(website)

	type DNSResult struct {
		DNS     string
		Success bool
		IPs     []net.IPAddr
		Error   string
		HTTPOk  bool
	}

	results := make([]DNSResult, len(dnsServers))
	var wg sync.WaitGroup
	var mu sync.Mutex

	fmt.Printf("در حال بررسی %d DNS سرور برای %s...\n\n", len(dnsServers), hostname)

	for i, dnsServer := range dnsServers {
		wg.Add(1)
		go func(index int, dns string) {
			defer wg.Done()

			result := DNSResult{
				DNS: dns,
			}

			ips, err := resolveDns.ResolveWithDNS(hostname, dns)
			if err != nil {
				result.Success = false
				result.Error = err.Error()
			} else if len(ips) > 0 {
				result.Success = true
				result.IPs = ips

				result.HTTPOk = http.CheckHTTP(website, dns)
			} else {
				result.Success = false
				result.Error = "no IPs found"
			}

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, dnsServer)
	}

	wg.Wait()

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("نتایج بررسی DNS برای %s:\n\n", hostname)

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			fmt.Printf("✅ %s\n", result.DNS)
			fmt.Printf("   IP ها: ")
			for i, ip := range result.IPs {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(ip.String())
			}
			fmt.Println()

			if result.HTTPOk {
				fmt.Printf("   HTTP: ✅ قابل دسترسی\n")
			} else {
				fmt.Printf("   HTTP: ⚠️  قابل دسترسی نیست\n")
			}
		} else {
			fmt.Printf("❌ %s\n", result.DNS)
			fmt.Printf("   خطا: %s\n", result.Error)
		}
		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("\n📊 خلاصه: %d از %d DNS سرور موفق بودند\n", successCount, len(dnsServers))

	if successCount > 0 {
		fmt.Println("\n✅ DNS سرورهای موفق:")
		for _, result := range results {
			if result.Success {
				status := "❌ HTTP"
				if result.HTTPOk {
					status = "✅ HTTP"
				}
				fmt.Printf("   - %s (%s)\n", result.DNS, status)
			}
		}
	}
}