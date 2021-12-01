package grpc

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	pb "integration-test/pkg/proto/foo"
	"log"
	"net"
	"testing"
)

type mongoContainer struct {
	testcontainers.Container
	URI string
}

func setUpServ() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	m, err := setUpMongo()
	if err != nil {
		return nil
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(m.URI))
	collection := client.Database("integration").Collection("test")
	collection.InsertOne(context.Background(), bson.M{"name": "bob"})

	pb.RegisterFooServiceServer(server, NewFooServer(client))

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func setUpMongo() (*mongoContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:4.2",
		ExposedPorts: []string{"27017/tcp"},
	}

	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(context.Background(), "27017")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(context.Background())
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("mongodb://%s:%s/integration", hostIP, mappedPort.Port())
	fmt.Println(uri)

	return &mongoContainer{
		Container: container,
		URI:       uri,
	}, nil
}

func TestFoo_GetName_ShouldBeSuccess(t *testing.T) {
	tests := []struct {
		name    string
		fooName string
		res     *pb.GetResponse
		errCode codes.Code
	}{
		{
			"invalid request with no have req param",
			"",
			&pb.GetResponse{
				Name: "",
			},
			codes.InvalidArgument,
		},
		{
			"valid request with have req param",
			"bob",
			&pb.GetResponse{
				Name: "bob",
			},
			codes.OK,
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(setUpServ()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewFooServiceClient(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &pb.GetRequest{Name: tt.fooName}

			response, err := client.Get(ctx, request)

			if response != nil {
				if response.GetName() != tt.res.GetName() {
					t.Error("response: expected", tt.res.GetName(), response.GetName())
				}
			}

			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != tt.errCode {
						t.Error("error code: expected", codes.InvalidArgument, er.Code())
					}
				}
			}
		})
	}
}
