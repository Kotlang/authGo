package models

var tenantCollectionName string = "tenant"

type TenantModel struct {
	Name  string `bson:"_id"`
	Token string `bson:"token"`
	Stage string `bson:"stage"`
}

func (m *TenantModel) Id() string {
	return m.Name
}

func (m *TenantModel) Collection() string {
	return tenantCollectionName
}
