package models

import (
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson"
)

var loginCollectionNamePrefix string = "login_"

type LoginModel struct {
	IdVal     string `bson:"_id"`
	Email     string `bson:"email"`
	Phone     string `bson:"phone"`
	Otp       string `bson:"otp"`
	UserType  string `bson:"userType"`
	CreatedOn string `bson:"createdOn"`

	//internal field
	Domain string
}

func (m *LoginModel) Id() string {
	if len(m.Email) > 0 {
		m.IdVal = base64.StdEncoding.EncodeToString([]byte(m.Email))
	} else if len(m.Phone) > 0 {
		m.IdVal = base64.StdEncoding.EncodeToString([]byte(m.Phone))
	}
	return m.IdVal
}

func (m *LoginModel) Document() bson.M {
	return bson.M{
		"_id":       m.IdVal,
		"email":     m.Email,
		"phone":     m.Phone,
		"otp":       m.Otp,
		"userType":  m.UserType,
		"createdOn": m.CreatedOn,
	}
}

func (m *LoginModel) Collection() string {
	return loginCollectionNamePrefix + m.Domain
}
