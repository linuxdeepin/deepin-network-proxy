// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package iproute

// run command to route
type RunCommand struct {
	soft string // ip rule and ip route

	action string
	mark   string
	table  string
}
