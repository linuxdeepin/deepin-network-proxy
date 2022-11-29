// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/linuxdeepin/deepin-network-proxy/netlink"
)

func main() {
	err := netlink.CreateProcsService()
	if err != nil {
		return
	}
}
