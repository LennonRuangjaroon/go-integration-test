package rest

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type Bar struct {
	mgClient *mongo.Client
}

func NewBar(c *mongo.Client) *Bar {
	return &Bar{mgClient: c}
}

func (b *Bar) Get(w http.ResponseWriter, r *http.Request) {
	nameVal := r.URL.Query().Get("name")
	if nameVal == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid req query param!"))
		return
	}

	collection := b.mgClient.Database("integration").Collection("test")

	var result struct {
		Name string `bson:"name" json:"name"`
	}

	filter := bson.M{"name": nameVal}
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if errors.Is(err, mongo.ErrNoDocuments) {
		fmt.Println("record does not exist!")
	} else if err != nil {
		fmt.Printf("cannot query mongo: [%v]", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.Name))
}
