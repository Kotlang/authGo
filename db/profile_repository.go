package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
)

type ProfileRepository struct {
	odm.AbstractRepository[models.ProfileModel]
}

func (p *ProfileRepository) FindByIds(ids []string) (chan []models.ProfileModel, chan error) {
	return p.Find(bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}
