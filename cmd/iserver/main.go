// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/rpc"
	flag "github.com/spf13/pflag"
)

var (
	configfile = flag.StringP("config", "f", "", "Configuration `file`")
	help       = flag.BoolP("help", "h", false, "Display available options")
)

func getLogLevel(l string) ilog.Level {
	switch l {
	case "debug":
		return ilog.LevelDebug
	case "info":
		return ilog.LevelInfo
	case "warn":
		return ilog.LevelWarn
	case "error":
		return ilog.LevelError
	case "fatal":
		return ilog.LevelFatal
	default:
		return ilog.LevelDebug
	}
}

func initLogger(logConfig *common.LogConfig) {
	if logConfig == nil {
		return
	}
	logger := ilog.New()
	if logConfig.AsyncWrite {
		logger.AsyncWrite()
	}
	if logConfig.ConsoleLog != nil && logConfig.ConsoleLog.Enable {
		consoleWriter := ilog.NewConsoleWriter()
		consoleWriter.SetLevel(getLogLevel(logConfig.ConsoleLog.Level))
		logger.AddWriter(consoleWriter)
	}
	if logConfig.FileLog != nil && logConfig.FileLog.Enable {
		fileWriter := ilog.NewFileWriter(logConfig.FileLog.Path)
		fileWriter.SetLevel(getLogLevel(logConfig.FileLog.Level))
		logger.AddWriter(fileWriter)
	}
	ilog.InitLogger(logger)
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
	}

	if *configfile == "" {
		*configfile = os.Getenv("GOPATH") + "/src/github.com/iost-official/Go-IOS-Protocol/config/iserver.yaml"
	}

	conf := common.NewConfig(*configfile)

	initLogger(conf.Log)

	ilog.Infof("Config Information:\n%v", conf.YamlString())

	glb, err := global.New(conf)
	if err != nil {
		ilog.Fatalf("create global failed. err=%v", err)
	}

	var app common.App

	p2pService, err := p2p.NewNetService(conf.P2P)
	if err != nil {
		ilog.Fatalf("network initialization failed, stop the program! err:%v", err)
	}
	app = append(app, p2pService)

	accSecKey := glb.Config().ACC.SecKey
	acc, err := account.NewAccount(common.Base58Decode(accSecKey))
	if err != nil {
		ilog.Fatalf("NewAccount failed, stop the program! err:%v", err)
	}
	account.MainAccount = acc

	blkCache, err := blockcache.NewBlockCache(glb)
	if err != nil {
		ilog.Fatalf("blockcache initialization failed, stop the program! err:%v", err)
	}

	sync, err := synchronizer.NewSynchronizer(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("synchronizer initialization failed, stop the program! err:%v", err)
	}
	app = append(app, sync)

	var txp txpool.TxPool
	txp, err = txpool.NewTxPoolImpl(glb, blkCache, p2pService)
	if err != nil {
		ilog.Fatalf("txpool initialization failed, stop the program! err:%v", err)
	}
	app = append(app, txp)

	rpcServer := rpc.NewRPCServer(txp, blkCache, glb)
	app = append(app, rpcServer)

	consensus, err := consensus.Factory(
		"pob",
		acc, glb, blkCache, txp, p2pService, sync, account.WitnessList) //witnessList)
	if err != nil {
		ilog.Fatalf("consensus initialization failed, stop the program! err:%v", err)
	}
	app = append(app, consensus)

	err = app.Start()
	if err != nil {
		ilog.Fatal("start iserver failed. err=%v", err)
	}

	waitExit()

	app.Stop()
	ilog.Stop()
}

func waitExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	ilog.Infof("IOST server received interrupt[%v], shutting down...", i)
}
