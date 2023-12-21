package db

import (
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type ProfileRepositoryInterface interface {
	odm.BootRepository[models.ProfileModel]
	FindByIds(ids []string) (chan []models.ProfileModel, chan error)
	GetUserIds(userfilters *pb.Userfilters, PageSize, PageNumber int64) []models.ProfileModel
}

type ProfileRepository struct {
	odm.UnimplementedBootRepository[models.ProfileModel]
}

func (p *ProfileRepository) FindByIds(ids []string) (chan []models.ProfileModel, chan error) {
	return p.Find(bson.M{"_id": bson.M{"$in": ids}}, nil, int64(len(ids)), 0)
}

func (t *ProfileRepository) GetUserIds(userfilters *pb.Userfilters, PageSize, PageNumber int64) []models.ProfileModel {
	filters := bson.M{}

	if userfilters != nil && len(userfilters.Name) > 0 {
		filters["name"] = userfilters.Name
	}
	if userfilters != nil && userfilters.Gender.String() != "Male" {
		println(userfilters.Gender.String())
		filters["gender"] = bson.M{"$eq": userfilters.Gender.String()}
	}
	if userfilters != nil && userfilters.FarmingType.String() != "ORGANIC" {
		filters["farmingType"] = bson.M{"$eq": userfilters.FarmingType.String()}
	}
	if userfilters != nil && userfilters.LandSizeInAcres.String() != "LESSTHAN2" {
		filters["landSizeInAcres"] = bson.M{"$eq": userfilters.LandSizeInAcres.String()}
	}
	if userfilters != nil && userfilters.YearsSinceOrganicFarming > 0 {
		filters["yearsSinceOrganicFarming"] = bson.M{"$eq": userfilters.YearsSinceOrganicFarming}
	}

	skip := PageNumber * PageSize

	resultChan, errChan := t.Find(filters, nil, PageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Error fetching user IDs", zap.Error(err))
		return []models.ProfileModel{}
	}
}
