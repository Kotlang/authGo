package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
)

type ProfileMasterRepository struct {
	odm.AbstractRepository[models.ProfileMasterModel]
}

func (p *ProfileMasterRepository) FindByLanguage(language string) (chan []models.ProfileMasterModel, chan error) {
	return p.Find(bson.M{"language": language}, nil, 0, 0)
}
