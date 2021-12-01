package rest

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"testing"
)

type BarTestSuite struct {
	suite.Suite
}

func (bt *BarTestSuite) SetupSuite() {

}

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarTestSuite))
}

func (bt *BarTestSuite) TestBar() {

}

type mongoContainer struct {
	testcontainers.Container
	URI string
}

func setUpServ() {
	m, err := setUpMongo()
	if err != nil {
		return
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(m.URI))
	collection := client.Database("integration").Collection("test")
	collection.InsertOne(context.Background(), bson.M{"name": "alice"})

	bar := NewBar(client)
	http.HandleFunc("/bars", bar.Get)

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("cannot start http :8080 [%v]", err)
		}
	}()
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

func TestBar_GetName_ShouldBeSuccess(t *testing.T) {
	tests := []struct {
		name       string
		barName    string
		res        string
		statusCode int
	}{
		{
			"invalid request with no have req param",
			"",
			"invalid req query param!",
			http.StatusBadRequest,
		},
		{
			"valid request with have req param",
			"alice",
			"alice",
			http.StatusOK,
		},
	}

	setUpServ()

	c := http.Client{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, "http://localhost:8080/bars?name="+tt.barName, nil)
			if err != nil {
				return
			}

			response, err := c.Do(request)
			if err != nil {
				t.Error()
			}

			res, _ := ioutil.ReadAll(response.Body)
			defer response.Body.Close()

			if response.StatusCode == tt.statusCode {
				if string(res) != tt.res {
					t.Error("response: expected", tt.res)
				}
			}
		})
	}
}
