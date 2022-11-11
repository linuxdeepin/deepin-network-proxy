package TProxy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	config "github.com/linuxdeepin/deepin-network-proxy/config"
	define "github.com/linuxdeepin/deepin-network-proxy/define"
)

type Sock4Handler struct {
	handlerPrv
}

func NewSock4Handler(scope define.Scope, key HandlerKey, proxy config.Proxy, lAddr net.Addr, rAddr net.Addr, lConn net.Conn) *Sock4Handler {
	// create new handler
	handler := &Sock4Handler{
		handlerPrv: createHandlerPrv(SOCKS4, scope, key, proxy, lAddr, rAddr, lConn),
	}
	// add self to private parent
	handler.saveParent(handler)
	return handler
}

func (handler *Sock4Handler) Tunnel() error {
	// dial proxy server
	rConn, err := handler.dialProxy()
	if err != nil {
		logger.Warningf("[sock4] failed to dial proxy server, err: %v", err)
		return err
	}
	// check type
	var port uint16
	var ip net.IP
	dominname := ""
	switch addr := handler.rAddr.(type) {
	case *net.TCPAddr:
		ip = addr.IP
	case *DomainAddr:
		port = uint16(addr.Port)
		ip = net.IPv4(0x00, 0x00, 0x00, 0x01)
		dominname = addr.Domain
	default:
		logger.Warning("[sock4] tunnel addr type is not tcp")
		return errors.New("type is not tcp")
	}

	// sock4 dont support password auth
	auth := auth{
		user: handler.proxy.UserName,
	}
	/*
					sock4 connect request
				+----+----+----+----+----+----+----+----+----+----+....+----+
				| VN | CD | DSTPORT |      DSTIP        | USERID       |NULL|
				+----+----+----+----+----+----+----+----+----+----+....+----+
		           1    1      2              4           variable       1
	*/
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(0x04) // sock version
	buf.WriteByte(0x01) // connect command

	// convert port 2 byte
	if port == 0 {
		port = 80
	}

	portByte := make([]byte, 2)
	binary.BigEndian.PutUint16(portByte, port)
	buf.Write(portByte)

	// add ip and user
	buf.Write(ip.To4())
	if auth.user != "" {
		buf.WriteString(auth.user)
	}
	buf.WriteByte(0x00)

	// add domainname
	if dominname != "" {
		buf.WriteString(dominname)
		buf.WriteByte(0x00)
	}

	// request proxy connect rConn server
	logger.Debugf("[sock4] send connect request, buf: %v", buf.Bytes())
	_, err = rConn.Write(buf.Bytes())
	if err != nil {
		logger.Warningf("[sock4] send connect request failed, err: %v", err)
		return err
	}

	// resp
	tmp := buf.Bytes()
	_, err = io.ReadFull(rConn, tmp[0:2])
	if err != nil {
		logger.Warningf("[sock4] connect response failed, err: %v", err)
		return err
	}
	/*
					sock4 server response
				+----+----+----+----+----+----+----+----+
				| VN | CD | DSTPORT |      DSTIP        |
				+----+----+----+----+----+----+----+----+
		          1    1      2              4

	*/
	// 0   0x5A
	if tmp[0] != 0 || tmp[1] != 90 {
		logger.Warningf("[sock4] proto is invalid, sock type: %v, code: %v", tmp[0], tmp[1])
		return fmt.Errorf("sock4 proto is invalid, sock type: %v, code: %v", tmp[0], tmp[1])
	}

	// port and ip
	_, err = io.ReadFull(rConn, tmp[0:6])
	if err != nil {
		logger.Warningf("[sock4] connect response failed, err: %v", err)
		return err
	}

	logger.Debugf("[sock4] port and ip: %v", tmp[0:6])
	logger.Debugf("[sock4] proxy: tunnel create success, [%s] -> [%s] -> [%s]",
		handler.lConn.RemoteAddr(), rConn.RemoteAddr(), handler.rAddr.String())
	// save rConn handler
	handler.rConn = rConn
	return nil
}
