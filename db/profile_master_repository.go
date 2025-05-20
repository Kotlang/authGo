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

func (m ProfileMasterModel) Id() string {
	return m.Language + "/" + m.Field
}

func (m ProfileMasterModel) CollectionName() string { return "profile_master" }

func FindByLanguage(mongo odm.MongoClient, tenant string, language string) (chan []ProfileMasterModel, chan error) {
	return odm.CollectionOf[ProfileMasterModel](mongo, tenant).Find(bson.M{"language": language}, nil, 0, 0)
}
