package db

import (
	"context"

	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type LoginModel struct {
	UserId               string       `bson:"_id"`
	Email                string       `bson:"email"`
	Phone                string       `bson:"phone"`
	Otp                  string       `bson:"otp"`
	UserType             string       `bson:"userType"`
	LastOtpSentTime      int64        `bson:"lastOtpSentTime"`
	OtpAuthenticatedTime int64        `bson:"otpAuthenticatedTime"`
	CreatedOn            int64        `bson:"createdOn,omitempty"`
	DeletionInfo         DeletionInfo `bson:"deletionInfo" json:"deletionInfo"`
	IsBlocked            bool         `bson:"isBlocked" json:"isBlocked"`
	LastActive           int64        `bson:"lastActive" json:"lastActive"`
}

func (m LoginModel) Id() string {
	if m.UserId == "" {
		m.UserId = uuid.New().String()
	}
	return m.UserId
}

func (m LoginModel) CollectionName() string { return "login" }

func FindOneByPhoneOrEmail(ctx context.Context, mongo odm.MongoClient, tenant, phone, email string) chan *LoginModel {
	ch := make(chan *LoginModel)

	go func() {
		filter := bson.M{}

		if len(phone) > 0 {
			filter["phone"] = phone
		} else {
			filter["email"] = email
		}

		result, error := odm.Await(odm.CollectionOf[LoginModel](mongo, tenant).FindOne(ctx, filter))
		if error != nil {
			ch <- nil
		} else {
			ch <- result
		}
	}()
	return ch
}

func FindLoginsByIds(ctx context.Context, mongo odm.MongoClient, tenant string, ids []string) <-chan odm.Result[[]LoginModel] {
	return odm.CollectionOf[LoginModel](mongo, tenant).Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}

func IsAdmin(mongo odm.MongoClient, tenant, id string) bool {
	loginInfo, err := odm.Await(odm.CollectionOf[LoginModel](mongo, tenant).FindOneByID(context.Background(), id))
	if err != nil {
		return false
	}

	return loginInfo.UserType == "admin"
}
