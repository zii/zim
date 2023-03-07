package base

import (
	"fmt"
	"testing"
	"time"
)

// mock time
func TestTime1(_ *testing.T) {
	//SetNow(time.Date(2021, 1, 1, 0, 0, 0, 0, time.Now().Location()))
	//now := Now().Unix()
	now := time.Now()
	st := GetDayStart(now)
	fmt.Println("的:", st)
	today := Today()
	fmt.Println("today:", today)
}
