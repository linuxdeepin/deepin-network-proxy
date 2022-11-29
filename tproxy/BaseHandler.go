// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tproxy

import (
	"fmt"
	"net"
	"sync"

	"github.com/linuxdeepin/deepin-network-proxy/config"
	"github.com/linuxdeepin/deepin-network-proxy/define"
	"github.com/linuxdeepin/go-lib/log"
)

var logger *log.Logger

// handler module

type BaseHandler interface {
	// connection
	Tunnel() error

	// close
	Close()  // direct close handler
	Remove() // remove self from map
	AddMgr(mgr *HandlerMgr)

	// write and read
	WriteRemote([]byte) error
	WriteLocal([]byte) error
	ReadRemote([]byte) error
	ReadLocal([]byte) error
	Communicate()
}

// proto
type ProtoTyp string

const (
	NoneProto ProtoTyp = "no-proto"
	HTTP      ProtoTyp = "http"
	SOCKS4    ProtoTyp = "socks4"
	SOCKS5TCP ProtoTyp = "socks5-tcp"
	SOCKS5UDP ProtoTyp = "socks5-udp"
)

func BuildProto(proto string) (ProtoTyp, error) {
	switch proto {
	case "no-proxy":
		return NoneProto, nil
	case "http":
		return HTTP, nil
	case "socks4":
		return SOCKS4, nil
	case "socks5-tcp":
		return SOCKS5TCP, nil
	case "socks5-udp":
		return SOCKS5UDP, nil
	default:
		return NoneProto, fmt.Errorf("scope is invalid, scope: %v", proto)
	}
}

func (Typ ProtoTyp) String() string {
	switch Typ {
	case NoneProto:
		return "no-proxy"
	case HTTP:
		return "http"
	case SOCKS4:
		return "socks4"
	case SOCKS5TCP:
		return "socks5-tcp"
	case SOCKS5UDP:
		return "socks5-udp"
	default:
		return "unknown-proto"
	}
}

// proxy server
type proxyServer struct {
	server string
	port   int
	auth
}

// auth message
type auth struct {
	user     string
	password string
}

// handler key in case keep the same handler
type HandlerKey struct {
	SrcAddr string
	DstAddr string
}

// manager all handler
type HandlerMgr struct {
	handlerLock sync.Mutex
	handlerMap  map[ProtoTyp]map[HandlerKey]BaseHandler // handlerMap sync.Map map[http udp]map[HandlerKey]BaseHandler

	// scope [global,app]
	scope define.Scope
	// chan to stop accept
	stop chan bool
}

func NewHandlerMgr(scope define.Scope) *HandlerMgr {
	return &HandlerMgr{
		scope:      scope,
		handlerMap: make(map[ProtoTyp]map[HandlerKey]BaseHandler),
		stop:       make(chan bool),
	}
}

// add handler to mgr
func (mgr *HandlerMgr) AddHandler(typ ProtoTyp, key HandlerKey, base BaseHandler) {
	// add lock
	mgr.handlerLock.Lock()
	defer mgr.handlerLock.Unlock()
	// check if handler already exist
	baseMap, ok := mgr.handlerMap[typ]
	if !ok {
		baseMap = make(map[HandlerKey]BaseHandler)
		mgr.handlerMap[typ] = baseMap
	}
	_, ok = baseMap[key]
	if ok {
		// if exist already, should ignore
		logger.Debugf("[%s] key has already in map, type: %v, key: %v", mgr.scope, typ, key)
		return
	}
	// add handler
	baseMap[key] = base
	logger.Debugf("[%s] handler add to manager success, type: %v, key: %v", mgr.scope, typ, key)
}

// close and remove base handler
func (mgr *HandlerMgr) CloseBaseHandler(typ ProtoTyp, key HandlerKey) {
	mgr.handlerLock.Lock()
	defer mgr.handlerLock.Unlock()
	baseMap, ok := mgr.handlerMap[typ]
	if !ok {
		logger.Debugf("[%s] delete base map dont exist in map", mgr.scope)
		return
	}
	base, ok := baseMap[key]
	if !ok {
		logger.Debugf("[%s] delete key dont exist in base map, key: %v", mgr.scope, key)
		return
	}
	// close and delete
	base.Close()
	delete(baseMap, key)
	logger.Debugf("[%s] delete key successfully, key: %v", mgr.scope, key)
}

// close handler according to proto
func (mgr *HandlerMgr) CloseTypHandler(typ ProtoTyp) {
	mgr.handlerLock.Lock()
	defer mgr.handlerLock.Unlock()
	baseMap, ok := mgr.handlerMap[typ]
	if !ok {
		return
	}
	// close handler
	for _, base := range baseMap {
		base.Close()
	}
	// delete proto handler
	delete(mgr.handlerMap, typ)
}

// close all handler
func (mgr *HandlerMgr) CloseAll() {
	for proto, _ := range mgr.handlerMap {
		mgr.CloseTypHandler(proto)
	}
}

func NewHandler(proto ProtoTyp, scope define.Scope, key HandlerKey, proxy config.Proxy, lAddr net.Addr, rAddr net.Addr, lConn net.Conn) BaseHandler {
	// search proto
	switch proto {
	case HTTP:
		return NewHttpHandler(scope, key, proxy, lAddr, rAddr, lConn)
	case SOCKS4:
		return NewSock4Handler(scope, key, proxy, lAddr, rAddr, lConn)
	case SOCKS5TCP:
		return NewTcpSock5Handler(scope, key, proxy, lAddr, rAddr, lConn)
	case SOCKS5UDP:
		return NewUdpSock5Handler(scope, key, proxy, lAddr, rAddr, lConn)
	default:
		logger.Warningf("unknown proto type: %v", proto)
	}
	return nil
}

func init() {
	logger = log.NewLogger("proxy/tproxy")
}
