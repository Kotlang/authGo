package db

import (
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type ProfileRepositoryInterface interface {
	odm.BootRepository[models.ProfileModel]
	FindByIds(ids []string) (chan []models.ProfileModel, chan error)
	GetProfiles(userfilters *authPb.Userfilters, PageSize, PageNumber int64) ([]models.ProfileModel, int)
}

type ProfileRepository struct {
	odm.UnimplementedBootRepository[models.ProfileModel]
}

func (p *ProfileRepository) FindByIds(ids []string) (chan []models.ProfileModel, chan error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	return p.Find(filter, nil, int64(len(ids)), 0)
}

func (t *ProfileRepository) GetProfiles(userfilters *authPb.Userfilters, PageSize, PageNumber int64) (profiles []models.ProfileModel, totalCount int) {
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

	resultChan, errChan := t.Find(filters, nil, PageSize, skip)
	totalCountResChan, countErrChan := t.CountDocuments(filters)
	totalCount = 0

	select {
	case count := <-totalCountResChan:
		totalCount = int(count)
	case err := <-countErrChan:
		logger.Error("Error fetching user count", zap.Error(err))
	}

	select {
	case res := <-resultChan:
		return res, totalCount
	case err := <-errChan:
		logger.Error("Error fetching user IDs", zap.Error(err))
		return []models.ProfileModel{}, totalCount
	}
}
