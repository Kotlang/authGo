package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Kotlang/authGo/appconfig"
	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/extensions"
	authPb "github.com/Kotlang/authGo/generated/auth"
	notificationPb "github.com/Kotlang/authGo/generated/notification"
	"github.com/SaiNageswarS/go-api-boot/async"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
	"github.com/SaiNageswarS/go-api-boot/cloud"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileService struct {
	authPb.UnimplementedProfileServer
	ccfg     *appconfig.AppConfig
	mongo    odm.MongoClient
	cloudFns cloud.Cloud
}

func ProvideProfileService(mongo odm.MongoClient, cloudFns cloud.Cloud, ccfg *appconfig.AppConfig) *ProfileService {
	return &ProfileService{
		mongo:    mongo,
		cloudFns: cloudFns,
		ccfg:     ccfg,
	}
}

// CreateOrUpdateProfile creates or updates profile for user.
// All the fields are optional and only the fields except name. Fields provided in request will be updated.
func (s *ProfileService) CreateOrUpdateProfile(ctx context.Context, req *authPb.CreateProfileRequest) (*authPb.UserProfileProto, error) {
	err := ValidateProfileRequest(req)
	if err != nil {
		return nil, err
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	logger.Info("Creating or updating profile", zap.String("userId", userId), zap.String("tenant", tenant))

	// get existing profile
	oldProfile, _ := async.Await(odm.CollectionOf[db.ProfileModel](s.mongo, tenant).FindOneByID(ctx, userId))

	isNewUser := false
	if oldProfile == nil {
		isNewUser = true
		oldProfile = &db.ProfileModel{
			UserId: userId,
		}
	}

	// merge old profile and new profile proto
	oldProfile = getProfileModel(req, oldProfile)

	// save profile to db
	_, err = async.Await(odm.CollectionOf[db.ProfileModel](s.mongo, tenant).Save(ctx, *oldProfile))

	// if user is new, register notification event for user created.
	if isNewUser {
		extensions.RegisterEvent(ctx, &notificationPb.RegisterEventRequest{
			EventType: "post.created",
			Title:     "नया उपयोगकर्ता हमारे साथ जुड़े हैं।",
			Body:      "",
			TemplateParameters: map[string]string{
				"userId": userId,
			},
			Topic:       fmt.Sprintf("%s.post.created", tenant),
			TargetUsers: []string{userId},
		})
	}

	userProfileProto := getProfileProto(oldProfile)
	return userProfileProto, err
}

// GetProfile returns profile for user. checks if user is blocked or marked for deletion.
func (s *ProfileService) GetProfileById(ctx context.Context, req *authPb.IdRequest) (*authPb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	_, err := async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	profile, err := async.Await(odm.CollectionOf[db.ProfileModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Profile not found")
		}
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	profileProto := getProfileProto(profile)
	return profileProto, nil
}

// BulkGetProfileByIds returns profiles for given user ids.
// Login info is fetched first and then profile info is fetched using userIds which are not marked for deletion or blocked.
func (s *ProfileService) BulkGetProfileByIds(ctx context.Context, req *authPb.BulkGetProfileRequest) (*authPb.ProfileListResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	// login info
	loginInfo, err := async.Await(db.FindLoginsByIds(ctx, s.mongo, tenant, req.UserIds))
	if err != nil {
		logger.Error("Failed getting login info", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info")
	}

	// profile info
	userIds := []string{}
	for _, login := range loginInfo {
		if !login.DeletionInfo.MarkedForDeletion && !login.IsBlocked {
			userIds = append(userIds, login.UserId)
		}
	}

	profileRes, err := async.Await(db.FindProfilesByIds(ctx, s.mongo, tenant, userIds))
	if err != nil {
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	profileProtoList := make([]*authPb.UserProfileProto, 0)
	for _, profile := range profileRes {
		profileProtoList = append(profileProtoList, getProfileProto(&profile))
	}

	return &authPb.ProfileListResponse{
		Profiles: profileProtoList,
	}, nil
}

// GetProfileImageUploadUrl returns presigned url for uploading profile image.
func (s *ProfileService) GetProfileImageUploadUrl(ctx context.Context, req *authPb.ProfileImageUploadRequest) (*authPb.ProfileImageUploadURL, error) {
	uploadInstructions := `
	| 1. Send profile image file to above uploadURL as a PUT request. 
	| 
	| curl --location --request PUT '<aboveURL>' 
	|      --header 'Content-Type: image/jpeg' 
	|      --data-binary '@/path/to/file.jpg'
	|      
	| 2. Send mediaUrl in createOrUpdateProfile request.`

	userId, tenant := auth.GetUserIdAndTenant(ctx)

	acceptableExtensions := []string{"jpg", "jpeg", "png"}
	if !slices.Contains(acceptableExtensions, req.MediaExtension) {
		return nil, status.Error(codes.InvalidArgument, "Invalid media extension")
	}

	if req.MediaExtension == "" {
		req.MediaExtension = "jpg"
	}
	contentType := fmt.Sprintf("image/%s", req.MediaExtension)
	key := fmt.Sprintf("%s/%s/%d.%s", tenant, userId, time.Now().Unix(), req.MediaExtension)
	profileBucket := s.ccfg.ProfileBucket
	if profileBucket == "" {
		return nil, status.Error(codes.Internal, "profile_bucket is not set")
	}

	preSignedUrl, downloadUrl := s.cloudFns.GetPresignedUrl(ctx, profileBucket, key, contentType, 10*time.Minute)
	return &authPb.ProfileImageUploadURL{
		UploadUrl:    preSignedUrl,
		MediaUrl:     downloadUrl,
		Instructions: uploadInstructions,
	}, nil
}

// UploadProfileImage uploads profile image to cloud bucket with max size of 5mb.
func (s *ProfileService) UploadProfileImage(stream grpc.ClientStreamingServer[authPb.UploadImageRequest, authPb.UploadImageResponse]) error {
	userId, tenant := auth.GetUserIdAndTenant(stream.Context())
	logger.Info("Uploading image", zap.String("userId", userId), zap.String("tenant", tenant))
	acceptableMimeTypes := []string{"image/jpeg", "image/png"}

	imageData, contentType, err := bootUtils.BufferGrpcServerStream(
		acceptableMimeTypes,
		5*1024*1024, // 5mb max file size.
		func() ([]byte, error) {
			err := bootUtils.StreamContextError(stream.Context())
			if err != nil {
				return nil, err
			}

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

	file_extension := bootUtils.GetFileExtension(contentType)
	// upload imageData to Azure bucket.
	path := fmt.Sprintf("%s/%s/%d.%s", tenant, userId, time.Now().Unix(), file_extension)
	profileBucket := s.ccfg.ProfileBucket
	if profileBucket == "" {
		return status.Error(codes.Internal, "profile_bucket is not set")
	}
	uploadPath, err := s.cloudFns.UploadBuffer(stream.Context(), profileBucket, path, imageData.Bytes())

	if err != nil {
		logger.Error("Failed uploading image to cloud", zap.Error(err))
		return err
	}

	stream.SendAndClose(&authPb.UploadImageResponse{UploadPath: uploadPath})
	return nil
}

// get profile proto from profile model
func getProfileProto(profileModel *db.ProfileModel) *authPb.UserProfileProto {
	result := &authPb.UserProfileProto{}

	if profileModel == nil {
		return result
	}

	copier.CopyWithOption(result, profileModel, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})

	// copy gender value
	value, ok := authPb.Gender_value[profileModel.Gender]
	if !ok {
		value = int32(authPb.Gender_Unspecified)
	}
	result.Gender = authPb.Gender(value)

	// copy farming type
	value, ok = authPb.FarmingType_value[profileModel.FarmingType]
	if !ok {
		value = int32(authPb.FarmingType_UnspecifiedFarming)
	}
	result.FarmingType = authPb.FarmingType(value)

	// copy land size
	value, ok = authPb.LandSizeInAcres_value[profileModel.LandSizeInAcres]
	if !ok {
		value = int32(authPb.LandSizeInAcres_UnspecifiedLandSize)
	}
	result.LandSizeInAcres = authPb.LandSizeInAcres(value)

	return result
}

// get profile model from profile proto
func getProfileModel(profileProto *authPb.CreateProfileRequest, profileModel *db.ProfileModel) *db.ProfileModel {

	if profileModel == nil {
		profileModel = &db.ProfileModel{}
	}

	copier.CopyWithOption(profileModel, profileProto, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	//copy gender if not unspecified
	if profileProto.Gender != authPb.Gender_Unspecified {
		value, ok := authPb.Gender_name[int32(profileProto.Gender)]
		if !ok {
			value = authPb.Gender_name[int32(authPb.Gender_Unspecified)]
		}
		profileModel.Gender = value
	}

	//copy farming type if not unspecified
	if profileProto.FarmingType != authPb.FarmingType_UnspecifiedFarming {
		value, ok := authPb.FarmingType_name[int32(profileProto.FarmingType)]
		if !ok {
			value = authPb.FarmingType_name[int32(authPb.FarmingType_UnspecifiedFarming)]
		}
		profileModel.FarmingType = value
	}

	//copy land size if not unspecified
	if profileProto.LandSizeInAcres != authPb.LandSizeInAcres_UnspecifiedLandSize {
		value, ok := authPb.LandSizeInAcres_name[int32(profileProto.LandSizeInAcres)]
		if !ok {
			value = authPb.LandSizeInAcres_name[int32(authPb.LandSizeInAcres_UnspecifiedLandSize)]
		}
		profileModel.LandSizeInAcres = value
	}
	return profileModel
}
