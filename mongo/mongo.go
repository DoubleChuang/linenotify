package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Mongo struct {
	DB      *mongo.Database
	Client  *mongo.Client
	Context context.Context
	timeOut time.Duration
}

func New(connection, dbName string) (*Mongo, error) {
	timeOut := 10
	timeOutFormat := time.Duration(timeOut)
	ctx, cancel := context.WithTimeout(context.Background(), timeOutFormat*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connection))
	if err != nil {
		return nil, err
	}
	c, cancel := context.WithTimeout(context.Background(), timeOutFormat*time.Second)
	defer cancel()

	err = client.Ping(c, readpref.Primary())
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	return &Mongo{DB: db, Client: client, Context: ctx, timeOut: timeOutFormat}, nil
}

func (m *Mongo) Close() {
	m.Client.Disconnect(m.Context)
}

func (m *Mongo) InsertOne(content interface{}, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeOut*time.Second)
	defer cancel()
	collection := m.DB.Collection(name)
	_, err := collection.InsertOne(ctx, content)
	if err != nil {
		return err
	}
	return nil
}
