package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.uber.org/zap"
)

type LoginRepository struct {
	odm.AbstractRepository
}

func (t *LoginRepository) FindOneByEmail(email string) chan *models.LoginModel {
	ch := make(chan *models.LoginModel)

	go func() {
		id := (&models.LoginModel{Email: email}).Id()
		res := <-t.FindOneById(id)

		if res.Err != nil {
			logger.Error("Error fetching login info", zap.Error(res.Err))
			ch <- nil
			return
		}
		ch <- res.Value.(*models.LoginModel)
	}()
	return ch
}

func (t *LoginRepository) FindOneByPhone(phone string) chan *models.LoginModel {
	ch := make(chan *models.LoginModel)

	go func() {
		id := (&models.LoginModel{Phone: phone}).Id()
		res := <-t.FindOneById(id)

		if res.Err != nil {
			logger.Error("Error fetching login info", zap.Error(res.Err))
			ch <- nil
			return
		}
		ch <- res.Value.(*models.LoginModel)
	}()
	return ch
}
