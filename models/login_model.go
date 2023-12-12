package models

import "github.com/google/uuid"

type LoginModel struct {
	userId               string `bson:"_id"`
	Email                string `bson:"email"`
	Phone                string `bson:"phone"`
	Otp                  string `bson:"otp"`
	UserType             string `bson:"userType"`
	LastOtpSentTime      int64  `bson:"lastOtpSentTime"`
	OtpAuthenticatedTime int64  `bson:"otpAuthenticatedTime"`
	CreatedOn            int64  `bson:"createdOn,omitempty"`
}

func (m *LoginModel) Id() string {
	if m.userId == "" {
		m.userId = uuid.New().String()
	}
	return m.userId
}
