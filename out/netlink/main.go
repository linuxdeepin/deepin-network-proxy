// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	netlink "github.com/ArisAachen/deepin-network-proxy/netlink"
)

func main() {
	err := netlink.CreateProcsService()
	if err != nil {
		return
	}
}
