package db

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/Kotlang/authGo/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var DatabaseName string = "auth"

type AuthDb struct {
	Db *mongo.Database
}

func NewAuthDb(mongo_uri string) *AuthDb {
	mongoOpts := options.Client().ApplyURI(mongo_uri)
	mongoOpts.TLSConfig.MinVersion = tls.VersionTLS12
	mongoOpts.TLSConfig.InsecureSkipVerify = true

	client, err := mongo.NewClient(mongoOpts)
	if err != nil {
		logger.Fatal("Failed to connect to mongo", zap.Error(err))
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		logger.Fatal("Failed to connect to mongo", zap.Error(err))
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatal("Failed to connect to mongo", zap.Error(err))
	}

	db := client.Database(DatabaseName)

	return &AuthDb{
		Db: db,
	}
}

func GetTenantCollectionName(collectionName, tenant string) string {
	return collectionName + "_" + tenant
}
