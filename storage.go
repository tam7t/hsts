package hsts

import (
	"strings"
	"sync"
)

// MemStorage is threadsafe hsts host storage backed by an in-memory map
type MemStorage struct {
	domains map[string]Domain
	mutex   sync.Mutex
}

// NewMemStorage initializes hsts in-memory datastructure
func NewMemStorage() *MemStorage {
	m := &MemStorage{}
	m.domains = make(map[string]Domain)
	return m
}

// Contains whether storage has host
func (hs *MemStorage) Contains(h string) bool {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	d, ok := hs.domains[h]
	if ok {
		// exact match!
		return d.Check(h)
	}

	// is h a subdomain of an hsts domain, walk the domain to see if it is a sub
	// sub ... sub domain of a domain that has the `includeSubdomains` rule
	l := len(h)
	originalHost := h
	for l > 0 {
		i := strings.Index(h, ".")
		if i > 0 {
			h = h[i+1:]
			d, ok := hs.domains[h]
			if ok {
				return d.Check(originalHost)
			}
			l = len(h)
		} else {
			break
		}
	}

	return false
}

// Add a domain to hsts storage
func (hs *MemStorage) Add(d *Domain) {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	if hs.domains == nil {
		hs.domains = make(map[string]Domain)
	}

	if d.MaxAge == 0 && !d.Permanent {
		check, ok := hs.domains[d.Host]
		if ok {
			if !check.Permanent {
				delete(hs.domains, d.Host)
			}
		}
	} else {
		hs.domains[d.Host] = *d
	}
}
