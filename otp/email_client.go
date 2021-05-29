package otp

import "regexp"

type EmailClient struct{}

func NewEmailClient() *EmailClient {
	return &EmailClient{}
}

func (c *EmailClient) IsValidEmail(emailOrPhone string) bool {
	match, _ := regexp.MatchString("^(.+)@(.+)\\$", emailOrPhone)
	return match
}

func (c *EmailClient) SendOtp() {}
