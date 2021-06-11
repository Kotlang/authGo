package service

import (
	"context"
	"encoding/json"

	"github.com/Kotlang/authGo/auth"
	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/jinzhu/copier"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	db *db.AuthDb
}

func NewProfileService(db *db.AuthDb) *ProfileService {
	return &ProfileService{
		db: db,
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

	oldProfile := &models.ProfileModel{}
	oldProfile.LoginId = userId
	oldProfile.Tenant = domain

	<-s.db.Profile.FindOneById(oldProfile)

	// merge old profile and new profile
	newMetadata := copyAll(oldProfile.MetadataMap, getMapFromJson(req.MetaDataMap))
	copier.CopyWithOption(oldProfile, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	oldProfile.MetadataMap = newMetadata

	err := <-s.db.Profile.Save(oldProfile)

	userProfileProto := &pb.UserProfileProto{}
	copier.Copy(userProfileProto, oldProfile)

	return userProfileProto, err
}
