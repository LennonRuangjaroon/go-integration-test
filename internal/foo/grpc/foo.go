package grpc

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "integration-test/pkg/proto/foo"
)

type Server struct {
	pb.UnimplementedFooServiceServer
	mgClient *mongo.Client
}

func NewFooServer(client *mongo.Client) *Server {
	return &Server{
		mgClient: client,
	}
}

func (f *Server) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResponse, error) {
	return &pb.InsertResponse{
		Ok: true,
	}, nil
}

func (f *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "cannot get status %v", req.GetName())
	}

	var result struct {
		Name string `bson:"name" json:"name"`
	}

	collection := f.mgClient.Database("integration").Collection("test")
	filter := bson.M{"name": req.GetName()}

	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		fmt.Println("record does not exist!")
	} else if err != nil {
		fmt.Printf("cannot query mongo: [%v]", err)
	}

	return &pb.GetResponse{Name: result.Name}, nil
}
