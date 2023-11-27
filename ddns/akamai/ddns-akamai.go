package akamai

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/MeteorsLiu/CloudflareDDNS/ddns/ip"
)

const (
	defaultTimeout    = 15 * time.Second
	defaultChinaQuery = "http://whatismyip.akamai.com/"
)

type AkamaiDDNS struct {
	query         string
	timeout       time.Duration
	parentContext context.Context
}

func NewAkamaiDDNS(ctx context.Context, opts ...ip.Options) ip.Getter {
	c := &AkamaiDDNS{
		query:         defaultChinaQuery,
		timeout:       defaultTimeout,
		parentContext: ctx,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (a *AkamaiDDNS) SetQueryURL(u string) {
	a.query = u
}
func (a *AkamaiDDNS) SetTimeout(t time.Duration) {
	a.timeout = t
}

func (a *AkamaiDDNS) GetIP() (string, error) {
	_ctx, cancel := context.WithTimeout(a.parentContext, a.timeout)
	defer cancel()
	req, err := http.NewRequest("GET", a.query, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req.WithContext(_ctx))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body), nil
}
