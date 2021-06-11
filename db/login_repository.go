package db

import (
	"github.com/Kotlang/authGo/logger"
	"github.com/Kotlang/authGo/models"
	odm "github.com/SaiNageswarS/mongo-odm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type LoginRepository struct {
	odm.AbstractRepository
}

func NewLoginRepository(db *mongo.Database) *LoginRepository {
	return &LoginRepository{
		odm.AbstractRepository{db},
	}
}

func (t *LoginRepository) FindOneByEmail(domain string, email string) chan *models.LoginModel {
	ch := make(chan *models.LoginModel)

	go func() {
		login := &models.LoginModel{
			Email:  email,
			Tenant: domain,
		}
		err := <-t.FindOneById(login)

		if err != nil {
			logger.Error("Error fetching login info", zap.Error(err))
			ch <- nil
			return
		}
		ch <- login
	}()
	return ch
}
