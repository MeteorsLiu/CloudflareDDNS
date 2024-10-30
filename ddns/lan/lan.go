package lan

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/MeteorsLiu/CloudflareDDNS/ddns/ip"
)

type Lan struct {
	devName       string
	parentContext context.Context
}

func NewLan(ctx context.Context, opts ...ip.Options) ip.Getter {
	c := &Lan{
		parentContext: ctx,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Lan) SetQueryURL(u string) {
	c.devName = u
}
func (c *Lan) SetTimeout(t time.Duration) {}

func getInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}
	if addrs, err = ief.Addrs(); err != nil { // get addresses
		return
	}
	for _, addr := range addrs { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		return "", fmt.Errorf("interface %s don't have an ipv4 address", interfaceName)
	}
	return ipv4Addr.String(), nil
}

func (c *Lan) GetIP() (string, error) {
	return getInterfaceIpv4Addr(c.devName)
}
