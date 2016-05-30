package hsts_test

import (
	"fmt"
	"net/http"

	"github.com/tam7t/hsts"
)

func Example() {
	hosts := hsts.NewMemStorage()

	tr := &hsts.Transport{
		Transport: http.DefaultTransport,
		Storage:   hosts,
	}
	client := &http.Client{Transport: tr}

	// http://github.com will redirect to https, then set hsts header
	fmt.Println("GET: http://github.com")
	_, err := client.Get("http://github.com")
	if err != nil {
		fmt.Println(err)
	}

	// subsequent request to http://github.com will be intercepted and forced over
	// https
	fmt.Println("GET: http://github.com")
	_, err = client.Get("http://github.com")
	if err != nil {
		fmt.Println(err)
	}
}
