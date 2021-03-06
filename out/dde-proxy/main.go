package main

import (
	proxyDBus "github.com/ArisAachen/deepin-network-proxy/dbus"
	"github.com/linuxdeepin/go-lib/log"
)

func main() {
	logger := log.NewLogger("proxy")
	manager := proxyDBus.NewManager()
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
