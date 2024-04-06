package db

import (
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type LeadRepositoryInterface interface {
	odm.BootRepository[models.LeadModel]
	FindByIds(ids []string) (chan []models.LeadModel, chan error)
	GetLeads(leadFilters *pb.LeadFilters, PageSize, PageNumber int64) (leads []models.LeadModel, totalCount int)
}

type LeadRepository struct {
	odm.UnimplementedBootRepository[models.LeadModel]
}

func (l *LeadRepository) FindByIds(ids []string) (chan []models.LeadModel, chan error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	return l.Find(filter, nil, 0, 0)
}

func (l *LeadRepository) GetLeads(leadFilters *pb.LeadFilters, PageSize, PageNumber int64) (leads []models.LeadModel, totalCount int) {

	// get the filter
	filter := getLeadFilter(leadFilters)

	// get the leads and total count
	skip := PageNumber * PageSize

	// sort by created at
	sort := bson.D{
		{Key: "createdAt", Value: -1},
	}

	// get the leads
	resultChan, errChan := l.Find(filter, sort, PageSize, skip)
	totalCountResChan, countErrChan := l.CountDocuments(filter)
	totalCount = 0

	select {
	case count := <-totalCountResChan:
		totalCount = int(count)
	case err := <-countErrChan:
		logger.Error("Error fetching lead count", zap.Error(err))
	}

	select {
	case res := <-resultChan:
		leads = res
	case err := <-errChan:
		logger.Error("Error fetching leads", zap.Error(err))
	}

	return leads, totalCount
}

func getLeadFilter(leadFilters *pb.LeadFilters) bson.M {

	if leadFilters == nil {
		return bson.M{}
	}

	filter := bson.M{}

	// if operator type is not unspecified then add it to filter
	if leadFilters.OperatorType != pb.OperatorType_UNSPECIFIED_OPERATOR {
		filter["operatorType"] = leadFilters.OperatorType.String()
	}

	// if channel is not unspecified then add it to filter
	if leadFilters.Channel != pb.LeadChannel_UNSPECIFIED_CHANNEL {
		filter["channel"] = leadFilters.Channel.String()
	}

	if leadFilters.Source != "" {
		filter["source"] = leadFilters.Source
	}

	if leadFilters.LandSizeInAcres != pb.LandSizeInAcres_UnspecifiedLandSize {
		filter["landSizeInAcres"] = leadFilters.LandSizeInAcres.String()
	}

	if leadFilters.FarmingType != pb.FarmingType_UnspecifiedFarming {
		filter["farmingType"] = leadFilters.FarmingType.String()
	}

	// if certification details are not nil then add it to filter if not empty
	if leadFilters.CertificationDetails != nil {

		filter["certificationDetails.isCertified"] = leadFilters.CertificationDetails.IsCertified

		if leadFilters.CertificationDetails.CertificationAgency != "" {
			filter["certificationDetails.certificationAgency"] = leadFilters.CertificationDetails.CertificationAgency
		}
		if leadFilters.CertificationDetails.CertificationName != "" {
			filter["certificationDetails.certificationName"] = leadFilters.CertificationDetails.CertificationName
		}
	}

	if leadFilters.MainProfession != "" {
		filter["mainProfession"] = leadFilters.MainProfession
	}

	if leadFilters.OrganizationName != "" {
		filter["organizationName"] = leadFilters.OrganizationName
	}

	if leadFilters.SideProfession != "" {
		filter["sideProfession"] = leadFilters.SideProfession
	}

	if leadFilters.Education != "" {
		filter["education"] = leadFilters.Education
	}

	if leadFilters.Status != pb.Status_UNSPECIFIED_STATUS {
		filter["status"] = leadFilters.Status.String()
	}

	if leadFilters.AddressFilters != nil {
		addressFilters := getAddressFilter(leadFilters.AddressFilters)

		if len(addressFilters) > 0 {
			filter["addresses"] = bson.M{
				"$elemMatch": addressFilters,
			}
		}
	}

	if leadFilters.PhoneNumbers != nil {
		filter["phoneNumber"] = bson.M{
			"$in": leadFilters.PhoneNumbers,
		}
	}

	return filter
}

func getAddressFilter(addressfilter *pb.AddressFilters) bson.M {
	filter := bson.M{}
	if addressfilter.City != "" {
		filter["city"] = addressfilter.City
	}
	if addressfilter.State != "" {
		filter["state"] = addressfilter.State
	}
	if addressfilter.Country != "" {
		filter["country"] = addressfilter.Country
	}
	return filter
}
