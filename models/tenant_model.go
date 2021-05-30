package models

import "go.mongodb.org/mongo-driver/bson"

var tenantCollectionName string = "tenant"

type TenantModel struct {
	Name  string `bson:"_id"`
	Token string `bson:"token"`
	Stage string `bson:"stage"`
}

func (m *TenantModel) Id() string {
	return m.Name
}

func (m *TenantModel) Document() bson.M {
	return bson.M{
		"_id":   m.Name,
		"token": m.Token,
		"Stage": m.Stage,
	}
}

func (m *TenantModel) Collection() string {
	return tenantCollectionName
}
