/*
 * Copyright © 2019 Hedzr Yeh.
 */

package im_grpc_server

import (
	"fmt"
	"github.com/hedzr/voxr-common/tool"
	"github.com/hedzr/voxr-common/vxconf"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"strconv"
)

// var server *grpc.Server

func StopServer(server *grpc.Server) {
	closeServer(server)
}

func closeServer(server *grpc.Server) {
	if server != nil {
		server.Stop()
		server = nil
	}
}

// StartServer
func StartServer(initClient func(*grpc.Server)) (server *grpc.Server) {
	if server != nil {
		return
	}

	var listen net.Listener
	server, listen = newServer(initClient)
	if server == nil || listen == nil {
		return
	}

	go func() {
		logrus.Printf("      gRPC Started at %v successfully...", listen)
		// grpc Serve: it will block here
		if err := server.Serve(listen); err != nil {
			logrus.Fatalf("      gRPC Failed to serve: %v", err)
		}
		logrus.Println("      gRPC exiting...")
	}()
	return
}

func newServer(initFunc func(*grpc.Server)) (server *grpc.Server, listen net.Listener) {
	// if server != nil {
	// 	return
	// }

	var main_grpc = fmt.Sprintf("server.grpc.%v", vxconf.GetStringR("server.grpc.main", "voxr-lite"))
	if len(main_grpc) == 0 {
		return
	}

	grpcListen, id, disabled, port := tool.LoadGRPCListen(main_grpc)
	if disabled {
		logrus.Warnf("      gRPC listen on %v but disabled. id = %v", grpcListen, id)
		logrus.Println("      gRPC exiting...")
		return
	}

	if s := os.Getenv("GRPC_PORT"); len(s) > 0 {
		if port1, err := strconv.Atoi(s); err == nil {
			port = port1
		}
	}

	// find an available port between starting port and portMax
	var portMax = port + 10
	var err error
	for {
		listen, err = net.Listen("tcp", grpcListen) // listen on tcp4 and tcp6
		if err != nil {
			logrus.Warnf("      gRPC Failed to listen: %v", err)
			if port > portMax {
				logrus.Fatalf("      gRPC Failed to listen: %v, port = %v", err, port)
				return
			}
			grpcListen, port = tool.IncGrpcListen(main_grpc)
		} else {
			break
		}
	}

	logrus.Infof("      gRPC Listening at :%v....", port)

	// grpc 注册相应的 rpc 服务
	server = grpc.NewServer()
	initFunc(server)
	// switch id {
	// case "inx.im.apply":
	// 	pb.RegisterApplyServiceServer(server, &ApplyService{})
	// case "inx.im.core":
	// 	cs := &ImCoreService{}
	// 	cs.Init()
	// 	core.RegisterImCoreServer(server, cs)
	//
	// }
	reflection.Register(server)

	return
}
