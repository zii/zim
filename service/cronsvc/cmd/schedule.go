package main

import (
	"runtime/debug"
	"time"

	"zim.cn/base/log"

	cmap "github.com/orcaman/concurrent-map"
)

// jobname: 本轮开始时间戳
var CheckPoints = cmap.New() // jobname: time

func schedule(jobname string, f func(), interval int) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			debug.PrintStack()
		}
	}()

	now := int(time.Now().Unix())
	v, ok := CheckPoints.Get(jobname)
	var t int
	if ok {
		t = v.(int)
	}
	already := now - t
	var until int
	if already > 0 && already < interval {
		until = interval - already
		time.Sleep(time.Duration(until) * time.Second)
	}
	log.Printf("schedule: %d/%d %s \n", until, interval, jobname)
	for {
		f()
		CheckPoints.Set(jobname, int(time.Now().Unix()))
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
