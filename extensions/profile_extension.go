package extensions

import (
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
)

func AttachAttributes(userids *[]models.ProfileModel, userProfileProtos []*pb.UserProfileProto) {
	genderMap := map[string]pb.Gender{
		"Male":        pb.Gender_Male,
		"Female":      pb.Gender_Female,
		"Unspecified": pb.Gender_Unspecified,
	}

	farmingTypeMap := map[string]pb.FarmingType{
		"MIX":      pb.FarmingType_MIX,
		"CHEMICAL": pb.FarmingType_CHEMICAL,
		"ORGANIC":  pb.FarmingType_ORGANIC,
	}

	landSizeMap := map[string]pb.LandSizeInAcres{
		"BETWEEN2AND10": pb.LandSizeInAcres_BETWEEN2AND10,
		"GREATERTHAN10": pb.LandSizeInAcres_GREATERTHAN10,
		"LESSTHAN2":     pb.LandSizeInAcres_LESSTHAN2,
	}

	for i, userprotos := range *userids {
		userProfileProtos[i].Gender = genderMap[userprotos.Gender]
		userProfileProtos[i].FarmingType = farmingTypeMap[userprotos.FarmingType]
		userProfileProtos[i].LandSizeInAcres = landSizeMap[userprotos.LandSizeInAcres]
	}
}
