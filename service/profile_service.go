package service

import (
	"context"
	"encoding/json"

	"github.com/Kotlang/authGo/auth"
	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	profileRepo *db.ProfileRepository
}

func NewProfileService(profileRepo *db.ProfileRepository) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
	}
}

func getNonEmpty(actualVal, defaultVal string) string {
	if len(actualVal) > 0 {
		return actualVal
	}
	return defaultVal
}

func copyAll(oldMap, newMap map[string]interface{}) map[string]interface{} {
	for k, v := range newMap {
		oldMap[k] = v
	}
	return oldMap
}

func getMapFromJson(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

func (s *ProfileService) CreateOrUpdateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.UserProfileProto, error) {
	userId, domain := auth.GetUserIdAndTenant(ctx)

	oldProfile := &models.ProfileModel{
		LoginId: userId,
		Tenant:  domain,
	}
	<-s.profileRepo.FindOneById(oldProfile)

	newMetadata := copyAll(oldProfile.MetadataMap, getMapFromJson(req.MetaDataMap))
	newProfile := &models.ProfileModel{
		LoginId:           userId,
		Tenant:            domain,
		Name:              getNonEmpty(req.Name, oldProfile.Name),
		PhotoUrl:          getNonEmpty(req.PhotoUrl, oldProfile.PhotoUrl),
		Gender:            getNonEmpty(req.Gender, oldProfile.Gender),
		IsVerified:        oldProfile.IsVerified,
		PreferredLanguage: getNonEmpty(req.PreferredLanguage, oldProfile.PreferredLanguage),
		MetadataMap:       newMetadata,
	}

	err := <-s.profileRepo.Save(newProfile)
	return newProfile.GetProto(), err
}
