package db

import (
	"github.com/Kotlang/authGo/logger"
	"github.com/Kotlang/authGo/models"
	odm "github.com/SaiNageswarS/mongo-odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TenantRepository struct {
	odm.AbstractRepository
}

func (t *TenantRepository) FindOneByToken(token string) chan *models.TenantModel {
	ch := make(chan *models.TenantModel)

	go func() {
		res := <-t.FindOne(bson.M{"token": token})

		if res.Err != nil {
			logger.Error("Error fetching tenant", zap.Error(res.Err))
			ch <- nil
			return
		}

		ch <- res.Value.(*models.TenantModel)
	}()
	return ch
}
