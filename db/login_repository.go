package db

import (
	"github.com/Kotlang/authGo/logger"
	"github.com/Kotlang/authGo/models"
	odm "github.com/SaiNageswarS/mongo-odm"
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
