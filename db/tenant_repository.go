// Package db provides database functionalities and access methods related to tenant information.
package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// TenantRepository represents a repository to access the tenant collection in the database.
// It offers methods to retrieve tenant data based on specific criteria.
type TenantRepository struct {
	// Embedding the abstract repository for the TenantModel.
	odm.AbstractRepository[models.TenantModel]
}

// FindOneByToken asynchronously fetches a single tenant record based on the provided token.
// It returns a channel that will send the fetched TenantModel or remain empty in case of errors.
// Returns a channel containing a single *models.TenantModel.
func (t *TenantRepository) FindOneByToken(token string) chan *models.TenantModel {
	ch := make(chan *models.TenantModel)

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
