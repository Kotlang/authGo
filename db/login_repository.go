package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type LoginRepositoryInterface interface {
	odm.BootRepository[models.LoginModel]
	FindByIds(ids []string) (chan []models.LoginModel, chan error)
	FindOneByPhoneOrEmail(phone, email string) chan *models.LoginModel
	IsAdmin(id string) bool
}

type LoginRepository struct {
	odm.UnimplementedBootRepository[models.LoginModel]
}

func (t *LoginRepository) FindOneByPhoneOrEmail(phone, email string) chan *models.LoginModel {
	ch := make(chan *models.LoginModel)

	go func() {
		filter := bson.M{}

		if len(phone) > 0 {
			filter["phone"] = phone
		} else {
			filter["email"] = email
		}

		resultChan, errorChan := t.FindOne(filter)

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

func (t *LoginRepository) IsAdmin(id string) bool {
	loginInfoChan, errResChan := t.FindOneById(id)

	//get login info using userId
	select {
	case loginInfo := <-loginInfoChan:
		return loginInfo.UserType == "admin"
	case err := <-errResChan:
		logger.Error("Error fetching login info", zap.Error(err))
		return false
	}
}
