package models

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

type ProfileModel struct {
	LoginId                  string                 `bson:"_id" json:"loginId"`
	Name                     string                 `bson:"name,omitempty" json:"name"`
	PhotoUrl                 string                 `bson:"photoUrl" json:"photoUrl"`
	Addresses                map[string]interface{} `bson:"addresses" json:"addresses"`
	Location                 Location               `bson:"location" json:"location"`
	FarmingType              string                 `bson:"farmingType" json:"farmingType"`
	Bio                      string                 `bson:"bio" json:"bio"`
	Crops                    []string               `bson:"crops" json:"crops"`
	YearsSinceOrganicFarming int                    `bson:"yearsSinceOrganicFarming" json:"yearsSinceOrganicFarming"`
	Gender                   string                 `bson:"gender" json:"gender" copier:"-"`
	IsVerified               bool                   `bson:"isVerified" json:"isVerified"`
	PreferredLanguage        string                 `bson:"preferredLanguage" json:"preferredLanguage"`
	MetadataMap              map[string]interface{} `bson:"metadata"`
	CertificationDetails     CertificateModel       `bson:"certificationDetails" json:"certificationDetails"`
	CreatedOn                int64                  `bson:"createdOn,omitempty" json:"createdOn"`
}

func (m *ProfileModel) Id() string {
	return m.LoginId
}
