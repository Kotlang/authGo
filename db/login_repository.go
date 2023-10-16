package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type LoginRepositoryInterface interface {
	odm.AbstractRepositoryInterface[models.LoginModel]
	FindOneByEmail(email string) chan *models.LoginModel
	FindOneByPhone(phone string) chan *models.LoginModel
	FindByIds(ids []string) (chan []models.LoginModel, chan error)
}

type LoginRepository struct {
	odm.AbstractRepository[models.LoginModel]
}

func (t *LoginRepository) FindOneByEmail(email string) chan *models.LoginModel {
	ch := make(chan *models.LoginModel)

	go func() {
		id := (&models.LoginModel{Email: email}).Id()
		resultChan, errorChan := t.FindOneById(id)

		select {
		case res := <-resultChan:
			ch <- res
		case err := <-errorChan:
			logger.Error("Error fetching login info", zap.Error(err))
		}
	}()
	return ch
}

func (t *LoginRepository) FindOneByPhone(phone string) chan *models.LoginModel {
	ch := make(chan *models.LoginModel)

	go func() {
		id := (&models.LoginModel{Phone: phone}).Id()
		resultChan, errorChan := t.FindOneById(id)

		select {
		case res := <-resultChan:
			ch <- res
		case err := <-errorChan:
			logger.Error("Error fetching login info", zap.Error(err))
			ch <- nil
		}
	}()
	return ch
}

func (t *LoginRepository) FindByIds(ids []string) (chan []models.LoginModel, chan error) {
	return t.Find(bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}
