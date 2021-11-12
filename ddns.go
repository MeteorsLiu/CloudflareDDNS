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
	"sync"
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
	ipChan := make(chan string)
	quit := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	ticker := time.NewTicker(time.Duration(*waitTime) * time.Second)
	DDNS := &DDNS{api: api, FQDN: FQDN}
	var wg sync.WaitGroup

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer ticker.Stop()

	//Make sure that Goroutine exits successfully
	wg.Add(1)
	go func() {
		defer wg.Done()
		//Avoid declaring vars repeatedly.
		var ip string
		IP := DDNS.GetCFIP()
		for {
			select {
			case <-ticker.C:
				//Block Until the other side is ready
				ip = <-ipChan
				if ip != IP {
					log.Printf("IP has been changed to %s", ip)
					IP = ip
					if *hook != "" {
						ExecShell(*hook)
					}
					DDNS.UpdateCFIP(ip)
				}
			case <-quit:
				log.Println("Goroutine Exit")
				return
			}
		}

	}()

	CFunc := func() {
		defer close(ipChan)
		defer close(quit)
		var resp *http.Response
		for {
			select {
			case <-ticker.C:
				resp, err = http.Get(DefaultQuery)
				if err != nil {
					log.Fatal(err)
				}
				body, _ := io.ReadAll(resp.Body)
				ipChan <- string(body)
			case <-sigCh:
				resp.Body.Close()
				log.Println("Main Exit")
				return
			}
		}

	}
	CFunc()
	wg.Wait()

}
