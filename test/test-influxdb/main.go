package main

import (
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	"os"
	"time"
)

// 创建一个client
func ExampleClient() client.Client {
	// NOTE: this assumes you've setup a user and have setup shell env variables,
	// namely INFLUX_USER/INFLUX_PWD. If not just omit Username/Password below.
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://10.10.10.86:8086",
		Username: os.Getenv("root"),
		Password: os.Getenv("root"),
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	return cli
}

// 把数据写入influxdb
func ExampleClient_write(cli client.Client) {
	// Make client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://10.10.10.86:8086",
		Username: os.Getenv("root"),
		Password: os.Getenv("root"),
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "BumbleBeeTuna", //数据库名
		Precision: "s",             //时间精度秒
	})
	// Create a point and add to batch
	tags := map[string]string{"cpu": "cpu-total"} //查询的索引
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	} //记录值
	pt, err := client.NewPoint("cpu_usage", tags, fields, time.Now()) //将创建的表名为cpu_usage的表以及内容字段放入pt
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	bp.AddPoint(pt) //把表放入创建的point中

	// Write the batch
	c.Write(bp) //写入创建的client中
}

// 查询
func ExampleClient_query(cli client.Client) {
	// Make client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://10.10.10.86:8086",
		Username: os.Getenv("root"),
		Password: os.Getenv("root"),
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	q := client.NewQuery("select * from cpu_usage", "BumbleBeeTuna", "ns")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	}
}
func main() {
	conn := ExampleClient()
	ExampleClient_write(conn)
	ExampleClient_query(conn)
}
