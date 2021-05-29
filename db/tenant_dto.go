package db

import (
	"context"

	"github.com/Kotlang/authGo/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var tenantCollectionName string = "tenant"

type TenantDto struct {
	collection *mongo.Collection
}

type TenantModel struct {
	Name  string `bson:"_id"`
	Token string `bson:"token"`
	Stage string `bson:"stage"`
}

func NewTenantDto(db *AuthDb) *TenantDto {
	return &TenantDto{
		collection: db.Db.Collection(tenantCollectionName),
	}
}

func (t *TenantDto) FindOneByToken(token string) chan *TenantModel {
	ch := make(chan *TenantModel)

	go func() {
		tenantBson := t.collection.FindOne(context.Background(), bson.M{"token": token})

		if tenantBson.Err() != nil {
			logger.Error("Error fetching tenant", zap.Error(tenantBson.Err()))
			ch <- nil
			return
		}
		tenant := &TenantModel{}
		tenantBson.Decode(tenant)
		ch <- tenant
	}()
	return ch
}
