// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package proxy

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/godbus/dbus/v5"
	"github.com/linuxdeepin/deepin-network-proxy/cgroups"
	"github.com/linuxdeepin/deepin-network-proxy/com"
	"github.com/linuxdeepin/deepin-network-proxy/config"
	"github.com/linuxdeepin/deepin-network-proxy/define"
	"github.com/linuxdeepin/deepin-network-proxy/iproute"
	"github.com/linuxdeepin/deepin-network-proxy/iptables"
	"github.com/linuxdeepin/deepin-network-proxy/tproxy"
	"github.com/linuxdeepin/go-lib/dbusutil"
	"github.com/linuxdeepin/go-lib/log"
)

var logger *log.Logger

const (
	BusServiceName = "org.deepin.dde.NetworkProxy1"
	BusPath        = "/org/deepin/dde/NetworkProxy1"
	BusInterface   = BusServiceName
)

// must ignore proxy proc
var mainProxy = []string{
	"/usr/lib/deepin-daemon/dde-proxy",
	"Qv2ray",
}

type proxyPrv struct {
	scope    define.Scope
	priority define.Priority

	// proxy message
	Proxies config.ScopeProxies
	Proxy   config.Proxy // current proxy

	// if proxy opened
	Enabled bool

	// handler manager
	manager *Manager

	// listener
	tcpHandler net.Listener
	udpHandler net.PacketConn

	// cgroup controller
	controller *cgroups.Controller

	// iptables chain rule slice[3]
	chains [2]*iptables.Chain

	// route rule
	ipRule *iproute.Rule

	// handler manager
	handlerMgr *tproxy.HandlerMgr

	dnsProxy *proxyDNS

	// handler
	uid uint32
	gid uint32

	// stop chan
	// stop bool
}

// init proxy private
func initProxyPrv(scope define.Scope, priority define.Priority) *proxyPrv {
	prv := &proxyPrv{
		scope:      scope,
		priority:   priority,
		handlerMgr: tproxy.NewHandlerMgr(scope),
		// stop:       true,
		Proxies: config.ScopeProxies{
			Proxies:      make(map[string][]config.Proxy),
			ProxyProgram: []string{},
			WhiteList:    []string{},
		},
	}

	prv.dnsProxy = newProxyDNS(prv)
	return prv
}

// proxy prepare
func (mgr *proxyPrv) startRedirect() error {
	// clean old redirect
	_ = mgr.firstClean()

	// make sure manager start init
	mgr.manager.Start()

	// create cgroups
	err := mgr.createCGroupController()
	if err != nil {
		logger.Warning("[%s] create cgroup failed, err: %v", mgr.scope, err)
	}

	// create iptables
	err = mgr.createTable()
	if err != nil {
		logger.Warning("[%s] create iptables failed, err: %v", mgr.scope, err)
		return err
	}
	err = mgr.appendRule()
	if err != nil {
		logger.Warning("[%s] append iptables failed, err: %v", mgr.scope, err)
		return err
	}

	err = mgr.createIpRule()
	if err != nil {
		logger.Warning("[%s] create ip rule failed, err: %v", err)
		return err
	}
	logger.Debugf("[%s] start tproxy iptables cgroups ipRule success", mgr.scope)

	//// first adjust cgroups
	//err = mgr.firstAdjustCGroups()
	//if err != nil {
	//	logger.Warningf("[%s] first adjust controller failed, err: %v", mgr.scope, err)
	//	return err
	//}
	//logger.Debugf("[%s] first adjust controller success", mgr.scope)
	return nil
}

//
func (mgr *proxyPrv) stopRedirect() error {
	// release iptables rules
	err := mgr.releaseRule()
	if err != nil {
		logger.Warningf("[%s] release iptables failed, err: %v", mgr.scope, err)
		return err
	}

	_ = mgr.attachBackUser()

	// release cgroups
	err = mgr.releaseController()
	if err != nil {
		logger.Warningf("[%s] release controller failed, err: %v", mgr.scope, err)
		return err
	}

	err = mgr.releaseIpRule()
	if err != nil {
		logger.Warningf("[%s] release ipRule failed, err: %v", mgr.scope, err)
	}

	// try to release manager
	err = mgr.manager.release()
	if err != nil {
		logger.Warningf("[%s] release manager failed, err: %v", mgr.scope, err)
		return err
	}

	logger.Debugf("[%s] stop tproxy iptables cgroups ipRule success", mgr.scope)
	return nil
}

// load config
func (mgr *proxyPrv) loadConfig() {
	// load proxy from manager
	mgr.Proxies, _ = mgr.manager.config.GetScopeProxies(mgr.scope)
	logger.Debugf("[%s] load config success, config: %v", mgr.scope, mgr.Proxies)
}

func (mgr *proxyPrv) saveManager(manager *Manager) {
	mgr.manager = manager
}

// write config
func (mgr *proxyPrv) writeConfig() error {
	// set and write config
	mgr.manager.config.SetScopeProxies(mgr.scope, mgr.Proxies)
	err := mgr.manager.WriteConfig()
	if err != nil {
		logger.Warning("[%s] write config failed, err:%v", mgr.scope, err)
		return err
	}
	return nil
}

// first clean
func (mgr *proxyPrv) firstClean() error {
	// get config path
	path, err := com.GetConfigDir()
	if err != nil {
		logger.Warningf("[%s] run first clean failed, config err: %v", mgr.scope, err)
		return err
	}
	// get script file path
	path = filepath.Join(path, define.ScriptName)
	// run script
	buf, err := com.RunScript(path, []string{"clear_" + mgr.scope.String()})
	if err != nil {
		logger.Debugf("[%s] run first clean script failed, out: %s, err: %v", mgr.scope, string(buf), err)
		return err
	}
	logger.Debugf("[%s] run first clean script success", mgr.scope)
	return nil
}

// cgroups
func (mgr *proxyPrv) GetCGroups() (string, *dbus.Error) {
	if mgr.controller == nil {
		return "", nil
	}
	path := mgr.controller.GetCGroupPath()
	// path := "/sys/fs/cgroup/unified/App.slice/cgroups.procs"
	_, err := os.Stat(path)
	if err != nil {
		logger.Warningf("app cgroups not exist, err: %v", err)
		return "", dbusutil.ToError(err)
	}
	return path, nil
}

// add pid to proc
func (mgr *proxyPrv) AddProc(pid int32) *dbus.Error {
	// controller
	if mgr.controller == nil {
		return dbusutil.ToError(errors.New("controller not exist"))
	}
	// attach pid
	err := cgroups.Attach(strconv.Itoa(int(pid)), mgr.controller.GetControlPath())
	if err != nil {
		logger.Debugf("attach %d to %s failed, err: %v", pid, mgr.controller.GetControlPath(), err)
		return dbusutil.ToError(err)
	}
	logger.Debugf("attach %d to %s success", pid, mgr.controller.GetControlPath())
	return nil
}

//func (mgr *proxyPrv) CreateCGroups(sender dbus.Sender, cgroup string) *dbus.Error {
//	con, err := dbusutil.NewSystemService()
//	if err != nil {
//		logger.Warningf("get session service failed, err: %v", err)
//		return dbusutil.ToError(err)
//	}
//	uid, err := con.GetConnUID(string(sender))
//	if err != nil {
//		logger.Warningf("get name owner failed, err: %v", err)
//		return dbusutil.ToError(err)
//	}
//	mgr.uid = int(uid)
//	mgr.cgroup = cgroup
//	return nil
//}
