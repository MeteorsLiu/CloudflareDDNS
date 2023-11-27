package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MeteorsLiu/CloudflareDDNS/ddns"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/akamai"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/china"
	"github.com/MeteorsLiu/CloudflareDDNS/ddns/ip"
)

var (
	cfkey    = flag.String("key", "", "CF Key")
	cfemail  = flag.String("email", "", "CF Email")
	waitTime = flag.Int("time", 15, "Wait Time (Second).Default 15 Second")
	hook     = flag.String("hook", "", "Bash shell to execute when ip has been changed")
	cfdomain = flag.String("domain", "", "DDNS Domain")
	query    = flag.String("query", "", "Custom IP Query URL")
	mode     = flag.String("mode", "akamai", "Akamai mode: Akamai URL to get current IP, chian mode: use ipip.net")
	timeout  = flag.Int("timeout", 0, "IP Query Timeout")
)

func assertNotEmpty() {
	if *cfkey == "" {
		log.Fatal("no cloudflare key")
	}
	if *cfemail == "" {
		log.Fatal("no cloudflare email")
	}
	if *cfdomain == "" {
		log.Fatal("no ddns domain")
	}
}

func parseGetter(ctx context.Context) ip.Getter {
	var opts []ip.Options
	if *query != "" {
		opts = append(opts, ip.WithQueryURL(*query))
	}
	if *timeout > 0 {
		opts = append(opts, ip.WithTimeout(time.Duration(*timeout)*time.Second))
	}
	switch *mode {
	case "akamai":
		return akamai.NewAkamaiDDNS(ctx, opts...)
	case "china":
		return china.NewChinaDDNS(ctx, opts...)
	}
	panic("no ip query mode")
}

func main() {
	flag.Parse()
	assertNotEmpty()
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	DDNS := ddns.NewDDNS(
		ctx, cancel,
		parseGetter(ctx),
		*waitTime, *cfkey,
		*cfemail, *cfdomain,
		*hook,
	)
	DDNS.Run(sigCh)
}
