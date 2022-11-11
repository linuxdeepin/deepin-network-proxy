// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package Define

// proxy name
/*
	usage:
	1. use to create DBus project, to mark current proxy type
	2. use to create Iptables chain name
	3. use to create cgroups controller name
*/
type Scope string

const (
	Main   Scope = "Main"
	App    Scope = "App"
	Global Scope = "Global"
)

func (s Scope) String() string {
	switch s {
	case Main:
		return "Main"
	case App:
		return "App"
	case Global:
		return "Global"
	default:
		return "unknown scope"
	}
}

// proxy type
/*
	usage:
	1. use to check recv dbus method
	2. use to make
	use to mark current proxy type
*/
const (
	// basic type
	HTTP  = "http"
	SOCK4 = "sock4"
	SOCK5 = "sock5"

	// extends type
	SOCK5UDP = "sock5-udp"
	SOCK5TCP = "sock5-tcp"
)

type Priority int

// proxy priority
const (
	MainPriority Priority = iota
	AppPriority
	GlobalPriority
)

const (
	ConfigName = "proxy.yaml"
	ScriptName = "clean_script.sh"
)
