// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package DBus

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/linuxdeepin/deepin-network-proxy/fakeip"
	"github.com/miekg/dns"
)

type proxyDNS struct {
	prv    *proxyPrv
	server *dns.Server

	pool *fakeip.Pool
}

func newProxyDNS(prv *proxyPrv) *proxyDNS {
	p := &proxyDNS{
		prv: prv,
	}

	p.server = &dns.Server{
		Net:     "udp",
		Handler: p,
	}

	return p
}

func (p *proxyDNS) resolveDomain(domain string) netip.Addr {
	domain = strings.TrimRight(domain, ".")

	return p.pool.GetIP(domain)
}

func (p *proxyDNS) getDomainFromFakeIP(ip net.IP) (string, bool) {
	addr, _ := netip.AddrFromSlice(ip)

	return p.pool.GetDomain(addr)
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
			m.Answer = []dns.RR{}
			m.Authoritative = true
			m.RecursionAvailable = true
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

	if p.pool == nil {
		v := p.prv.Proxies.FakeIPRange
		if v == "" {
			v = "192.168.0.0/20"
		}
		logger.Info("fake ip range: ", v)
		prefix := netip.MustParsePrefix(v)
		p.pool, _ = fakeip.NewPool(&prefix, fakeip.CacheMaxSize)
	}

	return p.server.ListenAndServe()
}

func (p *proxyDNS) stopDNSProxy() error {
	p.pool.Clear()
	return p.server.Shutdown()
}
