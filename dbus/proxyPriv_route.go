// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package DBus

import (
	"strconv"

	route "github.com/linuxdeepin/deepin-network-proxy/ip_route"
)

// create ip rule
func (mgr *proxyPrv) createIpRule() error {
	action := route.RuleAction{}
	selector := route.RuleSelector{
		// fwmark 8080
		Fwmark: strconv.Itoa(mgr.Proxies.TPort),
	}
	// ip rule add fwmark 8080 table 100
	rule, err := mgr.manager.mainRoute.CreateRule(action, selector)
	if err != nil {
		return err
	}
	mgr.ipRule = rule
	return nil
}

// release ip rule
func (mgr *proxyPrv) releaseIpRule() error {
	buf, err := mgr.ipRule.Remove()
	if err != nil {
		logger.Warning("[%s] release rule failed, out: %s, err: %v", mgr.scope, string(buf), err)
		return err
	}
	logger.Debugf("[%s] release rule success", mgr.scope)
	return nil
}
