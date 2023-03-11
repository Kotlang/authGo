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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *ProfileService) CreateOrUpdateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginInfo, oldProfile := getExistingOrEmptyProfile(s.db, tenant, userId)

	if len(oldProfile.LoginId) == 0 {
		oldProfile.LoginId = userId
	}

	// merge old profile and new profile
	newMetadata := copyAll(oldProfile.MetadataMap, getMapFromJson(req.MetaDataMap))
	copier.CopyWithOption(oldProfile, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	oldProfile.Gender = pb.Gender_name[int32(req.Gender)]
	oldProfile.MetadataMap = newMetadata

	err := <-s.db.Profile(tenant).Save(oldProfile)

	userProfileProto := getProfileProto(loginInfo, oldProfile)
	return userProfileProto, err
}

func (s *ProfileService) GetProfileById(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	loginInfo, profile := getExistingOrEmptyProfile(s.db, tenant, userId)
	profileProto := getProfileProto(loginInfo, profile)

	return profileProto, nil
}

func (s *ProfileService) BulkGetProfileByIds(ctx context.Context, req *pb.BulkGetProfileRequest) (*pb.BulkGetProfileResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	profileResChan, profileErrorChan := s.db.Profile(tenant).FindByIds(req.UserIds)
	loginInfoChan, loginErrorChan := s.db.Login(tenant).FindByIds(req.UserIds)

	profileMap := make(map[string]models.ProfileModel)
	loginMap := make(map[string]models.LoginModel)

	select {
	case profileRes := <-profileResChan:
		for _, profile := range profileRes {
			profileMap[profile.Id()] = profile
		}
	case err := <-profileErrorChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profiles")
	}

	select {
	case loginRes := <-loginInfoChan:
		for _, login := range loginRes {
			loginMap[login.Id()] = login
		}
	case err := <-loginErrorChan:
		logger.Error("Failed getting login info", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info")
	}

	profileProtoList := make([]*pb.UserProfileProto, 0)
	for _, userId := range req.UserIds {
		loginInfo, profile := loginMap[userId], profileMap[userId]
		profileProtoList = append(profileProtoList, getProfileProto(&loginInfo, &profile))
	}

	return &pb.BulkGetProfileResponse{
		Profiles: profileProtoList,
	}, nil
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
	resultChan, errorChan := azure.Storage.UploadStream("profile-photos", path, imageData)

	select {
	case result := <-resultChan:
		stream.SendAndClose(&pb.UploadImageResponse{UploadPath: result})
		return nil
	case err := <-errorChan:
		logger.Error("Failed uploading image", zap.Error(err))
		return err
	}
}

func copyAll(oldMap, newMap map[string]interface{}) map[string]interface{} {
	if oldMap == nil {
		oldMap = make(map[string]interface{})
	}

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

// gets profile for userId or return empty model if doesn't exist.
func getExistingOrEmptyProfile(db *db.AuthDb, tenant, userId string) (*models.LoginModel, *models.ProfileModel) {
	profile := &models.ProfileModel{}
	loginInfo := &models.LoginModel{}

	profileResChan, profileErrorChan := db.Profile(tenant).FindOneById(userId)
	loginInfoChan, loginErrorChan := db.Login(tenant).FindOneById(userId)

	// in case of error, return empty profile.
	select {
	case profileRes := <-profileResChan:
		profile = profileRes
	case <-profileErrorChan:
		logger.Error("Failed getting profile", zap.String("userId", userId), zap.String("tenant", tenant))
	}

	select {
	case loginRes := <-loginInfoChan:
		loginInfo = loginRes
	case <-loginErrorChan:
		logger.Error("Failed getting login info", zap.String("userId", userId), zap.String("tenant", tenant))
	}

	return loginInfo, profile
}

func getProfileProto(loginModel *models.LoginModel, profileModel *models.ProfileModel) *pb.UserProfileProto {
	result := &pb.UserProfileProto{}

	if profileModel == nil {
		return result
	}

	copier.Copy(result, profileModel)
	copier.CopyWithOption(result, loginModel, copier.Option{IgnoreEmpty: true})
	result.Gender = pb.Gender(pb.Gender_value[profileModel.Gender])
	// serialize metadata map.
	metadataString, err := json.Marshal(profileModel.MetadataMap)
	if err != nil {
		logger.Error("Failed serializing metadata json", zap.Any("MetadataMap", profileModel.MetadataMap))
	}

	result.MetaDataMap = string(metadataString)
	return result
}
