package db

import (
	odm "github.com/SaiNageswarS/mongo-odm"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProfileRepository struct {
	odm.AbstractRepository
}

func NewProfileRepository(db *mongo.Database) *ProfileRepository {
	return &ProfileRepository{odm.AbstractRepository{db}}
}
