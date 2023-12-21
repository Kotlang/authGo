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

	if name := userfilters.Name; name != "" {
		filters["name"] = userfilters.Name
	}
	if gender := userfilters.Gender.String(); gender != "Male" {
		filters["gender"] = userfilters.Gender.String()
	}
	if farmingType := userfilters.FarmingType.String(); farmingType != "ORGANIC" {
		filters["farmingType"] = userfilters.FarmingType.String()
	}
	if land := userfilters.LandSizeInAcres.String(); land != "LESSTHAN2" {
		filters["landSizeInAcres"] = userfilters.LandSizeInAcres.String()
	}
	if year := userfilters.YearsSinceOrganicFarming; year > 0 {
		filters["yearsSinceOrganicFarming"] = userfilters.YearsSinceOrganicFarming
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
