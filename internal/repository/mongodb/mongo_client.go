/*
---------------------------------------------------------------------------------------
File: books_client.go
Autor: Aldenor
Data: 18-04-2026
Finalidade:
Tipo concreto que realiza a conexão com o banco de dados MongoDB, criando um cliente
para a collection Books.
---------------------------------------------------------------------------------------
//Conexão com o Banco de Dados - MongoDB

	ctx := context.Background()
	mongoClient := mongodb.NewMongoDB(cfg)
	if err := mongoClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mongoClient.DisconnectMongo(ctx); err != nil {
			log.Println("erro ao encerrar conexão com MongoDB:", err)
		}
	}()
	log.Println("MongoDB conectado com sucesso")

	//Criação do Cliente Collection
	booksCollection, err := mongoClient.GetMongoCollection(ctx, cfg.MdDB, "books")
	if err != nil {
		log.Fatal(err)
	}
*/
package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"microsrv/internal/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	mongoOnce   sync.Once
	mongoClient *mongo.Client
	mongoErr    error
)

type MongoClient struct {
	Cfg *config.Config
}

func NewMongoDB(cfg *config.Config) *MongoClient {
	return &MongoClient{Cfg: cfg}
}
func (obj *MongoClient) Connect(ctx context.Context) error {
	_, err := obj.GetMongoClient(ctx)
	return err
}

func (db *MongoClient) GetMongoClient(ctx context.Context) (*mongo.Client, error) {
	mongoOnce.Do(func() {
		host := db.Cfg.MdHost
		port := db.Cfg.MdPort
		user := db.Cfg.MdUser
		pass := db.Cfg.MdPass

		if host == "" || port == "" || user == "" || pass == "" {
			mongoErr = fmt.Errorf("configuração do MongoDB incompleta")
			return
		}

		escapedUser := url.QueryEscape(user)
		escapedPass := url.QueryEscape(pass)

		mongoURI := fmt.Sprintf(
			"mongodb://%s:%s@%s:%s/?authSource=admin",
			escapedUser,
			escapedPass,
			host,
			port,
		)

		_, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
		if err != nil {
			mongoErr = fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
			return
		}

		pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
		defer pingCancel()

		if err := client.Ping(pingCtx, nil); err != nil {
			_ = client.Disconnect(context.Background())
			mongoErr = fmt.Errorf("erro ao validar conexão com MongoDB: %w", err)
			return
		}

		mongoClient = client
	})

	return mongoClient, mongoErr
}

func (db *MongoClient) GetMongoCollection(ctx context.Context, dbName, collectionName string) (*mongo.Collection, error) {
	client, err := db.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}

	return client.Database(dbName).Collection(collectionName), nil
}

func (db *MongoClient) DisconnectMongo(ctx context.Context) error {
	if mongoClient == nil {
		return nil
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		return fmt.Errorf("erro ao desconectar MongoDB: %w", err)
	}

	mongoClient = nil
	return nil
}
