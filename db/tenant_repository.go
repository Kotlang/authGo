package db

import (
	"github.com/Kotlang/authGo/logger"
	"github.com/Kotlang/authGo/models"
	odm "github.com/SaiNageswarS/mongo-odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TenantRepository struct {
	odm.AbstractRepository
}

func NewTenantRepository(db *mongo.Database) *TenantRepository {
	return &TenantRepository{odm.AbstractRepository{db}}
}

func (t *TenantRepository) FindOneByToken(token string) chan *models.TenantModel {
	ch := make(chan *models.TenantModel)

	go func() {
		tenant := &models.TenantModel{}
		err := <-t.FindOne(tenant, bson.M{"token": token})

		if err != nil {
			logger.Error("Error fetching tenant", zap.Error(err))
			ch <- nil
			return
		}

		ch <- tenant
	}()
	return ch
}
