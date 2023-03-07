package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"zim.cn/base"

	client2 "github.com/rpcxio/rpcx-consul/client"
	"github.com/smallnest/rpcx/client"

	"github.com/rpcxio/rpcx-consul/serverplugin"
	"github.com/smallnest/rpcx/server"
)

type Args struct {
	A int
	B int
}

type Reply struct {
	C int
}

func mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func testConsul() {
	s := server.NewServer()
	r := serverplugin.NewConsulRegisterPlugin(
		serverplugin.WithConsulServiceAddress("tcp@127.0.0.1:1850"),
		serverplugin.WithConsulServers([]string{"127.0.0.1:8500"}),
		serverplugin.WithConsulBasePath("/rpcx_test"),
		//serverplugin.WithConsulMetrics(metrics.NewRegistry()),
		serverplugin.WithConsulUpdateInterval(time.Minute),
	)
	err := r.Start()
	if err != nil {
		return
	}
	s.Plugins.Add(r)

	s.RegisterFunctionName("zim", "mul", mul, "")
	go s.Serve("tcp", "0.0.0.0:1850")
	defer s.Close()

	// client
	d, _ := client2.NewConsulDiscovery("/rpcx_test", "zim", []string{"127.0.0.1:8500"}, nil)
	if len(r.Services) != 1 {
		log.Fatal("failed to register services in consul")
	}
	xclient := client.NewXClient("zim", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xclient.Close()

	args := &Args{
		A: 2,
		B: 3,
	}
	reply := &Reply{}
	err = xclient.Call(context.Background(), "mul", args, reply)
	base.Raise(err)
	fmt.Println("reply:", reply.C)

	if err := r.Stop(); err != nil {
		log.Fatal(err)
	}
}

func testNormal() {
	port := 1850
	s := server.NewServer()
	s.RegisterFunctionName("zim", "mul", mul, "")
	go s.Serve("tcp", fmt.Sprintf("0.0.0.0:%d", port))

	addr := fmt.Sprintf("localhost:%d", port)
	d, _ := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	xclient := client.NewXClient("zim", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xclient.Close()

	args := &Args{
		A: 2,
		B: 3,
	}
	reply := &Reply{}
	err := xclient.Call(context.Background(), "mul", args, reply)
	base.Raise(err)
	fmt.Println("reply:", reply.C)
}

func main() {
	//testNormal()
	testConsul()
}
