// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package fakeip

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createPool(prefix *netip.Prefix, size int) (*Pool, error) {
	pool, err := NewPool(prefix, size)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func TestPool_Basic(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.0.0/24")
	pool, err := createPool(&prefix, CacheMaxSize)
	assert.Nil(t, err)

	first := pool.GetIP("foo.com")
	last := pool.GetIP("bar.com")
	bar, exist := pool.GetDomain(last)

	assert.True(t, first == netip.AddrFrom4([4]byte{192, 168, 0, 2}))
	assert.True(t, pool.GetIP("foo.com") == netip.AddrFrom4([4]byte{192, 168, 0, 2}))
	assert.True(t, last == netip.AddrFrom4([4]byte{192, 168, 0, 3}))
	assert.True(t, exist)
	assert.Equal(t, bar, "bar.com")

	_, exist = pool.GetDomain(netip.AddrFrom4([4]byte{192, 168, 0, 3}))
	assert.True(t, exist)
	_, exist = pool.GetDomain(netip.AddrFrom4([4]byte{192, 168, 0, 4}))
	assert.False(t, exist)
	_, exist = pool.GetDomain(netip.MustParseAddr("::1"))
	assert.False(t, exist)

}

func TestPool_BasicV6(t *testing.T) {
	prefix := netip.MustParsePrefix("2001:4860:4860::8888/118")
	pool, err := createPool(&prefix, CacheMaxSize)
	assert.Nil(t, err)

	first := pool.GetIP("foo.com")
	last := pool.GetIP("bar.com")
	bar, exist := pool.GetDomain(last)

	assert.True(t, first == netip.MustParseAddr("2001:4860:4860:0000:0000:0000:0000:8802"))
	assert.True(t, pool.GetIP("foo.com") == netip.MustParseAddr("2001:4860:4860:0000:0000:0000:0000:8802"))
	assert.True(t, last == netip.MustParseAddr("2001:4860:4860:0000:0000:0000:0000:8803"))
	assert.True(t, exist)
	assert.Equal(t, bar, "bar.com")

	_, exist = pool.GetDomain(netip.MustParseAddr("2001:4860:4860:0000:0000:0000:0000:8803"))
	assert.True(t, exist)
	_, exist = pool.GetDomain(netip.MustParseAddr("2001:4860:4860:0000:0000:0000:0000:8804"))
	assert.False(t, exist)
	_, exist = pool.GetDomain(netip.MustParseAddr("127.0.0.1"))
	assert.False(t, exist)

}

func TestPool_CycleUsed(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.0.16/28")
	pool, err := createPool(&prefix, 10)
	assert.Nil(t, err)

	foo := pool.GetIP("foo.com")
	bar := pool.GetIP("bar.com")
	for i := 0; i < 11; i++ {
		pool.GetIP(fmt.Sprintf("%d.com", i))
	}
	baz := pool.GetIP("baz.com")
	next := pool.GetIP("foo.com")
	assert.True(t, foo == baz)
	assert.True(t, next == bar)
}

func TestPool_MaxCacheSize(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.0.1/24")
	pool, _ := NewPool(&prefix, 2)

	first := pool.GetIP("foo.com")
	pool.GetIP("bar.com")
	pool.GetIP("baz.com")
	next := pool.GetIP("foo.com")

	assert.False(t, first == next)
}

func TestPool_DoubleMapping(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.0.1/24")
	pool, _ := NewPool(&prefix, 2)

	// fill cache
	fooIP := pool.GetIP("foo.com")
	bazIP := pool.GetIP("baz.com")

	// make foo.com hot
	pool.GetIP("foo.com")

	// should drop baz.com
	barIP := pool.GetIP("bar.com")

	_, fooExist := pool.GetDomain(fooIP)
	_, bazExist := pool.GetDomain(bazIP)
	_, barExist := pool.GetDomain(barIP)

	newBazIP := pool.GetIP("baz.com")

	assert.True(t, fooExist)
	assert.False(t, bazExist)
	assert.True(t, barExist)

	assert.False(t, bazIP == newBazIP)
}

func TestPool_Error(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.0.1/31")
	_, err := NewPool(&prefix, CacheMaxSize)

	assert.Error(t, err)
}

func TestPool_Clear(t *testing.T) {
	prefix := netip.MustParsePrefix("192.168.0.0/20")
	pool, err := NewPool(&prefix, CacheMaxSize)

	assert.Nil(t, err)

	fooIP := pool.GetIP("foo.com")
	_, fooExist := pool.GetDomain(fooIP)

	assert.True(t, fooExist)

	pool.Clear()

	_, fooExist = pool.GetDomain(fooIP)
	assert.False(t, fooExist)
}
