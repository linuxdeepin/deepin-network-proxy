module github.com/linuxdeepin/deepin-network-proxy

go 1.15

replace github.com/linuxdeepin/go-lib => github.com/Decodetalkers/go-lib v0.0.0-20230207102150-285b65f72371

replace github.com/linuxdeepin/go-dbus-factory => github.com/Decodetalkers/go-dbus-factory v0.0.0-20230214081229-2794c96a723b

require (
	github.com/godbus/dbus/v5 v5.1.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/linuxdeepin/go-dbus-factory v0.0.0-00010101000000-000000000000
	github.com/linuxdeepin/go-lib v0.0.0-00010101000000-000000000000
	github.com/miekg/dns v1.1.52
	golang.org/x/sys v0.6.0
	gopkg.in/yaml.v2 v2.4.0
)
