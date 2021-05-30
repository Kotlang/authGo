package models

var profileCollectionNamePrefix string = "profile_"

type ProfileModel struct {
	LoginId           string            `bson:"_id"`
	Name              string            `bson:"name"`
	PhotoUrl          string            `bson:"photoUrl"`
	Gender            string            `bson:"gender"`
	IsVerified        bool              `bson:"isVerified"`
	PreferredLanguage string            `bson:"preferredLanguage"`
	MetadataMap       map[string]string `bson:"metadata"`
	CreatedOn         string            `bson:"createdOn"`
}
