package db

import (
	"context"

	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/SaiNageswarS/go-api-boot/async"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type CertificateModel struct {
	IsCertified         bool   `bson:"isCertified" json:"isCertified"`
	CertificationId     string `bson:"certificateId" json:"certificateId"`
	CertificationName   string `bson:"certificateName" json:"certificateName"`
	CertificationAgency string `bson:"certificationAgency" json:"certificationAgency"`
}
type Location struct {
	Lat  float64 `bson:"lat" json:"lat"`
	Long float64 `bson:"long" json:"long"`
}

type Addresses struct {
	Type    string `bson:"type" json:"type"`
	Address string `bson:"address" json:"address"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	Country string `bson:"country" json:"country"`
}

type DeletionInfo struct {
	MarkedForDeletion bool   `bson:"markedForDeletion" json:"markedForDeletion"`
	DeletionTime      int64  `bson:"deletionTime" json:"deletionTime"`
	Reason            string `bson:"reason" json:"reason"`
}

type ProfileModel struct {
	UserId                   string           `bson:"_id" json:"userId"`
	Name                     string           `bson:"name,omitempty" json:"name"`
	PhotoUrl                 string           `bson:"photoUrl" json:"photoUrl"`
	Addresses                []Addresses      `bson:"addresses" json:"addresses"`
	Location                 Location         `bson:"location" json:"location"`
	FarmingType              string           `bson:"farmingType" json:"farmingType"`
	Bio                      string           `bson:"bio" json:"bio"`
	Crops                    []string         `bson:"crops" json:"crops"`
	YearsSinceOrganicFarming int              `bson:"yearsSinceOrganicFarming" json:"yearsSinceOrganicFarming"`
	Gender                   string           `bson:"gender" json:"gender" copier:"-"`
	IsVerified               bool             `bson:"isVerified" json:"isVerified"`
	PreferredLanguage        string           `bson:"preferredLanguage" json:"preferredLanguage"`
	CertificationDetails     CertificateModel `bson:"certificationDetails" json:"certificationDetails"`
	CreatedOn                int64            `bson:"createdOn,omitempty" json:"createdOn"`
	LandSizeInAcres          string           `bson:"landSizeInAcres" json:"landSizeInAcres"`
}

func (m ProfileModel) Id() string {
	return m.UserId
}

func (m ProfileModel) CollectionName() string { return "profiles" }

func FindProfilesByIds(ctx context.Context, mongo odm.MongoClient, tenant string, ids []string) <-chan async.Result[[]ProfileModel] {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	return odm.CollectionOf[ProfileModel](mongo, tenant).Find(ctx, filter, nil, int64(len(ids)), 0)
}

func GetProfiles(ctx context.Context, mongo odm.MongoClient, tenant string, userfilters *authPb.Userfilters, PageSize, PageNumber int64) (profiles []ProfileModel, totalCount int) {
	filters := bson.M{}

	if name := userfilters.Name; name != "" {
		filters["name"] = userfilters.Name
	}
	if gender := userfilters.Gender.String(); gender != authPb.Gender_Unspecified.String() {
		filters["gender"] = userfilters.Gender.String()
	}
	if farmingType := userfilters.FarmingType.String(); farmingType != authPb.FarmingType_UnspecifiedFarming.String() {
		filters["farmingType"] = userfilters.FarmingType.String()
	}
	if land := userfilters.LandSizeInAcres.String(); land != authPb.LandSizeInAcres_UnspecifiedLandSize.String() {
		filters["landSizeInAcres"] = userfilters.LandSizeInAcres.String()
	}
	if year := userfilters.YearsSinceOrganicFarming; year > 0 {
		filters["yearsSinceOrganicFarming"] = userfilters.YearsSinceOrganicFarming
	}

	skip := PageNumber * PageSize

	resultChan := odm.CollectionOf[ProfileModel](mongo, tenant).Find(ctx, filters, nil, PageSize, skip)
	totalCountResChan := odm.CollectionOf[ProfileModel](mongo, tenant).Count(ctx, filters)
	totalCount = 0

	count, err := async.Await(totalCountResChan)
	if err != nil {
		logger.Error("Error fetching total count of profiles", zap.Error(err))
	} else {
		totalCount = int(count)
	}

	profiles, err = async.Await(resultChan)
	if err != nil {
		logger.Error("Error fetching profiles", zap.Error(err))
		return []ProfileModel{}, totalCount
	}

	return profiles, totalCount
}
