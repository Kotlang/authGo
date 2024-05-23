package db

import (
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TenantModel struct {
	Name  string `bson:"_id"`
	Token string `bson:"token"`
	Stage string `bson:"stage"`
}

func (m *TenantModel) Id() string {
	return m.Name
}

type TenantRepositoryInterface interface {
	odm.BootRepository[TenantModel]
	FindOneByToken(token string) chan *TenantModel
}

type TenantRepository struct {
	odm.UnimplementedBootRepository[TenantModel]
}

func (t *TenantRepository) FindOneByToken(token string) chan *TenantModel {
	ch := make(chan *TenantModel)

	go func() {
		resultChan, errorChan := t.FindOne(bson.M{"token": token})

		select {
		case res := <-resultChan:
			ch <- res
		case err := <-errorChan:
			logger.Error("Error fetching tenant info", zap.Error(err))
			ch <- nil
		}
	}()
	return ch
}
