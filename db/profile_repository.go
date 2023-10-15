// Package db provides database functionalities and access methods related to profile information.
package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
)

// ProfileRepository represents a repository to access the profile collection in the database.
// It offers methods to retrieve profile data based on various criteria.
type ProfileRepository struct {
	// Embedding the abstract repository for the ProfileModel.
	odm.AbstractRepository[models.ProfileModel]
}

// FindByIds fetches profile records based on a provided list of IDs.
// It returns two channels: one for the list of retrieved ProfileModels and another for errors.
func (p *ProfileRepository) FindByIds(ids []string) (chan []models.ProfileModel, chan error) {
	return p.Find(bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}
