package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.uber.org/zap"
)

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
