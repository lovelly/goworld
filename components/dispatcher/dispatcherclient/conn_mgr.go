package dispatcherclient

import (
	"time"

	"sync/atomic"

	"unsafe"

	"net"

	"github.com/pkg/errors"
	"github.com/lovelly/goworld/engine/config"
	"github.com/lovelly/goworld/engine/consts"
	"github.com/lovelly/goworld/engine/gwioutil"
	"github.com/lovelly/goworld/engine/gwlog"
	"github.com/lovelly/goworld/engine/gwutils"
	"github.com/lovelly/goworld/engine/netutil"
	"github.com/lovelly/goworld/engine/proto"
)

const (
	_LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR = time.Second
)

var (
	isReconnect               = false
	_dispatcherClient         *DispatcherClient // DO NOT access it directly
	dispatcherClientDelegate  IDispatcherClientDelegate
	dispatcherClientAutoFlush bool
	errDispatcherNotConnected = errors.New("dispatcher not connected")
)

func getDispatcherClient() *DispatcherClient { // atomic
	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
	return (*DispatcherClient)(unsafe.Pointer(atomic.LoadUintptr(addr)))
}

func setDispatcherClient(dc *DispatcherClient) { // atomic
	addr := (*uintptr)(unsafe.Pointer(&_dispatcherClient))
	atomic.StoreUintptr(addr, uintptr(unsafe.Pointer(dc)))
}

func assureConnectedDispatcherClient() *DispatcherClient {
	var err error
	dispatcherClient := getDispatcherClient()
	//gwlog.Debugf("assureConnectedDispatcherClient: _dispatcherClient", _dispatcherClient)
	for dispatcherClient == nil || dispatcherClient.IsClosed() {
		dispatcherClient, err = connectDispatchClient()
		if err != nil {
			gwlog.Errorf("Connect to dispatcher failed: %s", err.Error())
			time.Sleep(_LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}
		dispatcherClientDelegate.OnDispatcherClientConnect(dispatcherClient, isReconnect)

		setDispatcherClient(dispatcherClient)
		isReconnect = true

		gwlog.Infof("dispatcher_client: connected to dispatcher: %s", dispatcherClient)
	}

	return dispatcherClient
}

func connectDispatchClient() (*DispatcherClient, error) {
	dispatcherConfig := config.GetDispatcher()
	conn, err := netutil.ConnectTCP(dispatcherConfig.Ip, dispatcherConfig.Port)
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetReadBuffer(consts.DISPATCHER_CLIENT_READ_BUFFER_SIZE)
	tcpConn.SetWriteBuffer(consts.DISPATCHER_CLIENT_WRITE_BUFFER_SIZE)
	return newDispatcherClient(conn, dispatcherClientAutoFlush), nil
}

// IDispatcherClientDelegate defines functions that should be implemented by dispatcher clients
type IDispatcherClientDelegate interface {
	OnDispatcherClientConnect(dispatcherClient *DispatcherClient, isReconnect bool)
	HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet)
	HandleDispatcherClientDisconnect()
	HandleDispatcherClientBeforeFlush()
	//HandleDeclareService(entityID common.EntityID, serviceName string)
	//HandleCallEntityMethod(entityID common.EntityID, method string, args []interface{})
}

// Initialize the dispatcher client, only called by engine
func Initialize(delegate IDispatcherClientDelegate, autoFlush bool) {
	dispatcherClientDelegate = delegate
	dispatcherClientAutoFlush = autoFlush

	assureConnectedDispatcherClient()
	go gwutils.RepeatUntilPanicless(serveDispatcherClient) // start the recv routine
}

// GetDispatcherClientForSend returns the current dispatcher client for sending messages
func GetDispatcherClientForSend() *DispatcherClient {
	dispatcherClient := getDispatcherClient()
	return dispatcherClient
}

// serve the dispatcher client, receive RESPs from dispatcher and process
func serveDispatcherClient() {
	gwlog.Debugf("serveDispatcherClient: start serving dispatcher client ...")
	for {
		dispatcherClient := assureConnectedDispatcherClient()
		var msgtype proto.MsgType
		pkt, err := dispatcherClient.Recv(&msgtype)

		if err != nil {
			if gwioutil.IsTimeoutError(err) {
				continue
			}

			gwlog.TraceError("serveDispatcherClient: RecvMsgPacket error: %s", err.Error())
			dispatcherClient.Close()
			dispatcherClientDelegate.HandleDispatcherClientDisconnect()
			time.Sleep(_LOOP_DELAY_ON_DISPATCHER_CLIENT_ERROR)
			continue
		}

		if consts.DEBUG_PACKETS {
			gwlog.Debugf("%s.RecvPacket: msgtype=%v, payload=%v", dispatcherClient, msgtype, pkt.Payload())
		}
		dispatcherClientDelegate.HandleDispatcherClientPacket(msgtype, pkt)
	}
}
