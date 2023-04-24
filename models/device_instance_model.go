package models

type DeviceInstanceModel struct {
	LoginId string `bson:"_id" json:"loginId"`
	Token   string `bson:"token" json:"token"`
}

func (m *DeviceInstanceModel) Id() string {
	return m.LoginId
}
