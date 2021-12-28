package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

var (
	cfkey        = flag.String("key", "", "CF Key")
	cfemail      = flag.String("email", "", "CF Email")
	waitTime     = flag.Int("time", 1, "Wait Time (Second).Default 1 Second")
	hook         = flag.String("hook", "", "Bash shell to execute when ip has been changed")
	ctx          = context.Background()
	cfdomain     = flag.String("domain", "", "DDNS Domain")
	DefaultQuery = "http://whatismyip.akamai.com/"
)

func SplitFQDN(domain string) string {
	DomainSlice := strings.Split(domain, ".")
	SliceLength := len(DomainSlice)
	if SliceLength > 2 {
		return strings.Join(DomainSlice[SliceLength-2:], ".")
	} else {
		return domain
	}
}

func ExecShell(cmd string) {
	err := exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		log.Printf("Execute Hook Failed")
	} else {
		log.Printf("Execute Hook Success")
	}

}

type DDNS struct {
	api      *cloudflare.API
	FQDN     cloudflare.DNSRecord
	RecordID string
}

func (d *DDNS) GetCFIP() string {
	recs, err := d.api.DNSRecords(ctx, d.FQDN.ZoneID, d.FQDN)
	if err != nil {
		log.Fatal(err)
	}
	//Ignore other records
	r := recs[0]
	//Init RecordID
	//RecordID is Fixed.If you don't remove the record.
	if d.RecordID == "" {
		d.RecordID = r.ID
	}

	return r.Content

}
func (d *DDNS) UpdateCFIP(ip string) {
	//There is no need to modify FQDN.
	//We just need a copy of FQDN and modify the copy.
	d.FQDN.Content = ip
	err := d.api.UpdateDNSRecord(ctx, d.FQDN.ZoneID, d.RecordID, d.FQDN)
	if err != nil {
		log.Fatal(err)
	}

}

func DoGETTimeout(URL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body), nil

}

func main() {
	flag.Parse()

	api, err := cloudflare.New(*cfkey, *cfemail)
	if err != nil {
		log.Fatal(err)
	}

	id, err := api.ZoneIDByName(SplitFQDN(*cfdomain))
	if err != nil {
		log.Fatal(err)
	}
	FQDN := cloudflare.DNSRecord{Name: *cfdomain, ZoneID: id}
	sigCh := make(chan os.Signal, 1)
	ticker := time.NewTicker(time.Duration(*waitTime) * time.Second)
	DDNS := &DDNS{api: api, FQDN: FQDN}

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer ticker.Stop()

	var ip string
	IP := DDNS.GetCFIP()
	for {
		select {
		case <-ticker.C:
			ip, err = DoGETTimeout(DefaultQuery)
			if err != nil {
				log.Printf("%v", err)
				continue
			}
			if ip != IP {
				log.Printf("IP has been changed to %s", ip)
				IP = ip
				if *hook != "" {
					ExecShell(*hook)
				}
				DDNS.UpdateCFIP(ip)
			}
		case <-sigCh:
			log.Println("Goroutine Exit")
			return
		}
	}

}
