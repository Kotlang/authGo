package otp

import (
	"regexp"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/logger"
	"github.com/Kotlang/authGo/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmailClient struct {
	db *db.AuthDb
}

func NewEmailClient(db *db.AuthDb) *EmailClient {
	return &EmailClient{
		db: db,
	}
}

func (c *EmailClient) IsValidEmail(emailOrPhone string) bool {
	match, _ := regexp.MatchString("^(.+)@(.+)$", emailOrPhone)
	return match
}

func (c *EmailClient) SendOtp(tenant, emailId string) {
	loginInfo := c.getOrCreateByEmail(tenant, emailId)
	logger.Info("Fetched login info is", zap.Any("loginInfo", loginInfo))
	loginInfo.Otp = "1214"
	<-c.db.Login.Save(loginInfo)
}

func (c *EmailClient) getOrCreateByEmail(tenant, emailId string) *models.LoginModel {
	loginInfo := <-c.db.Login.FindOneByEmail(tenant, emailId)
	if loginInfo == nil {
		loginInfo = &models.LoginModel{
			Email:     emailId,
			CreatedOn: time.Now().Format("UnixDate"),
			UserType:  "default",
			Tenant:    tenant,
		}
		<-c.db.Login.Save(loginInfo)
	}
	return loginInfo
}

func (c *EmailClient) ValidateOtpAndGetLoginInfo(tenant, emailId, otp string) (*models.LoginModel, error) {
	loginInfo := c.getOrCreateByEmail(tenant, emailId)
	if otp != loginInfo.Otp {
		return nil, status.Error(codes.PermissionDenied, "Wrong OTP")
	}
	return loginInfo, nil
}
