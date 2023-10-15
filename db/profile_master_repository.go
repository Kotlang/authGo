// Package db provides database functionalities and access methods related to profile master information.
package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
)

// ProfileMasterRepository represents a repository to access the profile master collection in the database.
// It provides methods to fetch profile master data based on different criteria.
type ProfileMasterRepository struct {
	// Embedding the abstract repository for the ProfileMasterModel.
	odm.AbstractRepository[models.ProfileMasterModel]
}

// FindByLanguage fetches profile master records based on the provided language.
// Returns two channels: one for the list of fetched ProfileMasterModels and another for errors.
func (p *ProfileMasterRepository) FindByLanguage(language string) (chan []models.ProfileMasterModel, chan error) {
	return p.Find(bson.M{"language": language}, nil, 0, 0)
}
