// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package fakeip

import (
	"errors"
	"net/netip"
	"sync"
)

const CacheMaxSize = 1000

type Pool struct {
	gateway netip.Addr
	first   netip.Addr
	last    netip.Addr
	offset  netip.Addr
	cycle   bool
	mut     sync.Mutex
	cache   *Cache
}

func NewPool(prefix *netip.Prefix, size int) (*Pool, error) {
	hostAddr := prefix.Masked().Addr()
	gateway := hostAddr.Next()
	first := gateway.Next() // default start with 198.168.0.2
	last := unMasked(*prefix)

	if !prefix.IsValid() || !first.IsValid() || !first.Less(last) {
		return nil, errors.New("ipnet don't have valid ip")
	}

	if size <= 0 {
		size = CacheMaxSize
	}

	pool := &Pool{
		gateway: gateway,
		first:   first,
		last:    last,
		offset:  first.Prev(),
		cycle:   false,
		cache:   newCache(size),
	}

	return pool, nil
}

func (p *Pool) get(domain string) netip.Addr {
	p.offset = p.offset.Next()

	if !p.offset.Less(p.last) {
		p.cycle = true
		p.offset = p.first
	}

	_, exist := p.cache.GetByIP(p.offset)

	if p.cycle || exist {
		p.cache.RemoveByIP(p.offset)
	}

	p.cache.Add(domain, p.offset)
	return p.offset
}

func (p *Pool) GetIP(host string) netip.Addr {
	p.mut.Lock()
	defer p.mut.Unlock()

	if ip, exist := p.cache.GetByDomain(host); exist {
		return ip
	}

	ip := p.get(host)

	return ip
}

func (p *Pool) GetDomain(ip netip.Addr) (string, bool) {
	p.mut.Lock()
	defer p.mut.Unlock()

	return p.cache.GetByIP(ip)
}

func (p *Pool) Clear() {
	p.mut.Lock()
	defer p.mut.Unlock()

	p.cache.Clear()
}
