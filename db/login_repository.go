// Package db provides database functionalities and access methods related to login information.
package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// LoginRepository represents a repository to access the login collection in the database.
// It provides various methods to fetch login data based on different criteria.
type LoginRepository struct {
	// Embedding the abstract repository for the LoginModel.
	odm.AbstractRepository[models.LoginModel]
}

// FindOneByEmail fetches a single login record based on the provided email.
// It performs an asynchronous operation and returns a channel that will
// send the fetched LoginModel or remain empty in case of errors.
// Returns a channel containing a single *models.LoginModel.
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

// FindOneByPhone fetches a single login record based on the provided phone number.
// It performs an asynchronous operation and returns a channel that will
// send the fetched LoginModel or nil in case of errors.
// Returns a channel containing a single *models.LoginModel.
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

// FindByIds fetches multiple login records based on the provided list of IDs.
// It returns two channels: one for the list of fetched LoginModels and another for errors.
// Returns a channel containing a slice of models.LoginModel and an error channel.
func (t *LoginRepository) FindByIds(ids []string) (chan []models.LoginModel, chan error) {
	return t.Find(bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}
