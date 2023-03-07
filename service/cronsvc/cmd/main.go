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

	"github.com/gorilla/mux"

	"zim.cn/base/log"

	"zim.cn/service/cronsvc/jobs"

	"zim.cn/biz/config"

	"zim.cn/base"
)

const (
	VERSION = "0.1.0"
)

var args struct {
	Port     int
	ConfFile string
	DumpFile string
}

func healthHandler(_ http.ResponseWriter, _ *http.Request) {
}

func init() {
	if base.RELEASE {
		log.SetLevel(log.LvInfo)
	}
	flag.IntVar(&args.Port, "port", 1860, "server listening port")
	flag.StringVar(&args.ConfFile, "config", "config.toml", "config file")
	flag.StringVar(&args.DumpFile, "dump", "cronjob.dump", "dump file")
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.Parse()

	log.Println("version:", VERSION)
	log.Println("RELEASE:", base.RELEASE)
	log.Println("config:", args.ConfFile)

	// 加载全局配置
	conf := config.LoadConfigFile(args.ConfFile)
	conf.Init()
	DUMPFILE = args.DumpFile
	log.Println("DUMPFILE:", args.DumpFile)

	router := mux.NewRouter()
	// 健康检查
	router.HandleFunc("/health", healthHandler)

	// restore
	Restore()
	log.Println("restored!", CheckPoints.Count())

	// schedule
	go schedule("ClearOnline", jobs.ClearOnline, 10)

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		listenAddr := fmt.Sprintf("0.0.0.0:%d", args.Port)
		log.Println("listening:", listenAddr)
		server := &http.Server{
			Addr:        listenAddr,
			Handler:     router,
			IdleTimeout: 300 * time.Second,
		}
		err := server.ListenAndServe()
		//err := http.ListenAndServe(listenAddr, router)
		base.Raise(err)
	}()

	// Run!
	log.Println("exit", <-errc)

	// dump
	Dump()
	log.Println("dumped!", len(CheckPoints))
}
