package db

import (
	"context"
	"crypto/tls"
	"reflect"
	"time"

	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var DatabaseName string = "auth"

type AuthDb struct {
	db *mongo.Database
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
	return &AuthDb{db: db}
}

func (a *AuthDb) Login(tenant string) *LoginRepository {
	repo := odm.AbstractRepository{
		Db:             a.db,
		CollectionName: "login_" + tenant,
		Model:          reflect.TypeOf(models.LoginModel{}),
	}
	return &LoginRepository{repo}
}

func (a *AuthDb) Profile(tenant string) *ProfileRepository {
	repo := odm.AbstractRepository{
		Db:             a.db,
		CollectionName: "profile_" + tenant,
		Model:          reflect.TypeOf(models.ProfileModel{}),
	}
	return &ProfileRepository{repo}
}

func (a *AuthDb) Tenant() *TenantRepository {
	repo := odm.AbstractRepository{
		Db:             a.db,
		CollectionName: "tenant",
		Model:          reflect.TypeOf(models.TenantModel{}),
	}
	return &TenantRepository{repo}
}
