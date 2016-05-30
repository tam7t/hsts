package hsts

import (
	"net/http"
	"reflect"
	"testing"
)

func TestTransport_RoundTrip(t *testing.T) {
	memStorage := &MemStorage{}
	hstsTransport := &Transport{
		Storage: memStorage,
	}

	expRedirect, _ := redirect("https://example.com/index.html", newRequest("http://example.com/index.html"))

	tests := []struct {
		name     string
		server   func(req *http.Request) (*http.Response, error)
		req      *http.Request
		expected *http.Response
	}{
		{
			name: "normal",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme": {req.URL.Scheme},
					},
				}, nil
			},
			req: newRequest("http://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme": {"http"},
				},
			},
		},
		{
			name: "set hsts header under http",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme":                    {req.URL.Scheme},
						"Strict-Transport-Security": {"max-age=10000"},
					},
				}, nil
			},
			req: newRequest("http://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme":                    {"http"},
					"Strict-Transport-Security": {"max-age=10000"},
				},
			},
		},
		{
			name: "ensure http STS header was ignored",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme": {req.URL.Scheme},
					},
				}, nil
			},
			req: newRequest("http://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme": {"http"},
				},
			},
		},
		{
			name: "set hsts header",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme":                    {req.URL.Scheme},
						"Strict-Transport-Security": {"max-age=10000"},
					},
				}, nil
			},
			req: newRequest("https://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme":                    {"https"},
					"Strict-Transport-Security": {"max-age=10000"},
				},
			},
		},
		{
			name: "verify hsts header was set",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme": {req.URL.Scheme},
					},
				}, nil
			},
			req:      newRequest("http://example.com/index.html"),
			expected: expRedirect,
		},
		{
			name: "remove hsts",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme":                    {req.URL.Scheme},
						"Strict-Transport-Security": {"max-age=0"},
					},
				}, nil
			},
			req: newRequest("https://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme":                    {"https"},
					"Strict-Transport-Security": {"max-age=0"},
				},
			},
		},
		{
			name: "verify hsts header is no longer set",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme": {req.URL.Scheme},
					},
				}, nil
			},
			req: newRequest("http://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme": {"http"},
				},
			},
		},
		{
			name: "set hsts header for max-age verification",
			server: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Header: map[string][]string{
						"scheme":                    {req.URL.Scheme},
						"Strict-Transport-Security": {"max-age=10000"},
					},
				}, nil
			},
			req: newRequest("https://example.com/index.html"),
			expected: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"scheme":                    {"https"},
					"Strict-Transport-Security": {"max-age=10000"},
				},
			},
		},
	}

	for _, test := range tests {
		// inject the stub transport
		hstsTransport.Transport = &stubTrip{f: test.server}

		// perform the call
		out, err := hstsTransport.RoundTrip(test.req)
		if err != nil {
			t.Logf("got err:%v", err)
			t.Fatalf("test case failed: %s", test.name)
		}

		if !equalResponses(out, test.expected) {
			t.Logf("got:%v", out)
			t.Logf("want:%v", test.expected)
			t.Fatalf("test case failed: %s", test.name)
		}
	}

	if memStorage.domains["example.com"].MaxAge != 10000 {
		t.Fatalf("max-age not set properly")
	}
}

func newRequest(url string) *http.Request {
	r, _ := http.NewRequest("GET", url, nil)
	return r
}

type stubTrip struct {
	f func(req *http.Request) (*http.Response, error)
}

func (s *stubTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	return s.f(req)
}

// equalResponses is a custom comparator for http.Response since the Body
// parameter does not compare well with reflect.DeepEqual
func equalResponses(a, b *http.Response) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if a.Status != b.Status {
		return false
	}

	if a.Proto != b.Proto {
		return false
	}

	if !reflect.DeepEqual(a.Header, b.Header) {
		return false
	}

	if !reflect.DeepEqual(a.Request, b.Request) {
		return false
	}

	return true
}
