package hsts

import (
	"testing"
	"time"
)

func TestDomain_Check(t *testing.T) {
	now := time.Now().Unix()
	fiveHours := int64(5 * 60 * 60)
	tenHours := int64(fiveHours * 2)

	tests := []struct {
		name     string
		host     string
		domain   Domain
		expected bool
	}{
		{
			name: "subdomain enabled",
			host: "z.a.example.com",
			domain: Domain{
				Host:       "a.example.com",
				Subdomains: true,
				Created:    now,
				MaxAge:     tenHours,
			},
			expected: true,
		},
		{
			name: "subdomain disabled",
			host: "z.b.example.com",
			domain: Domain{
				Host:       "a.example.com",
				Subdomains: true,
				Created:    now,
				MaxAge:     tenHours,
			},
			expected: false,
		},
		{
			name: "exact match",
			host: "b.example.com",
			domain: Domain{
				Host:       "b.example.com",
				Subdomains: true,
				Created:    now,
				MaxAge:     tenHours,
			},
			expected: true,
		},
		{
			name: "exact match with ports",
			host: "b.example.com:8443",
			domain: Domain{
				Host:       "b.example.com:8443",
				Subdomains: true,
				Created:    now,
				MaxAge:     tenHours,
			},
			expected: true,
		},
		{
			name: "missmatch ports",
			host: "b.example.com:8443",
			domain: Domain{
				Host:       "b.example.com:8442",
				Subdomains: true,
				Created:    now,
				MaxAge:     tenHours,
			},
			expected: false,
		},
		{
			name: "expired",
			host: "z.example.com",
			domain: Domain{
				Host:       "z.example.com",
				Subdomains: true,
				Created:    now - tenHours,
				MaxAge:     fiveHours,
			},
			expected: false,
		},
		{
			name: "permanent",
			host: "z.example.com",
			domain: Domain{
				Host:       "z.example.com",
				Subdomains: true,
				Permanent:  true,
				Created:    0,
				MaxAge:     0,
			},
			expected: true,
		},
	}

	for _, test := range tests {
		out := test.domain.Check(test.host)
		if out != test.expected {
			t.Logf("host: %s", test.host)
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			t.Fatalf("test case failed: %s", test.name)
		}
	}
}
