package dispatcherclient

import (
	"net"

	"time"

	"github.com/lovelly/goworld/engine/netutil"
	"github.com/lovelly/goworld/engine/proto"
)

// DispatcherClient is a client connection to the dispatcher
type DispatcherClient struct {
	*proto.GoWorldConnection
}

func newDispatcherClient(conn net.Conn, autoFlush bool) *DispatcherClient {
	gwc := proto.NewGoWorldConnection(netutil.NewBufferedConnection(netutil.NetConnection{conn}), false, "")

	dc := &DispatcherClient{
		GoWorldConnection: gwc,
	}
	if autoFlush {
		go func() {
			//defer gwlog.Debugf("%s: auto flush routine quited", gwc)
			for !gwc.IsClosed() {
				time.Sleep(time.Millisecond * 10)
				dispatcherClientDelegate.HandleDispatcherClientBeforeFlush()

				err := gwc.Flush("DispatcherClient")
				if err != nil {
					break
				}
			}
		}()
	}
	return dc
}

// Close the dispatcher client
func (dc *DispatcherClient) Close() error {
	return dc.GoWorldConnection.Close()
}
