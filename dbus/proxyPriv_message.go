// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package DBus

import (
	"github.com/godbus/dbus"
	define "github.com/linuxdeepin/deepin-network-proxy/define"
)

// scope
// rewrite get scope
func (mgr *proxyPrv) getScope() define.Scope {
	return mgr.scope
}

func (mgr *proxyPrv) getDBusPath() dbus.ObjectPath {
	path := BusPath + "/" + mgr.scope.String()
	return dbus.ObjectPath(path)
}
