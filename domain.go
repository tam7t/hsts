package hsts

import (
	"strings"
	"time"
)

// Domain holds hsts information about a domain
type Domain struct {
	Host       string
	Subdomains bool
	Permanent  bool
	Created    int64
	MaxAge     int64
}

// Check returns whether the given host should use TLS based on the Domain rule
func (d *Domain) Check(h string) bool {
	hostMatch := strings.HasSuffix(h, d.Host)
	if !hostMatch {
		return false
	}
	if !d.Subdomains && h != d.Host {
		// it is a subdomain, but subdomains arnt enabled
		return false
	}

	if d.Permanent {
		return true
	}

	if time.Now().Unix() > time.Unix(d.Created, 0).Add(time.Duration(d.MaxAge)*time.Second).Unix() {
		return false
	}
	return true
}
