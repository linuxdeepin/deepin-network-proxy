// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package proxy

import (
	"net"
	"sync"

	"github.com/golang/groupcache/lru"
)

const cacheMaxSize = 1000

type fakeIPCache struct {
	domainCache *lru.Cache
	ipCache     *lru.Cache

	mut sync.Mutex
}

func newFakeIPCache() *fakeIPCache {
	f := new(fakeIPCache)
	f.domainCache = lru.New(cacheMaxSize)
	f.ipCache = lru.New(cacheMaxSize)

	return f
}

func (f *fakeIPCache) Add(domain string, ip net.IP) {
	f.mut.Lock()

	ipUint := ipToUint(ip)

	f.domainCache.Add(domain, ipUint)
	f.ipCache.Add(ipUint, domain)

	logger.Debugf("ipuint: %d", ipUint)

	f.mut.Unlock()
}

func (f *fakeIPCache) GetByDomain(domain string) (net.IP, bool) {
	f.mut.Lock()
	defer f.mut.Unlock()

	ipUintI, ok := f.domainCache.Get(domain)
	if !ok {
		return net.IP{}, false
	}

	f.domainCache.Get(ipUintI)

	ipUint := ipUintI.(uint32)
	return uintToIP(ipUint), true
}

func (f *fakeIPCache) GetByIP(ip net.IP) (domain string, ok bool) {
	f.mut.Lock()
	defer f.mut.Unlock()

	ipUint := ipToUint(ip.To4())
	logger.Debugf("ipuint: %d", ipUint)

	domainIfc, ok := f.ipCache.Get(ipUint)
	if !ok {
		logger.Info("ip not found")
		return
	}

	domain = domainIfc.(string)

	f.domainCache.Get(domain)
	return
}
