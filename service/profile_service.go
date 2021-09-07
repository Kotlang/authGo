package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/aws"
	"github.com/SaiNageswarS/go-api-boot/azure"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
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
	key := fmt.Sprintf("%s/%s/%d.jpg", tenant, userId, time.Now().Unix())
	preSignedUrl, downloadUrl := aws.S3.GetPresignedUrl("kotlang-assets", key)
	return &pb.ProfileImageUploadURL{
		UploadUrl:    preSignedUrl,
		MediaUrl:     downloadUrl,
		Instructions: uploadInstructions,
	}, nil
}

func (s *ProfileService) UploadProfileImage(stream pb.Profile_UploadProfileImageServer) error {
	userId, tenant := auth.GetUserIdAndTenant(stream.Context())
	logger.Info("Uploading image", zap.String("userId", userId), zap.String("tenant", tenant))
	imageData, err := bootUtils.BufferGrpcServerStream(stream, func() ([]byte, error) {
		req, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return req.ChunkData, nil
	})
	if err != nil {
		logger.Error("Failed uploading image", zap.Error(err))
		return err
	}

	// upload imageData to Azure bucket.
	path := fmt.Sprintf("%s/%s/%d.jpg", tenant, userId, time.Now().Unix())
	res := <-azure.Storage.UploadStream("profile-photos", path, imageData)

	if res.Err != nil {
		logger.Error("Failed uploading profile image.", zap.Error(res.Err))
		return res.Err
	}
	stream.SendAndClose(&pb.UploadImageResponse{UploadPath: res.Value.(string)})
	return nil
}
