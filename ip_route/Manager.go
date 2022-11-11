// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package IpRoute

import "github.com/linuxdeepin/go-lib/log"

var logger *log.Logger

type Manager struct {
	routes map[string]*Route
}

// create manager
func NewManager() *Manager {
	manager := &Manager{
		routes: make(map[string]*Route),
	}
	return manager
}

// create route
func (m *Manager) CreateRoute(name string, node RouteNodeSpec, info RouteInfoSpec) (*Route, error) {
	// create route
	route := &Route{
		table: name,
		Node:  node,
		Info:  info,
	}
	err := route.create()
	if err != nil {
		return nil, err
	}
	return route, nil
}

func init() {
	logger = log.NewLogger("proxy/iproute")
}
