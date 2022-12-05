package models

type ProfileMasterModel struct {
	Language string   `bson:"language"`
	Field    string   `bson:"field"`
	Type     string   `bson:"type"`
	Options  []string `bson:"options"`
}

func (m *ProfileMasterModel) Id() string {
	return m.Language + "/" + m.Field
}
