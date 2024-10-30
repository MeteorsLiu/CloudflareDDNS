package ddns

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/MeteorsLiu/CloudflareDDNS/ddns/ip"
	"github.com/cloudflare/cloudflare-go"
)

type DDNS struct {
	ctx      context.Context
	api      *cloudflare.API
	getter   ip.Getter
	waitTime time.Duration
	doStop   context.CancelFunc
	FQDN     cloudflare.DNSRecord
	logger   *log.Logger
	verbose  bool
	RecordID string
	shell    string
}

func splitFQDN(domain string) string {
	DomainSlice := strings.Split(domain, ".")
	SliceLength := len(DomainSlice)
	if SliceLength > 2 {
		return strings.Join(DomainSlice[SliceLength-2:], ".")
	} else {
		return domain
	}
}

func NewDDNS(
	ctx context.Context, stop context.CancelFunc,
	getter ip.Getter,
	verbose bool, wait int,
	key, email, domain,
	shell string,
) *DDNS {
	api, err := cloudflare.New(key, email)
	if err != nil {
		log.Fatal(err)
	}
	id, err := api.ZoneIDByName(splitFQDN(domain))
	if err != nil {
		log.Fatal(err)
	}
	FQDN := cloudflare.DNSRecord{Name: domain, ZoneID: id}
	d := &DDNS{
		api:      api,
		FQDN:     FQDN,
		ctx:      ctx,
		doStop:   stop,
		getter:   getter,
		verbose:  verbose,
		waitTime: time.Duration(wait) * time.Second,
		logger:   log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags),
	}
	return d
}

func (d *DDNS) Println(s string) {
	if d.verbose {
		d.logger.Output(2, s)
	}
}

func (d *DDNS) Printf(f string, s ...any) {
	if d.verbose {
		d.logger.Output(2, fmt.Sprintf(f, s...))
	}
}

func (d *DDNS) execShell() {
	if d.shell != "" {
		err := exec.Command("bash", "-c", d.shell).Run()
		if err != nil {
			d.Println("Execute Hook Failed")
		} else {
			d.Println("Execute Hook Success")
		}
	}
}

func (d *DDNS) getCFIP() string {
	recs, err := d.api.DNSRecords(d.ctx, d.FQDN.ZoneID, d.FQDN)
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
func (d *DDNS) updateCFIP(ip string) {
	//There is no need to modify FQDN.
	//We just need a copy of FQDN and modify the copy.
	d.FQDN.Content = ip
	err := d.api.UpdateDNSRecord(d.ctx, d.FQDN.ZoneID, d.RecordID, d.FQDN)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *DDNS) Run(ch chan os.Signal, once bool) {
	ip := d.getCFIP()

	update := func() {
		current, err := d.getter.GetIP()
		if err != nil {
			d.Printf("GetIP: %v", err)
			return
		}
		if ip != current {
			ip = current
			d.updateCFIP(current)
			d.execShell()
			d.Printf("IP has been changed to %s", current)
		}
	}
	update()

	if once {
		return
	}

	ticker := time.NewTicker(d.waitTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			update()
		case <-ch:
			d.doStop()
			return
		}
	}
}
