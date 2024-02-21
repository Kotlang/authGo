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

type Addresses struct {
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
	UserId                   string               `bson:"_id" json:"userId"`
	Name                     string               `bson:"name,omitempty" json:"name"`
	PhotoUrl                 string               `bson:"photoUrl" json:"photoUrl"`
	Addresses                map[string]Addresses `bson:"addresses" json:"addresses"`
	Location                 Location             `bson:"location" json:"location"`
	FarmingType              string               `bson:"farmingType" json:"farmingType"`
	Bio                      string               `bson:"bio" json:"bio"`
	Crops                    []string             `bson:"crops" json:"crops"`
	YearsSinceOrganicFarming int                  `bson:"yearsSinceOrganicFarming" json:"yearsSinceOrganicFarming"`
	Gender                   string               `bson:"gender" json:"gender" copier:"-"`
	IsVerified               bool                 `bson:"isVerified" json:"isVerified"`
	PreferredLanguage        string               `bson:"preferredLanguage" json:"preferredLanguage"`
	CertificationDetails     CertificateModel     `bson:"certificationDetails" json:"certificationDetails"`
	CreatedOn                int64                `bson:"createdOn,omitempty" json:"createdOn"`
	LandSizeInAcres          string               `bson:"landSizeInAcres" json:"landSizeInAcres"`
	DeletionInfo             DeletionInfo         `bson:"deletionInfo" json:"deletionInfo"`
	IsBlocked                bool                 `bson:"isBlocked" json:"isBlocked"`
}

func (m *ProfileModel) Id() string {
	return m.UserId
}
