package DBus

import (
	define "github.com/ArisAachen/deepin-network-proxy/define"
	"github.com/godbus/dbus"
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
