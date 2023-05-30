package models

type ProfileModel struct {
	LoginId           string                 `bson:"_id" json:"loginId"`
	Name              string                 `bson:"name,omitempty" json:"name"`
	PhotoUrl          string                 `bson:"photoUrl" json:"photoUrl"`
	Gender            string                 `bson:"gender" json:"gender" copier:"-"`
	IsVerified        bool                   `bson:"isVerified" json:"isVerified"`
	PreferredLanguage string                 `bson:"preferredLanguage" json:"preferredLanguage"`
	MetadataMap       map[string]interface{} `bson:"metadata"`
	CreatedOn         int64                  `bson:"createdOn,omitempty" json:"createdOn"`
}

func (m *ProfileModel) Id() string {
	return m.LoginId
}
