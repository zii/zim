package main

import (
	"encoding/json"
	"os"

	"zim.cn/base/log"
)

type DumpTable struct {
	CheckPoints map[string]int `json:"check_points"`
}

var DUMPFILE = "./cronjob.dump"

func Dump() {
	dupTab := new(DumpTable)
	dupTab.CheckPoints = make(map[string]int)
	for t := range CheckPoints.IterBuffered() {
		dupTab.CheckPoints[t.Key] = t.Val.(int)
	}
	f, err := os.OpenFile(DUMPFILE, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Error("dump:", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.Encode(dupTab)
}

func Restore() {
	dupTab := new(DumpTable)
	f, err := os.OpenFile(DUMPFILE, os.O_RDONLY, 0600)
	if err != nil {
		log.Error("restore open:", err)
		return
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&dupTab)
	if err != nil {
		log.Error("restore:", err)
		return
	}
	for k, v := range dupTab.CheckPoints {
		CheckPoints.Set(k, v)
	}
	os.Remove(DUMPFILE)
}
