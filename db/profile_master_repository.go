package db

import (
	"context"

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

func FindByLanguage(ctx context.Context, mongo odm.MongoClient, tenant string, language string) <-chan odm.Result[[]ProfileMasterModel] {
	return odm.CollectionOf[ProfileMasterModel](mongo, tenant).Find(ctx, bson.M{"language": language}, nil, 0, 0)
}
