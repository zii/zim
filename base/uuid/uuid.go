// 生成全局唯一ID
// sonyflake缺点是较慢较每秒最多生成25600,优点是可以用到170多年才消耗完.
// snowflake缺点是50多年消耗完, 优点是更精确(1ms), 每秒生成的多
package uuid

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"zim.cn/base"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

// counter map
var sfmap map[string]*sonyflake.Sonyflake

var ErrorInitUUID = fmt.Errorf("ErrorInitUUID")

func init() {
	sfmap = make(map[string]*sonyflake.Sonyflake)
}

func getIntEnv(name string) (int, error) {
	s := os.Getenv(name)
	if s == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func lower8BitPrivateIP() (uint8, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint8(ip[3]), nil
}

// name: 计数器名称
// env DC_ID: 数据中心ID, 0～15
// env MACHINE_ID: 机器ID, 0~4095
//
// 如果DC_ID和MACHINE_ID都是0, 则按ip+进程id作为MachineID
// 如果DC_ID!=0, MACHINE_ID=0, 则按DC_ID+ip+pid作为MachineID
func InitSF(name string) (*sonyflake.Sonyflake, error) {
	if sf, ok := sfmap[name]; ok {
		return sf, nil
	}
	dc_id, err := getIntEnv("DC_ID")
	if err != nil {
		return nil, errors.New("DC_ID_ENV_INVALID")
	}
	machine_id, err := getIntEnv("MACHINE_ID")
	if err != nil {
		return nil, errors.New("MACHINE_ID_ENV_INVALID")
	}
	if dc_id > 15 || dc_id < 0 {
		return nil, errors.New("DC_ID_INVALID")
	}
	if machine_id > 4095 || machine_id < 0 {
		return nil, errors.New("MACHINE_ID_INVALID")
	}
	st, err := time.Parse("2006-01-02", "2022-09-01")
	if err != nil {
		return nil, err
	}
	sf = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: st, // 固定:项目开始日期
		MachineID: func() (uint16, error) {
			if dc_id != 0 && machine_id != 0 {
				id := (uint16(dc_id) << 4) ^ uint16(machine_id)
				return id, nil
			}
			ip3, err := lower8BitPrivateIP()
			if err != nil {
				return 0, err
			}
			id := (uint16(ip3) << 8) ^ uint16(os.Getpid())
			if dc_id != 0 {
				id = (uint16(dc_id) << 12) ^ id
			}
			return id, nil
		}, // nil:以内网ip
	})
	if sf == nil {
		return nil, ErrorInitUUID
	}
	sfmap[name] = sf
	return sf, nil
}

func MustInit(name string) *sonyflake.Sonyflake {
	sf, err := InitSF(name)
	if err != nil {
		panic(err)
	}
	return sf
}

// 老式初始化
func InitUUID() error {
	_, err := InitSF("")
	return err
}

// 通过名称获取计数器对象
func GetSF(name string) *sonyflake.Sonyflake {
	sf := sfmap[name]
	if sf == nil {
		sf = MustInit(name)
	}
	return sf
}

// 桶的个数必须是奇数才会分布均匀
func NextID(name string) int64 {
	sf := GetSF(name)
	v, err := sf.NextID()
	base.Raise(err)
	return int64(v & 0x7fffffffffffffff)
}

func NextIDString(name string) string {
	id := NextID(name)
	return strconv.FormatInt(id, 36)
}
