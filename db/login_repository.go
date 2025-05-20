package db

import (
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
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

func FindOneByPhoneOrEmail(mongo odm.MongoClient, tenant, phone, email string) chan *LoginModel {
	ch := make(chan *LoginModel)

	go func() {
		filter := bson.M{}

		if len(phone) > 0 {
			filter["phone"] = phone
		} else {
			filter["email"] = email
		}

		resultChan, errorChan := odm.CollectionOf[LoginModel](mongo, tenant).FindOne(filter)

		select {
		case res := <-resultChan:
			ch <- res
		case err := <-errorChan:
			logger.Error("Error fetching login info", zap.Error(err))
			ch <- nil
		}
	}()
	return ch
}

func FindLoginsByIds(mongo odm.MongoClient, tenant string, ids []string) (chan []LoginModel, chan error) {
	return odm.CollectionOf[LoginModel](mongo, tenant).Find(bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}

func IsAdmin(mongo odm.MongoClient, tenant, id string) bool {
	loginInfoChan, errResChan := odm.CollectionOf[LoginModel](mongo, tenant).FindOneById(id)

	//get login info using userId
	select {
	case loginInfo := <-loginInfoChan:
		return loginInfo.UserType == "admin"
	case err := <-errResChan:
		logger.Error("Error fetching login info", zap.Error(err))
		return false
	}
}
