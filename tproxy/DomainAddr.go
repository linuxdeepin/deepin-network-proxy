package TProxy

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
