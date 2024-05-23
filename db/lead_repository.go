package db

import (
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type LeadModel struct {
	LeadId               string           `bson:"_id"`
	Name                 string           `bson:"name"`
	PhoneNumber          string           `bson:"phoneNumber"`
	OperatorType         string           `bson:"operatorType"`
	Channel              string           `bson:"channel"`
	Source               string           `bson:"source"`
	Addresses            []Addresses      `bson:"addresses"`
	LandSizeInAcres      string           `bson:"landSizeInAcres"`
	FarmingType          string           `bson:"farmingType"`
	CertificationDetails CertificateModel `bson:"certificationDetails"`
	Crops                []string         `bson:"crops"`
	MainProfession       string           `bson:"mainProfession"`
	OrganizationName     string           `bson:"organizationName"`
	SideProfession       string           `bson:"sideProfession"`
	UserInterviewNotes   string           `bson:"userInterviewNotes"`
	Education            string           `bson:"education"`
}

func (m *LeadModel) Id() string {
	if m.LeadId == "" {
		m.LeadId = uuid.New().String()
	}

	return m.LeadId
}

type LeadRepositoryInterface interface {
	odm.BootRepository[LeadModel]
	FindByIds(ids []string) (chan []LeadModel, chan error)
	GetLeads(leadFilters *authPb.LeadFilters, PageSize, PageNumber int64) (leads []LeadModel, totalCount int)
}

type LeadRepository struct {
	odm.UnimplementedBootRepository[LeadModel]
}

func (l *LeadRepository) FindByIds(ids []string) (chan []LeadModel, chan error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	return l.Find(filter, nil, 0, 0)
}

func (l *LeadRepository) GetLeads(leadFilters *authPb.LeadFilters, PageSize, PageNumber int64) (leads []LeadModel, totalCount int) {

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

func getLeadFilter(leadFilters *authPb.LeadFilters) bson.M {

	if leadFilters == nil {
		return bson.M{}
	}

	filter := bson.M{}

	// if operator type is not unspecified then add it to filter
	if leadFilters.OperatorType != authPb.OperatorType_UNSPECIFIED_OPERATOR {
		filter["operatorType"] = leadFilters.OperatorType.String()
	}

	// if channel is not unspecified then add it to filter
	if leadFilters.Channel != authPb.LeadChannel_UNSPECIFIED_CHANNEL {
		filter["channel"] = leadFilters.Channel.String()
	}

	if leadFilters.Source != "" {
		filter["source"] = leadFilters.Source
	}

	if leadFilters.LandSizeInAcres != authPb.LandSizeInAcres_UnspecifiedLandSize {
		filter["landSizeInAcres"] = leadFilters.LandSizeInAcres.String()
	}

	if leadFilters.FarmingType != authPb.FarmingType_UnspecifiedFarming {
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

	if leadFilters.Status != authPb.Status_UNSPECIFIED_STATUS {
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

func getAddressFilter(addressfilter *authPb.AddressFilters) bson.M {
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
