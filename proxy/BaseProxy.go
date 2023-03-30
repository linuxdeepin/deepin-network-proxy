// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package proxy

import (
	"github.com/godbus/dbus/v5"
	"github.com/linuxdeepin/deepin-network-proxy/config"
	"github.com/linuxdeepin/deepin-network-proxy/define"
	"github.com/linuxdeepin/go-lib/dbusutil"
)

// https://www.kernel.org/doc/Documentation/networking/tproxy.txt

type BaseProxy interface {
	// DBus method
	StartProxy(sender dbus.Sender, proto string, name string, udp bool) *dbus.Error
	StopProxy() *dbus.Error
	SetProxies(proxies config.ScopeProxies) *dbus.Error
	ClearProxy() *dbus.Error
	GetProxy() (string, *dbus.Error)
	AddProxy(proto string, name string, jsonProxy []byte) *dbus.Error
	GetCGroups() (string, *dbus.Error)

	// manager
	loadConfig()
	saveManager(manager *Manager)

	// getScope() tProxy.ProxyScope
	getDBusPath() dbus.ObjectPath
	getScope() define.Scope

	// get cgroup v2 level
	getCGroupPriority() define.Priority

	//// cgroup v2
	//addCGroupExes(procs []string)
	//delCGroupExes(procs []string)

	// iptables
	appendRule() error
	releaseRule() error

	// export DBus service
	export(service *dbusutil.Service) error
}

// new proxy according to scope
func newProxy(scope define.Scope) BaseProxy {
	switch scope {
	case define.App:
		return NewAppProxy()
	case define.Global:
		return NewGlobalProxy()
	default:
		logger.Warning("init unknown scope type")
		return nil
	}
}

func init() {}
