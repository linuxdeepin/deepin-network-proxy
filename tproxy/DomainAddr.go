// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tproxy

import "strconv"

type DomainAddr struct {
	network string
	Domain  string
	Port    int
}

func NewDomainAddr(network string, domain string, port int) *DomainAddr {
	return &DomainAddr{
		network: network,
		Domain:  domain,
		Port:    port,
	}
}

func (a *DomainAddr) Network() string {
	return a.network
}

func (a *DomainAddr) String() string {
	return a.Domain + ":" + strconv.Itoa(a.Port)
}
