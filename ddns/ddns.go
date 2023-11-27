package ddns

import (
	"context"
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
	getter ip.Getter, wait int,
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
		getter:   getter,
		doStop:   stop,
		waitTime: time.Duration(wait) * time.Second,
	}
	return d
}

func (d *DDNS) execShell() {
	if d.shell != "" {
		err := exec.Command("bash", "-c", d.shell).Run()
		if err != nil {
			log.Printf("Execute Hook Failed")
		} else {
			log.Printf("Execute Hook Success")
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

func (d *DDNS) Run(ch chan os.Signal) {
	ip := d.getCFIP()
	ticker := time.NewTicker(d.waitTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			current, err := d.getter.GetIP()
			if err != nil {
				log.Println("GetIP: ", err)
				continue
			}
			if ip != current {
				ip = current
				d.updateCFIP(current)
				d.execShell()
			}
		case <-ch:
			d.doStop()
			return
		}
	}
}
