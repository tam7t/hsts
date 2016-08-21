package hsts

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Storage is threadsafe hsts storage interface
type Storage interface {
	Contains(host string) bool
	Add(d *Domain)
}

// Transport contains the hosts that must be https
type Transport struct {
	Transport http.RoundTripper
	Storage   Storage
}

// RoundTrip satisfies the RoundTripper interface for implementing HSTS
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// is request insecure?
	if req.URL.Scheme == "http" {
		// should it be secure?
		if t.Storage.Contains(req.URL.Host) {
			// "RoundTrip should not modify the request"
			// so we will fake a redirect response.
			// chrome dev tools show 307 redirects when enforcing HSTS
			newURL := *req.URL
			newURL.Scheme = "https"
			return redirect(newURL.String(), req)
		}
	}

	// service normal requests
	resp, err := t.Transport.RoundTrip(req)

	// "RoundTrip should not attempt to interpret the response"
	// https://golang.org/src/net/http/client.go?s=#L92
	// But pretty sure they don't mean us.
	if resp != nil && req.URL.Scheme == "https" {
		hsts := resp.Header.Get("Strict-Transport-Security")
		if hsts != "" {
			sub, max := parseHeader(hsts)
			t.Storage.Add(&Domain{
				Host:       req.URL.Host,
				Subdomains: sub,
				Permanent:  false,
				Created:    time.Now().Unix(),
				MaxAge:     max,
			})
		}
	}

	return resp, err
}

func parseHeader(h string) (subdomains bool, maxAge int64) {
	subdomains = strings.Contains(h, "includeSubDomains")
	maxAge = 10

	for _, v := range strings.Split(h, ";") {
		i := strings.LastIndex(v, "max-age=")
		if i > -1 {
			ma, err := strconv.Atoi(v[i+8:])
			if err == nil {
				maxAge = int64(ma)
			}
		}
	}

	return subdomains, maxAge
}

const redirectFormatString string = `HTTP/1.1 307 Moved Permanently
Location: %s
Non-Authoritative-Reason: HSTS

`

func redirect(url string, req *http.Request) (*http.Response, error) {
	buf := &bytes.Buffer{}
	_, err := buf.WriteString(fmt.Sprintf(redirectFormatString, url))
	if err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(buf), req)
}
