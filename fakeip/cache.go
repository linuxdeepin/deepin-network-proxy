// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package fakeip

import (
	"net/netip"

	"github.com/golang/groupcache/lru"
)

type Cache struct {
	domainCache *lru.Cache
	ipCache     *lru.Cache
}

func newCache(size int) *Cache {
	return &Cache{
		domainCache: lru.New(size),
		ipCache:     lru.New(size),
	}
}

func (c *Cache) Add(domain string, ip netip.Addr) {
	c.domainCache.Add(domain, ip)
	c.ipCache.Add(ip, domain)
}

func (c *Cache) RemoveByIP(ip netip.Addr) {
	val, ok := c.ipCache.Get(ip)
	if ok {
		c.ipCache.Remove(ip)
		c.domainCache.Remove(val)
	}
}

func (c *Cache) GetByDomain(domain string) (netip.Addr, bool) {
	val, ok := c.domainCache.Get(domain)
	if !ok {
		return netip.Addr{}, false
	}

	// hot
	c.ipCache.Get(val)

	ip, ok := val.(netip.Addr)

	return ip, ok
}

func (c *Cache) GetByIP(ip netip.Addr) (string, bool) {
	val, ok := c.ipCache.Get(ip)
	if !ok {
		return "", false
	}

	// hot
	c.domainCache.Get(val)

	domain, ok := val.(string)

	return domain, ok
}

func (c *Cache) Clear() {
	c.ipCache.Clear()
	c.domainCache.Clear()
}
