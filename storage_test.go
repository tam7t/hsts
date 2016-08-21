package hsts

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMemStorage_Contains(t *testing.T) {
	m := NewMemStorage()
	m.Add(&Domain{
		Host:       "example.org",
		Subdomains: false,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     100,
	})
	m.Add(&Domain{
		Host:       "a.example.org",
		Subdomains: true,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     100,
	})

	m.Add(&Domain{
		Host:       "a.example.com",
		Subdomains: true,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     100,
	})
	m.Add(&Domain{
		Host:       "b.example.com",
		Subdomains: false,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     100,
	})

	done := make(chan bool)

	var orgErr error
	go func() {
		orgErr = orgTest(m, t)
		// try to make a data race
		m.Add(&Domain{
			Host:       "example.org",
			Subdomains: false,
			Permanent:  false,
			Created:    time.Now().Unix(),
			MaxAge:     100,
		})
		done <- true
	}()

	var comErr error
	go func() {
		comErr = comTest(m, t)
		// try to make a data race
		m.Add(&Domain{
			Host:       "a.example.com",
			Subdomains: true,
			Permanent:  false,
			Created:    time.Now().Unix(),
			MaxAge:     100,
		})
		done <- true
	}()

	// wait for tests to finish
	<-done
	<-done

	if orgErr != nil {
		t.Fatal(orgErr)
	}

	if comErr != nil {
		t.Fatal(comErr)
	}
}

func orgTest(m *MemStorage, t *testing.T) error {
	tests := []struct {
		name     string
		host     string
		expected bool
	}{
		{
			name:     "root match org",
			host:     "example.org",
			expected: true,
		},
		{
			name:     "subdomain match org",
			host:     "a.example.org",
			expected: true,
		},
		{
			name:     "subdomain miss-match org",
			host:     "b.example.org",
			expected: false,
		},
	}

	for _, test := range tests {
		out := m.Contains(test.host)
		if out != test.expected {
			t.Logf("host: %s", test.host)
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			return fmt.Errorf("test case failed: %s", test.name)
		}
	}

	return nil
}

func comTest(m *MemStorage, t *testing.T) error {
	tests := []struct {
		name     string
		host     string
		expected bool
	}{
		{
			name:     "subdomain enabled",
			host:     "z.a.example.com",
			expected: true,
		},
		{
			name:     "sub-subdomain",
			host:     "z.y.a.example.com",
			expected: true,
		},
		{
			name:     "subdomain disabled",
			host:     "z.b.example.com",
			expected: false,
		},
		{
			name:     "exact match",
			host:     "b.example.com",
			expected: true,
		},
		{
			name:     "complete missmatch",
			host:     "z.example.com",
			expected: false,
		},
	}

	for _, test := range tests {
		out := m.Contains(test.host)
		if out != test.expected {
			t.Logf("host: %s", test.host)
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			return fmt.Errorf("test case failed: %s", test.name)
		}
	}

	return nil
}

func TestMemStorage_Add(t *testing.T) {
	m := &MemStorage{}

	permanentDomain := Domain{
		Host:       "example.org",
		Subdomains: false,
		Permanent:  true,
		Created:    time.Now().Unix(),
		MaxAge:     0,
	}

	normalDomain := Domain{
		Host:       "a.example.org",
		Subdomains: false,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     100,
	}

	removeNormalDomain := Domain{
		Host:       "a.example.org",
		Subdomains: false,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     0,
	}

	removePermanetDomain := Domain{
		Host:       "example.org",
		Subdomains: false,
		Permanent:  false,
		Created:    time.Now().Unix(),
		MaxAge:     0,
	}

	// permanent
	m.Add(&permanentDomain)

	expected := map[string]Domain{
		"example.org": permanentDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after permanent")
	}

	// normal
	m.Add(&normalDomain)

	expected = map[string]Domain{
		"example.org":   permanentDomain,
		"a.example.org": normalDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after adding normal")
	}

	// remove normal
	m.Add(&removeNormalDomain)

	expected = map[string]Domain{
		"example.org": permanentDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after removing normal")
	}

	// attempt to remove the permanent
	m.Add(&removePermanetDomain)

	expected = map[string]Domain{
		"example.org": permanentDomain,
	}

	if !reflect.DeepEqual(m.domains, expected) {
		t.Logf("want:%v", expected)
		t.Logf("got:%v", m.domains)
		t.Fatal("Add failed after attempting to remove the permanent domain")
	}
}
