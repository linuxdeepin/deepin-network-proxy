// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package proxy

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type proxyDNS struct {
	prv    *proxyPrv
	server *dns.Server

	fIP   fakeIP
	cache *fakeIPCache
}

func newProxyDNS(prv *proxyPrv) *proxyDNS {
	p := &proxyDNS{
		prv: prv,
	}

	p.fIP = newFakeIP(net.IP{192, 168, 135, 0}, 24)
	p.cache = newFakeIPCache()

	p.server = &dns.Server{
		Net:     "udp",
		Handler: p,
	}

	return p
}

func (p *proxyDNS) resolveDomain(domain string) net.IP {
	domain = strings.TrimRight(domain, ".")

	i, ok := p.cache.GetByDomain(domain)
	if ok {
		return i
	}

	ip := p.fIP.new()
	logger.Debugf("Query A for %s: %s", domain, ip)
	p.cache.Add(domain, ip)

	return ip
}

func (p *proxyDNS) getDomainFromFakeIP(ip net.IP) (string, bool) {
	return p.cache.GetByIP(ip)
}

func (p *proxyDNS) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			ip := p.resolveDomain(q.Name)
			rr, err := dns.NewRR(fmt.Sprintf("%s 0 A %s", q.Name, ip))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		case dns.TypeAAAA:
			logger.Debugf("Query AAAA for %s", q.Name)
		}
	}
}

func (p *proxyDNS) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		p.parseQuery(m)
	}

	w.WriteMsg(m)
}

func (p *proxyDNS) startDNSProxy() error {
	p.server.Addr = fmt.Sprintf("127.0.0.1:%d", p.prv.Proxies.DNSPort)
	logger.Info("dns listen addr:", p.server.Addr)

	return p.server.ListenAndServe()
}

func (p *proxyDNS) stopDNSProxy() error {
	return p.server.Shutdown()
}
