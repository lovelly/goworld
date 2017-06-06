package main

import (
	"net"

	"fmt"

	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/proto"
	"github.com/xiaonanln/vacuum/netutil"
)

type DispatcherClientProxy struct {
	proto.GoWorldConnection
}

func newDispatcherClientProxy(conn net.Conn) *DispatcherClientProxy {
	return &DispatcherClientProxy{GoWorldConnection: proto.NewGoWorldConnection(conn)}
}

func (dcp *DispatcherClientProxy) serve() {
	// Serve the dispatcher client from game / gate
	defer func() {
		dcp.Close()

		err := recover()
		if err != nil && !netutil.IsConnectionClosed(err) {
			gwlog.Error("Client %s paniced with error: %v", dcp, err)
		}
	}()

	gwlog.Info("New dispatcher client: %s", dcp)
	for {
		var msgtype proto.MsgType_t
		var data []byte
		_, err := dcp.Recv(&msgtype, &data)
		if err != nil {
			gwlog.Panic(err)
		}

		gwlog.Info("%s.RecvPacket: msgtype=%v, data=%v", dcp, msgtype, data)

	}
}

func (dcp *DispatcherClientProxy) String() string {
	return fmt.Sprintf("DispatcherClientProxy<%s>", dcp.RemoteAddr())
}