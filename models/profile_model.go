package models

var profileCollectionNamePrefix string = "profile_"

type ProfileModel struct {
	LoginId           string                 `bson:"_id" json:"loginId"`
	Name              string                 `bson:"name" json:"name"`
	PhotoUrl          string                 `bson:"photoUrl" json:"photoUrl"`
	Gender            string                 `bson:"gender" json:"gender"`
	IsVerified        bool                   `bson:"isVerified" json:"isVerified"`
	PreferredLanguage string                 `bson:"preferredLanguage" json:"preferredLanguage"`
	MetadataMap       map[string]interface{} `bson:"metadata"`
	CreatedOn         int64                  `bson:"createdOn" json:"createdOn"`

	//internal field
	Tenant string
}

func (m *ProfileModel) Id() string {
	return m.LoginId
}

func (m *ProfileModel) Collection() string {
	return profileCollectionNamePrefix + m.Tenant
}
