PREFIX=/usr
PREFIXETC=/etc

LIB=lib
DEEPIN=deepin
PROXYFILE=deepin-proxy
DAEMON=deepin-daemon

GOPATH=/usr/share/gocode

GOPATH_DIR=gopath
GOPKG_PREFIX=github.com/linuxdeepin/deepin-network-proxy

GOBUILD = go build $(GO_BUILD_FLAGS)

all: build

prepare:
	@mkdir -p bin
	@mkdir -p ${GOPATH_DIR}/src/$(dir ${GOPKG_PREFIX});
	@ln -snf ../../../.. ${GOPATH_DIR}/src/${GOPKG_PREFIX};

Out/%:  prepare
	@echo $(GOPATH)
	GOPATH="${CURDIR}/${GOPATH_DIR}:$(GOPATH)" ${GOBUILD} -o bin/${@F} ${GOBUILD_OPTIONS} ${GOPKG_PREFIX}/out/${@F}

install:
	mkdir -p ${DESTDIR}${PREFIXETC}/${DEEPIN}/${PROXYFILE}
	install -v -D -m755 -t ${DESTDIR}${PREFIXETC}/${DEEPIN}/${PROXYFILE} misc/script/clean_script.sh
	install -v -D -m755 -t ${DESTDIR}${PREFIXETC}/${DEEPIN}/${PROXYFILE} misc/proxy/proxy.yaml
	install -v -D -m755 -t ${DESTDIR}${PREFIX}/share/dbus-1/system.d misc/proxy/org.deepin.dde.NetworkProxy1.conf
	install -v -D -m755 -t ${DESTDIR}${PREFIX}/share/dbus-1/system-services misc/proxy/org.deepin.dde.NetworkProxy1.service
	install -v -D -m755 -t ${DESTDIR}${PREFIX}/${LIB}/${DAEMON} bin/dde-proxy


clean:
	-rm -rf bin


build: prepare Out/dde-proxy
