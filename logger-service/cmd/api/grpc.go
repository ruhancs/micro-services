package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"logger-service/logs"
	"net"

	"google.golang.org/grpc"
)


type LogServer struct {
	logs.UnimplementedLogServiceServer
	Model data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, in *logs.LogRequest) (*logs.LogResponse, error) {
	//pegar a input do enviada para o servico grpc
	input := in.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Model.LogEntry.Insert(logEntry)
	if err != nil {
		resp := &logs.LogResponse{Result: "failed"}
		return resp,err
	}

	res := &logs.LogResponse{Result: "logged"}
	return res,nil
}

func (app *Config) GrpcListen() {
	lis,err := net.Listen("tcp", fmt.Sprintf(":%s", gRPCPort))
	if err != nil {
		log.Fatalf("failed to listen for grpc: %v", err)
	}

	s := grpc.NewServer()
	logs.RegisterLogServiceServer(s, &LogServer{Model: app.Models})

	log.Printf("grpc server running on port %s", gRPCPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to listen for grpc: %v", err)
	}
}