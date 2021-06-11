package models

import (
	"encoding/base64"
)

type LoginModel struct {
	Email     string `bson:"email"`
	Phone     string `bson:"phone"`
	Otp       string `bson:"otp"`
	UserType  string `bson:"userType"`
	CreatedOn int64  `bson:"createdOn"`
}

func (m *LoginModel) Id() string {
	if len(m.Email) > 0 {
		return base64.StdEncoding.EncodeToString([]byte(m.Email))
	} else {
		return base64.StdEncoding.EncodeToString([]byte(m.Phone))
	}
}
