package db

import (
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
)

type ProfileMasterModel struct {
	Language string   `bson:"language"`
	Field    string   `bson:"field"`
	Type     string   `bson:"type"`
	Options  []string `bson:"options"`
}

func (m *ProfileMasterModel) Id() string {
	return m.Language + "/" + m.Field
}

type ProfileMasterRepositoryInterface interface {
	odm.BootRepository[ProfileMasterModel]
	FindByLanguage(language string) (chan []ProfileMasterModel, chan error)
}

type ProfileMasterRepository struct {
	odm.UnimplementedBootRepository[ProfileMasterModel]
}

func (p *ProfileMasterRepository) FindByLanguage(language string) (chan []ProfileMasterModel, chan error) {
	return p.Find(bson.M{"language": language}, nil, 0, 0)
}
