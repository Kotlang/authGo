package models

import (
	"encoding/json"

	pb "github.com/Kotlang/authGo/generated"
	"go.mongodb.org/mongo-driver/bson"
)

var profileCollectionNamePrefix string = "profile_"

type ProfileModel struct {
	LoginId           string                 `bson:"_id"`
	Name              string                 `bson:"name"`
	PhotoUrl          string                 `bson:"photoUrl"`
	Gender            string                 `bson:"gender"`
	IsVerified        bool                   `bson:"isVerified"`
	PreferredLanguage string                 `bson:"preferredLanguage"`
	MetadataMap       map[string]interface{} `bson:"metadata"`
	CreatedOn         string                 `bson:"createdOn"`

	//internal field
	Tenant string
}

func (m *ProfileModel) Id() string {
	return m.LoginId
}

func (m *ProfileModel) Document() bson.M {
	return bson.M{
		"_id":               m.Id(),
		"name":              m.Name,
		"photoUrl":          m.PhotoUrl,
		"gender":            m.Gender,
		"isVerified":        m.IsVerified,
		"preferredLanguage": m.PreferredLanguage,
		"metadata":          bson.M(m.MetadataMap),
		"createdOn":         m.CreatedOn,
	}
}

func (m *ProfileModel) Collection() string {
	return profileCollectionNamePrefix + m.Tenant
}

func (m *ProfileModel) GetProto() *pb.UserProfileProto {
	metadataJsonString, _ := json.Marshal(m.MetadataMap)

	return &pb.UserProfileProto{
		Id:                m.LoginId,
		Name:              m.Name,
		PhotoUrl:          m.PhotoUrl,
		Gender:            m.Gender,
		IsVerified:        m.IsVerified,
		MetaDataMap:       string(metadataJsonString),
		PreferredLanguage: m.PreferredLanguage,
		Domain:            m.Tenant,
	}
}
