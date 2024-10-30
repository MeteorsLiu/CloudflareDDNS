package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MeteorsLiu/CloudflareDDNS/ddns"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/akamai"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/china"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/ip"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/lan"
)

var (
	verbose  = flag.Bool("v", true, "Verbose mode")
	cfkey    = flag.String("key", "", "CF Key")
	cfemail  = flag.String("email", "", "CF Email")
	waitTime = flag.Int("time", 15, "Wait Time (Second).Default 15 Second")
	hook     = flag.String("hook", "", "Bash shell to execute when ip has been changed")
	cfdomain = flag.String("domain", "", "DDNS Domain")
	query    = flag.String("query", "", "Custom IP Query URL")
	mode     = flag.String("mode", "akamai", "Akamai mode: Akamai URL to get current IP, chian mode: use ipip.net, LAN Mode: use to get LAN ip, run once")
	timeout  = flag.Int("timeout", 0, "IP Query Timeout")
	dev      = flag.String("dev", "", "Get specific network gate ip in LAN mode")
)

func assertNotEmpty(m string) {
	if *cfkey == "" {
		log.Fatal("no cloudflare key")
	}
	if *cfemail == "" {
		log.Fatal("no cloudflare email")
	}
	if *cfdomain == "" {
		log.Fatal("no ddns domain")
	}
	if m == "lan" && *dev == "" {
		log.Fatal("Lan mode requres dev name")
	}
}

func parseGetter(ctx context.Context, m string) (ip.Getter, bool) {
	var opts []ip.Options
	if *query != "" {
		opts = append(opts, ip.WithQueryURL(*query))
	}
	if *dev != "" {
		opts = append(opts, ip.WithQueryURL(*dev))
	}
	if *timeout > 0 {
		opts = append(opts, ip.WithTimeout(time.Duration(*timeout)*time.Second))
	}
	switch m {
	case "akamai":
		return akamai.NewAkamaiDDNS(ctx, opts...), false
	case "china":
		return china.NewChinaDDNS(ctx, opts...), false
	case "lan":
		return lan.NewLan(ctx, opts...), true
	}
	panic("no ip query mode")
}

func main() {
	flag.Parse()
	m := strings.ToLower(*mode)
	assertNotEmpty(m)
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	getter, once := parseGetter(ctx, m)
	DDNS := ddns.NewDDNS(
		ctx, cancel,
		getter,
		*verbose, *waitTime,
		*cfkey, *cfemail,
		*cfdomain, *hook,
	)
	DDNS.Run(sigCh, once)
}
