package models

type ProfileDeletionModel struct {
	UserId       string `bson:"_id"`
	DeletionTime int64  `bson:"deletionTime"`
	Reason       string `bson:"reason"`
}

func (m *ProfileDeletionModel) Id() string {
	return m.UserId
}
