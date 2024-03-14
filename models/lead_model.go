package models

import "github.com/google/uuid"

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
