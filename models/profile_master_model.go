package models

import "encoding/base64"

type ProfileMasterModel struct {
	Language string   `bson:"language"`
	Field    string   `bson:"field"`
	Type     string   `bson:"type"`
	Options  []string `bson:"options"`
}

func (m *ProfileMasterModel) Id() string {
	return base64.StdEncoding.EncodeToString([]byte(m.Language + "/" + m.Field))
}
