package db

import (
	"context"
	"encoding/base64"

	"github.com/Kotlang/authGo/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

var loginCollectionNamePrefix string = "login_"

type LoginDto struct {
	db *AuthDb
}

type LoginModel struct {
	Id          string `bson:"_id"`
	DomainEmail string `bson:"domainEmail"`
	DomainPhone string `bson:"domainPhone"`
	Otp         string `bson:"otp"`
	UserType    string `bson:"userType"`
	CreatedOn   string `bson:"createdOn"`
}

func NewLoginDto(db *AuthDb) *LoginDto {
	return &LoginDto{
		db: db,
	}
}

func GetLoginId(domainEmailOrPhone string) string {
	uid := base64.StdEncoding.EncodeToString([]byte(domainEmailOrPhone))
	return uid
}

func (t *LoginDto) FindOneByDomainEmail(domain string, email string) chan *LoginModel {
	ch := make(chan *LoginModel)

	go func() {
		domainEmail := domain + "\\" + email
		collection := t.db.Db.Collection(loginCollectionNamePrefix + domain)

		id := GetLoginId(domainEmail)
		loginBson := collection.FindOne(context.Background(), bson.M{"_id": id})

		if loginBson.Err() != nil {
			logger.Error("Error fetching tenant", zap.Error(loginBson.Err()))
			ch <- nil
			return
		}
		login := &LoginModel{}
		loginBson.Decode(login)
		ch <- login
	}()
	return ch
}
