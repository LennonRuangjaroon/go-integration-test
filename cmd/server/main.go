package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	grpc2 "integration-test/internal/foo/grpc"
	"integration-test/internal/foo/rest"
	pb "integration-test/pkg/proto/foo"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Server running ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("error listener: %v \n", err)
	}

	server := grpc.NewServer()

	pb.RegisterFooServiceServer(server, grpc2.NewFooServer(client))

	go func() {
		err := server.Serve(listener)
		if err != nil {
			fmt.Printf("cannot start http :50051 [%v]", err)
			return
		}
	}()

	bar := rest.NewBar(client)
	http.HandleFunc("/bars", bar.Get)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Printf("cannot start http :8080 [%v]", err)
			return
		}
	}()

	gfs := make(chan os.Signal, 1)
	signal.Notify(gfs, syscall.SIGEMT, syscall.SIGINT)
	sig := <-gfs

	fmt.Printf("caught signal [%+v] then graceful stop.\n", sig)
}
