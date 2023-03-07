package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zim.cn/base/log"

	"github.com/gorilla/mux"
	"zim.cn/service"

	"zim.cn/base/uuid"
	"zim.cn/biz/config"

	"zim.cn/base"
)

const (
	VERSION = "0.1.0"
)

var args struct {
	Port     int
	ConfFile string
}

func healthHandler(_ http.ResponseWriter, _ *http.Request) {
}

func init() {
	if base.RELEASE {
		log.SetLevel(log.LvInfo)
	}
	flag.StringVar(&args.ConfFile, "config", "config.toml", "config file")
	flag.IntVar(&args.Port, "port", 1937, "server listening port")
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.Parse()

	log.Println("version:", VERSION)
	log.Println("RELEASE:", base.RELEASE)
	log.Println("config:", args.ConfFile)

	err := uuid.InitUUID()
	base.Raise(err)

	conf := config.LoadConfigFile(args.ConfFile)
	conf.Init()

	listenAddr := fmt.Sprintf("0.0.0.0:%d", args.Port)
	router := mux.NewRouter()
	// 健康检查
	router.HandleFunc("/health", healthHandler)
	router.HandleFunc("/ws", wsHandler)

	service.RegisterMethodRaw(router, "/node/statics", Node_statics)

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		log.Println("listening:", listenAddr)
		err := http.ListenAndServe(listenAddr, router)
		base.Raise(err)
	}()

	// subscribe loop
	go hub.run()

	// Run!
	log.Println("exit:", <-errc)
}
