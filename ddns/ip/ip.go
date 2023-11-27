package ip

import "time"

type Options func(Getter)

type Getter interface {
	GetIP() (string, error)
	SetQueryURL(string)
	SetTimeout(time.Duration)
}

func WithQueryURL(u string) Options {
	return func(g Getter) {
		g.SetQueryURL(u)
	}
}

func WithTimeout(t time.Duration) Options {
	return func(g Getter) {
		g.SetTimeout(t)
	}
}
