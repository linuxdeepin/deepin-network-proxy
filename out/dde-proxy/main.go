// SPDX-FileCopyrightText: 2022 UnionTech Software Technology Co., Ltd.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/linuxdeepin/deepin-network-proxy/proxy"
	"github.com/linuxdeepin/go-lib/log"
)

func main() {
	logger := log.NewLogger("proxy")
	manager := proxy.NewManager()
	err := manager.Init()
	if err != nil {
		logger.Warningf("manager init failed, err: %v", err)
		return
	}
	// load config
	_ = manager.LoadConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}
	// export dbus service
	err = manager.Export()
	if err != nil {
		logger.Warningf("manager export failed, err: %v", err)
		return
	}
	// wait
	manager.Wait()
}
