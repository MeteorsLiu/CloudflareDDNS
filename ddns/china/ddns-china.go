package china

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/MeteorsLiu/CloudflareDDNS/ddns/ip"
)

const (
	defaultTimeout    = 15 * time.Second
	defaultChinaQuery = "http://myip.ipip.net/ip"
)

type IPIP struct {
	IP string `json:"ip"`
}

type ChinaDDNS struct {
	query         string
	timeout       time.Duration
	parentContext context.Context
}

func NewChinaDDNS(ctx context.Context, opts ...ip.Options) ip.Getter {
	c := &ChinaDDNS{
		query:         defaultChinaQuery,
		timeout:       defaultTimeout,
		parentContext: ctx,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *ChinaDDNS) SetQueryURL(u string) {
	c.query = u
}
func (c *ChinaDDNS) SetTimeout(t time.Duration) {
	c.timeout = t
}

func (c *ChinaDDNS) GetIP() (string, error) {
	_ctx, cancel := context.WithTimeout(c.parentContext, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(_ctx, "GET", c.query, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var ret IPIP
	json.Unmarshal(body, &ret)
	if ret.IP == "" {
		return "", errors.New("cannot get the outer ip")
	}
	return ret.IP, nil
}
