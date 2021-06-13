package service

import (
	"context"
	"encoding/json"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	s3client "github.com/Kotlang/authGo/s3Client"
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
	profileRes := <-s.db.Profile(tenant).FindOneById(userId)

	var oldProfile *models.ProfileModel
	// old profile doesn't exist
	if profileRes.Err != nil {
		oldProfile = &models.ProfileModel{LoginId: userId}
	} else {
		oldProfile = profileRes.Value.(*models.ProfileModel)
	}

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

func (s *ProfileService) GetProfileById(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}
	profile := s.db.Profile(tenant).FindOneById(userId)
	loginInfo := s.db.Login(tenant).FindOneById(userId)

	profileProto := &pb.UserProfileProto{}

	copier.Copy(profileProto, (<-profile).Value)
	copier.CopyWithOption(profileProto, (<-loginInfo).Value, copier.Option{IgnoreEmpty: true})
	return profileProto, nil
}

func (s *ProfileService) GetProfileImageUploadUrl(ctx context.Context, req *pb.ProfileImageUploadRequest) (*pb.ProfileImageUploadURL, error) {
	uploadInstructions := `
	| 1. Send profile image file to above uploadURL as a PUT request. 
	| 
	| curl --location --request PUT '<aboveURL>' 
	|      --header 'Content-Type: image/jpeg' 
	|      --data-binary '@/path/to/file.jpg'
	|      
	| 2. Send mediaUrl in createOrUpdateProfile request.`

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	preSignedUrl, downloadUrl := s3client.GetPresignedUrlForProfilePic(tenant, userId, req.MediaExtension)
	return &pb.ProfileImageUploadURL{
		UploadUrl:    preSignedUrl,
		MediaUrl:     downloadUrl,
		Instructions: uploadInstructions,
	}, nil
}
