package models

import (
	"encoding/base64"
)

var loginCollectionNamePrefix string = "login_"

type LoginModel struct {
	Email     string `bson:"email"`
	Phone     string `bson:"phone"`
	Otp       string `bson:"otp"`
	UserType  string `bson:"userType"`
	CreatedOn string `bson:"createdOn"`

	//internal field
	Tenant string
}

func (m *LoginModel) Id() string {
	if len(m.Email) > 0 {
		return base64.StdEncoding.EncodeToString([]byte(m.Email))
	} else {
		return base64.StdEncoding.EncodeToString([]byte(m.Phone))
	}
}

func (m *LoginModel) Collection() string {
	return loginCollectionNamePrefix + m.Tenant
}
