package service

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/extensions"
	authPb "github.com/Kotlang/authGo/generated/auth"
	notificationPb "github.com/Kotlang/authGo/generated/notification"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
	"github.com/SaiNageswarS/go-api-boot/cloud"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileService struct {
	authPb.UnimplementedProfileServer
	db       db.AuthDbInterface
	cloudFns cloud.Cloud
}

func ProvideProfileService(db db.AuthDbInterface, cloudFns cloud.Cloud) *ProfileService {
	return &ProfileService{
		db:       db,
		cloudFns: cloudFns,
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

	// get existing profile
	oldProfile := getExistingOrEmptyProfile(s.db, tenant, userId)

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
	err = <-s.db.Profile(tenant).Save(oldProfile)

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

	loginResChan, errChan := s.db.Login(tenant).FindOneById(userId)
	var login *db.LoginModel
	select {
	case login = <-loginResChan:
		if login.DeletionInfo.MarkedForDeletion && login.UserType != "admin" {
			return nil, status.Error(codes.PermissionDenied, "Profile Marked for Deletion")
		}

		if login.IsBlocked && login.UserType != "admin" {
			return nil, status.Error(codes.PermissionDenied, "User is blocked")
		}
	case err := <-errChan:
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
	}

	profileResChan, errChan := s.db.Profile(tenant).FindOneById(userId)
	select {
	case profile := <-profileResChan:
		profileProto := getProfileProto(profile)

		if login.UserType == "admin" {
			copier.CopyWithOption(profileProto, login, copier.Option{IgnoreEmpty: true})
			profileProto.PhoneNumber = login.Phone
		}

		return profileProto, nil
	case err := <-errChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Profile not found")
		}
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}
}

// BulkGetProfileByIds returns profiles for given user ids.
// Login info is fetched first and then profile info is fetched using userIds which are not marked for deletion or blocked.
func (s *ProfileService) BulkGetProfileByIds(ctx context.Context, req *authPb.BulkGetProfileRequest) (*authPb.ProfileListResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	// login info
	loginResChan, errChan := s.db.Login(tenant).FindByIds(req.UserIds)
	var loginInfo []db.LoginModel
	select {
	case loginInfo = <-loginResChan:
	case <-errChan:
		logger.Error("Failed getting login info")
		return nil, status.Error(codes.Internal, "Failed getting login info")
	}

	// profile info
	userIds := []string{}
	for _, login := range loginInfo {
		if !login.DeletionInfo.MarkedForDeletion && !login.IsBlocked {
			userIds = append(userIds, login.UserId)
		}
	}

	profileResChan, profileErrorChan := s.db.Profile(tenant).FindByIds(userIds)

	select {
	case profileRes := <-profileResChan:
		// convert profile model to proto
		profileProtoList := make([]*authPb.UserProfileProto, 0)
		for _, profile := range profileRes {
			profileProtoList = append(profileProtoList, getProfileProto(&profile))
		}

		return &authPb.ProfileListResponse{
			Profiles: profileProtoList,
		}, nil

	case err := <-profileErrorChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profiles")
	}
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
	profileBucket := os.Getenv("profile_bucket")
	if profileBucket == "" {
		return nil, status.Error(codes.Internal, "profile_bucket is not set")
	}

	preSignedUrl, downloadUrl := s.cloudFns.GetPresignedUrl(profileBucket, key, contentType, 10*time.Minute)
	return &authPb.ProfileImageUploadURL{
		UploadUrl:    preSignedUrl,
		MediaUrl:     downloadUrl,
		Instructions: uploadInstructions,
	}, nil
}

// UploadProfileImage uploads profile image to cloud bucket with max size of 5mb.
func (s *ProfileService) UploadProfileImage(stream authPb.Profile_UploadProfileImageServer) error {
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
	profileBucket := os.Getenv("profile_bucket")
	if profileBucket == "" {
		return status.Error(codes.Internal, "profile_bucket is not set")
	}
	resultChan, errorChan := s.cloudFns.UploadStream(profileBucket, path, imageData)

	select {
	case result := <-resultChan:
		stream.SendAndClose(&authPb.UploadImageResponse{UploadPath: result})
		return nil
	case err := <-errorChan:
		logger.Error("Failed uploading image", zap.Error(err))
		return err
	}
}

// Admin only API
// GetProfileByPhoneOrEmail returns profile using email or phone and is used by admin only.
func (s *ProfileService) GetProfileByPhoneOrEmail(ctx context.Context, req *authPb.GetProfileByPhoneOrEmailRequest) (*authPb.UserProfileProto, error) {
	userID, tenant := auth.GetUserIdAndTenant(ctx)

	//validations
	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "Email or Phone is required")
	}

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userID) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userID+" don't have permission")
	}

	// get login info using email or phone
	loginModel := <-s.db.Login(tenant).FindOneByPhoneOrEmail(req.Phone, req.Email)

	if loginModel == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	profileResChan, errChan := s.db.Profile(tenant).FindOneById(loginModel.UserId)
	select {
	case profile := <-profileResChan:
		return getProfileProto(profile), nil
	case err := <-errChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Profile not found")
		}
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}
}

// Admin only API
// FetchProfiles returns list of profiles based on filters and pagination.
func (s *ProfileService) FetchProfiles(ctx context.Context, req *authPb.FetchProfilesRequest) (*authPb.ProfileListResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	profiles, totalCount := s.db.Profile(tenant).GetProfiles(req.Filters, int64(req.PageSize), int64(req.PageNumber))

	userIds := []string{}
	for _, profile := range profiles {
		userIds = append(userIds, profile.UserId)
	}

	// get login info using userId
	loginInfoChan, errChan := s.db.Login(tenant).FindByIds(userIds)

	userProfileProto := []*authPb.UserProfileProto{}
	for _, userModel := range profiles {
		userProfileProto = append(userProfileProto, getProfileProto(&userModel))
	}

	// populate phone number field in profile proto
	var loginInfo []db.LoginModel
	select {
	case loginInfo = <-loginInfoChan:
	case <-errChan:
		logger.Error("Failed getting login info")
	}

	if len(loginInfo) > 0 {
		populateLoginInfo(userProfileProto, loginInfo)
	}

	response := &authPb.ProfileListResponse{Profiles: userProfileProto, TotalUsers: int64(totalCount)}
	return response, nil
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

// gets profile for userId or return empty model if doesn't exist.
func getExistingOrEmptyProfile(db db.AuthDbInterface, tenant, userId string) *db.ProfileModel {
	profileResChan, profileErrorChan := db.Profile(tenant).FindOneById(userId)

	// in case of error, return empty profile.
	select {
	case profileRes := <-profileResChan:
		return profileRes
	case <-profileErrorChan:
		logger.Error("Failed getting profile", zap.String("userId", userId), zap.String("tenant", tenant))
		return nil
	}
}
