package db

import (
	odm "github.com/SaiNageswarS/mongo-odm"
)

type ProfileRepository struct {
	odm.AbstractRepository
}

func NewProfileRepository(db *AuthDb) *ProfileRepository {
	return &ProfileRepository{odm.AbstractRepository{db.Db}}
}
