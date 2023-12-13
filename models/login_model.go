package models

import "github.com/google/uuid"

type LoginModel struct {
	UserId               string `bson:"_id"`
	Email                string `bson:"email"`
	Phone                string `bson:"phone"`
	Otp                  string `bson:"otp"`
	UserType             string `bson:"userType"`
	LastOtpSentTime      int64  `bson:"lastOtpSentTime"`
	OtpAuthenticatedTime int64  `bson:"otpAuthenticatedTime"`
	CreatedOn            int64  `bson:"createdOn,omitempty"`
}

func (m *LoginModel) Id() string {
	if m.UserId == "" {
		m.UserId = uuid.New().String()
	}
	return m.UserId
}
