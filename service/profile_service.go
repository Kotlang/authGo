package service

import (
	"context"
	"encoding/json"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
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
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginResChannel := s.db.Login(tenant).FindOneById(userId)
	profileResChannel := s.db.Profile(tenant).FindOneById(userId)

	oldProfile := (<-profileResChannel).Value.(*models.ProfileModel)
	// merge old profile and new profile
	newMetadata := copyAll(oldProfile.MetadataMap, getMapFromJson(req.MetaDataMap))
	copier.CopyWithOption(oldProfile, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	oldProfile.MetadataMap = newMetadata

	err := <-s.db.Profile(tenant).Save(oldProfile)

	userProfileProto := &pb.UserProfileProto{}
	copier.Copy(userProfileProto, oldProfile)

	loginInfo := (<-loginResChannel).Value.(*models.LoginModel)
	copier.CopyWithOption(userProfileProto, loginInfo, copier.Option{IgnoreEmpty: true})

	return userProfileProto, err
}
