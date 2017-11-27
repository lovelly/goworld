package game

import (
	"flag"

	"math/rand"
	"time"

	"os"

	// for go tool pprof
	_ "net/http/pprof"

	"runtime"

	"os/signal"

	"syscall"

	"fmt"

	"github.com/lovelly/goworld/components/dispatcher/dispatcherclient"
	"github.com/lovelly/goworld/engine/binutil"
	"github.com/lovelly/goworld/engine/config"
	"github.com/lovelly/goworld/engine/crontab"
	"github.com/lovelly/goworld/engine/entity"
	"github.com/lovelly/goworld/engine/gwlog"
	"github.com/lovelly/goworld/engine/kvdb"
	"github.com/lovelly/goworld/engine/netutil"
	"github.com/lovelly/goworld/engine/proto"
	"github.com/lovelly/goworld/engine/storage"
)

var (
	gameid                       uint16
	configFile                   string
	logLevel                     string
	restore                      bool
	runInDaemonMode              bool
	gameService                  *_GameService
	signalChan                   = make(chan os.Signal, 1)
	gameDispatcherClientDelegate = &dispatcherClientDelegate{}
)

func parseArgs() {
	var gameidArg int
	flag.IntVar(&gameidArg, "gid", 1, "set gameid")
	flag.StringVar(&configFile, "configfile", "", "set config file path")
	flag.StringVar(&logLevel, "log", "", "set log level, will override log level in config")
	flag.BoolVar(&restore, "restore", false, "restore from freezed state")
	flag.BoolVar(&runInDaemonMode, "d", false, "run in daemon mode")
	flag.Parse()
	gameid = uint16(gameidArg)
}

// Run runs the game server
//
// This is the main game server loop
func Run(delegate IGameDelegate) {
	rand.Seed(time.Now().UnixNano())
	parseArgs()

	if runInDaemonMode {
		daemoncontext := binutil.Daemonize()
		defer daemoncontext.Release()
	}

	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	if gameid <= 0 {
		gwlog.Errorf("gameid %d is not valid, should be positive", gameid)
		os.Exit(1)
	}

	gameConfig := config.GetGame(gameid)
	if gameConfig == nil {
		gwlog.Errorf("game %d's config is not found", gameid)
		os.Exit(1)
	}

	if gameConfig.GoMaxProcs > 0 {
		gwlog.Infof("SET GOMAXPROCS = %d", gameConfig.GoMaxProcs)
		runtime.GOMAXPROCS(gameConfig.GoMaxProcs)
	}
	if logLevel == "" {
		logLevel = gameConfig.LogLevel
	}
	binutil.SetupGWLog(fmt.Sprintf("game%d", gameid), logLevel, gameConfig.LogFile, gameConfig.LogStderr)

	storage.Initialize()
	kvdb.Initialize()
	crontab.Initialize()

	binutil.SetupHTTPServer(gameConfig.HTTPIp, gameConfig.HTTPPort, nil)

	entity.SetSaveInterval(gameConfig.SaveInterval)

	gameService = newGameService(gameid, delegate)

	dispatcherclient.Initialize(gameDispatcherClientDelegate, false)

	setupSignals()

	gameService.run(restore)
}

func setupSignals() {
	gwlog.Infof("Setup signals ...")
	signal.Ignore(syscall.Signal(12), syscall.SIGPIPE)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.Signal(10))

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGTERM {
				// terminating game ...
				gwlog.Infof("Terminating game service ...")
				gameService.terminate()
				waitGameServiceStateSatisfied(func(rs int) bool {
					return rs != rsTerminating
				})
				if gameService.runState.Load() != rsTerminated {
					// game service is not terminated successfully, abort
					gwlog.Errorf("Game service is not terminated successfully, back to running ...")
					continue
				}

				waitEntityStorageFinish()

				gwlog.Infof("Game %d shutdown gracefully.", gameid)
				os.Exit(0)
			} else if sig == syscall.Signal(10) || sig == syscall.SIGINT {
				// SIGUSR1 => dump game and close
				// freezing game ...
				gwlog.Infof("Freezing game service ...")

				gameService.freeze()
				waitGameServiceStateSatisfied(func(rs int) bool { // wait until not running
					return rs != rsRunning
				})
				waitGameServiceStateSatisfied(func(rs int) bool {
					return rs != rsFreezing
				})

				if gameService.runState.Load() != rsFreezed {
					// game service is not freezed successfully, abort
					gwlog.Errorf("Game service is not freezed successfully, back to running ...")
					continue
				}

				waitEntityStorageFinish()

				gwlog.Infof("Game %d freezed gracefully.", gameid)
				os.Exit(0)
			} else {
				gwlog.Errorf("unexpected signal: %s", sig)
			}
		}
	}()
}

func waitGameServiceStateSatisfied(s func(rs int) bool) {
	waitCounter := 0
	for {
		state := gameService.runState.Load()
		if s(state) {
			break
		}
		waitCounter++
		if waitCounter%10 == 0 {
			gwlog.Infof("game service status: %d", state)
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func waitEntityStorageFinish() {
	// wait until entity storage's queue is empty
	gwlog.Infof("Closing Entity Storage ...")
	storage.Shutdown()
	gwlog.Infof("*** DB OK ***")
}

type dispatcherClientDelegate struct {
	lastCollectEntitySyncInfosTime time.Time
}

func (delegate *dispatcherClientDelegate) OnDispatcherClientConnect(dispatcherClient *dispatcherclient.DispatcherClient, isReconnect bool) {
	// called when connected / reconnected to dispatcher (not in main routine)
	var isRestore bool
	if !isReconnect {
		isRestore = restore
	}

	//go func() {
	//	for !dispatcherClient.IsClosed() {
	//		time.Sleep(time.Millisecond * 10)
	//		err := dispatcherClient.Flush()
	//		if err != nil {
	//			break
	//		}
	//	}
	//}()

	dispatcherClient.SendSetGameID(gameid, isReconnect, isRestore)
}

var lastWarnGateServiceQueueLen = 0

func (delegate *dispatcherClientDelegate) HandleDispatcherClientPacket(msgtype proto.MsgType, packet *netutil.Packet) {
	gameService.packetQueue <- packetQueueItem{ // may block the dispatcher client routine
		msgtype: msgtype,
		packet:  packet,
	}
}

func (delegate *dispatcherClientDelegate) HandleDispatcherClientDisconnect() {
	gwlog.Errorf("Disconnected from dispatcher, try reconnecting ...")
}

func (delegate *dispatcherClientDelegate) HandleDispatcherClientBeforeFlush() {
	// collect all sync infos from entities and group them by target gates
	now := time.Now()
	if now.Sub(delegate.lastCollectEntitySyncInfosTime) >= time.Millisecond*100 {
		delegate.lastCollectEntitySyncInfosTime = now
		entity.CollectEntitySyncInfos()
	}
}

// GetGameID returns the current Game Server ID
func GetGameID() uint16 {
	return gameid
}
