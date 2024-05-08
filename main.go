package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"strconv"
	"sxp-server/helper"
	"sxp-server/logger"
	"sxp-server/model"
	"sxp-server/pb"
	"sxp-server/service"
	"sxp-server/tracer"
)

type modelSever struct {
	log *logger.ZapLog
	pb.UnimplementedModelServer
}

func newModelServer() *modelSever {
	return &modelSever{
		log: logger.GetLogger(),
	}
}

func (m *modelSever) GetModel(ctx context.Context, request *pb.ModelRequest) (res *pb.ModelResponse, err error) {
	// 设置trailer
	defer func() {
		if err = helper.TrailerResponse(ctx); err != nil {
			return
		}
	}()
	err = helper.HeadResponse(ctx, "1")
	if err != nil {
		m.log.Error(err)
		return
	}
	pm := model.ProductMap
	name := pm[request.GetProductId()].Name
	res = &pb.ModelResponse{}
	res.Product = name
	fmt.Println(res)
	return
}

// UpdateModel
//
//	@Description: 更新product model
//	@receiver m
//	@param ctx
//	@param request
//	@return res
//	@return err
func (m *modelSever) UpdateModel(ctx context.Context, request *pb.UpdateRequest) (res *pb.UpdateResponse, err error) {
	// 设置trailer
	defer func() {
		if err = helper.TrailerResponse(ctx); err != nil {
			m.log.Error(err)
			return
		}
	}()
	err = helper.HeadResponse(ctx, "1")
	if err != nil {
		m.log.Error(err)
		return
	}

	pro := model.ProductMap[request.GetProductId()]
	pro.Name = request.GetProduct()
	model.ProductMap[request.GetProductId()] = pro
	res = &pb.UpdateResponse{
		Message: "更新数据成功",
	}
	fmt.Println(model.ProductMap)
	return
}

// GetByStatus
//
//	@Description: 流式处理
//	@receiver m
//	@param stream
//	@return err
func (m *modelSever) GetByStatus(stream pb.Model_GetByStatusServer) (err error) {
	closeCh := make(chan struct{})
	//// 在defer中创建trailer记录函数的返回时间.
	//defer func() {
	//	if err = helper.TrailerResponse(stream.Context()); err != nil {
	//		m.log.Error(err)
	//		return
	//	}
	//}()
	err = helper.HeadResponse(stream.Context(), "1")
	if err != nil {
		m.log.Error(err)
		return
	}
	data := make([]model.Product, 0)
	go func() {
		for {
			request, er := stream.Recv()
			if er != nil && er != io.EOF {
				m.log.Errorf("Recv error: %s", er.Error())
				return
			} else if er == io.EOF {
				m.log.Info("Recv EOF")
				closeCh <- struct{}{}
				return
			}
			pm := model.ProductMap
			for _, v := range pm {
				if v.Status == request.GetStatus() {
					data = append(data, v)
					err = stream.Send(&pb.StatusResponse{
						ProductId: strconv.Itoa(v.Id),
						Product:   v.Name,
						Status:    v.Status,
					})
					if err != nil {
						m.log.Error(err)
						return
					}
				}
			}
		}
	}()
	for {
		select {
		case <-stream.Context().Done():
			m.log.Error("超时错误！")
			return
		case <-closeCh:
			m.log.Info("发送完毕！")
			if err = helper.TrailerResponse(stream.Context()); err != nil {
				m.log.Error(err)
				return
			}
			return
		}
	}

}

func main() {
	model.Init()
	lis, err := net.Listen("tcp", ":9011")
	if err != nil {
		log.Fatalf("failed to listen: %s", err.Error())
		return
	}
	// 初始化tracer
	trace, _, err := tracer.NewJaegerTracer("sxp-server", "192.168.111.143:6831")
	l := service.NewZapLog()
	grpc_zap.ReplaceGrpcLoggerV2(l.Zl)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			service.UnaryInterceptor,
			grpc_zap.UnaryServerInterceptor(l.Zl),
			grpc_recovery.UnaryServerInterceptor(),
			tracer.UnaryTraceInterceptor(trace))),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(grpc_middleware.ChainStreamServer(
			service.StreamInterceptor,
			grpc_zap.StreamServerInterceptor(l.Zl),
			grpc_recovery.StreamServerInterceptor(),
			tracer.StreamTraceInterceptor(trace))))) // 创建gRPC服务器
	pb.RegisterModelServer(s, newModelServer()) // 在gRPC服务端注册服务
	// 启动服务
	err = s.Serve(lis)

}
