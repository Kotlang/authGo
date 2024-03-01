package service

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/extensions"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
	"github.com/SaiNageswarS/go-api-boot/cloud"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	db       db.AuthDbInterface
	cloudFns cloud.Cloud
}

func NewProfileService(db db.AuthDbInterface, cloudFns cloud.Cloud) *ProfileService {
	return &ProfileService{
		db:       db,
		cloudFns: cloudFns,
	}
}

func (s *ProfileService) CreateOrUpdateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.UserProfileProto, error) {
	err := ValidateProfileRequest(req)
	if err != nil {
		return nil, err
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// get existing profile
	oldProfile := getExistingOrEmptyProfile(s.db, tenant, userId)

	// check if profile is marked for deletion
	if oldProfile != nil && oldProfile.DeletionInfo.MarkedForDeletion {
		return nil, status.Error(codes.PermissionDenied, "Profile marked for deletion")
	}

	isNewUser := false
	if len(oldProfile.UserId) == 0 {
		isNewUser = true
		oldProfile.UserId = userId
	}

	// merge old profile and new profile proto
	oldProfile = getProfileModel(req, oldProfile)

	// save profile to db
	err = <-s.db.Profile(tenant).Save(oldProfile)

	// if user is new, register notification event for user created.
	if isNewUser {
		extensions.RegisterEvent(ctx, &pb.RegisterEventRequest{
			EventType: "user.created",
			TemplateParameters: map[string]string{
				"userId": userId,
				"body":   fmt.Sprintf("New user '%s' joined.", req.Name),
			},
			Topic: fmt.Sprintf("%s.user.created", tenant),
		})
	}

	userProfileProto := getProfileProto(oldProfile)
	return userProfileProto, err
}

func (s *ProfileService) GetProfileById(ctx context.Context, req *pb.IdRequest) (*pb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	// get profile using userId convert to proto and return it.
	filter := bson.M{
		"_id":                            userId,
		"deletionInfo.markedForDeletion": false,
	}

	profileResChan, errChan := s.db.Profile(tenant).FindOne(filter)

	select {
	case profile := <-profileResChan:
		profileProto := getProfileProto(profile)
		return profileProto, nil
	case err := <-errChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Profile not found")
		}
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}
}

// GetProfileByPhoneOrEmail returns profile using email or phone and is used by admin only.
func (s *ProfileService) GetProfileByPhoneOrEmail(ctx context.Context, req *pb.GetProfileByPhoneOrEmailRequest) (*pb.UserProfileProto, error) {
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

	// get profile using userId convert to proto and return it.
	filter := bson.M{
		"_id":                            loginModel.Id(),
		"deletionInfo.markedForDeletion": false,
	}
	profileResChan, errChan := s.db.Profile(tenant).FindOne(filter)
	select {
	case profile := <-profileResChan:
		fmt.Println("profile", profile)
		return getProfileProto(profile), nil
	case err := <-errChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Profile not found")
		}
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}
}

func (s *ProfileService) BulkGetProfileByIds(ctx context.Context, req *pb.BulkGetProfileRequest) (*pb.ProfileListResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	profileResChan, profileErrorChan := s.db.Profile(tenant).FindByIds(req.UserIds)

	select {
	case profileRes := <-profileResChan:
		// convert profile model to proto
		profileProtoList := make([]*pb.UserProfileProto, 0)
		for _, profile := range profileRes {
			profileProtoList = append(profileProtoList, getProfileProto(&profile))
		}

		return &pb.ProfileListResponse{
			Profiles: profileProtoList,
		}, nil

	case err := <-profileErrorChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profiles")
	}
}

func (s *ProfileService) RequestProfileDeletion(ctx context.Context, req *pb.ProfileDeletionRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Fetch profile info
	profileResChan, errChan := s.db.Profile(tenant).FindOneById(userId)

	select {
	case profileRes := <-profileResChan:
		// Save profile deletion request
		profileRes.DeletionInfo = models.DeletionInfo{
			MarkedForDeletion: true,
			DeletionTime:      time.Now().Unix(),
			Reason:            req.Reason,
		}
		err := <-s.db.Profile(tenant).Save(profileRes)

		if err != nil {
			logger.Error("Failed saving profile deletion request", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed saving profile deletion request")
		}
	case err := <-errChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	return &pb.StatusResponse{
		Status: "Profile deletion request sent successfully",
	}, nil

}

// GetPendingProfileDeletionRequests returns all profiles marked for deletion and is used by admin only.
func (s *ProfileService) GetPendingProfileDeletionRequests(ctx context.Context, req *pb.GetProfileDeletionRequest) (*pb.ProfileListResponse, error) {
	userID, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userID) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userID+" don't have permission")
	}

	// Fetch pending profile deletion requests
	filter := bson.M{
		"deletionInfo.markedForDeletion": true,
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	skip := int64(req.PageNumber * req.PageSize)
	profileDeletionRequestsChan, errChan := s.db.Profile(tenant).Find(filter, nil, int64(req.PageSize), skip)
	totalCountResChan, countErrChan := s.db.Profile(tenant).CountDocuments(filter)

	// get total count of pending profile deletion requests
	totalCount := 0
	select {
	case count := <-totalCountResChan:
		totalCount = int(count)
	case err := <-countErrChan:
		logger.Error("Error fetching user count", zap.Error(err))
	}

	// convert profile model to proto
	var profileProtoList []*pb.UserProfileProto
	var userIds []string
	select {
	case profileDeletionRequests := <-profileDeletionRequestsChan:
		for _, profile := range profileDeletionRequests {
			profileProtoList = append(profileProtoList, getProfileProto(&profile))
			userIds = append(userIds, profile.UserId)
		}

	case err := <-errChan:
		logger.Error("Failed getting profile deletion requests", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile deletion requests")
	}

	// get login info using userId
	loginInfoChan, errChan := s.db.Login(tenant).FindByIds(userIds)

	// populate phone number field in profile proto
	var loginInfo []models.LoginModel
	select {
	case loginInfo = <-loginInfoChan:
	case <-errChan:
		logger.Error("Failed getting login info")
	}

	if len(loginInfo) > 0 {
		populateLoginInfo(profileProtoList, loginInfo)
	}

	return &pb.ProfileListResponse{
		Profiles:   profileProtoList,
		TotalUsers: int64(totalCount),
	}, nil
}

// DeleteProfile deletes profile and login from db and is used by admin only.
func (s *ProfileService) DeleteProfile(ctx context.Context, req *pb.IdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// Check if profile exists
	isExists := s.db.Profile(tenant).IsExistsById(req.UserId)

	if !isExists {
		return &pb.StatusResponse{
			Status: "Profile not found",
		}, nil
	}

	// Delete profile from db
	err := <-s.db.Profile(tenant).DeleteById(req.UserId)
	if err != nil {
		logger.Error("Failed deleting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed deleting profile")
	}

	// Delete login from db
	err = <-s.db.Login(tenant).DeleteById(req.UserId)
	if err != nil {
		logger.Error("Failed deleting login", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed deleting login")
	}

	// TODO: Delete all posts, comments, notifications, etc. related to this user.

	return &pb.StatusResponse{
		Status: "Profile deleted successfully",
	}, nil
}

// CancelProfileDeletionRequest cancels profile deletion request and is used by admin only.
func (s *ProfileService) CancelProfileDeletionRequest(ctx context.Context, req *pb.IdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// Fetch profile info
	profileResChan, errChan := s.db.Profile(tenant).FindOneById(req.UserId)

	select {
	case profileRes := <-profileResChan:
		// Un mark profile for deletion
		profileRes.DeletionInfo = models.DeletionInfo{
			MarkedForDeletion: false,
			DeletionTime:      0,
			Reason:            "",
		}
		err := <-s.db.Profile(tenant).Save(profileRes)

		if err != nil {
			logger.Error("Failed saving profile deletion request", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed saving profile deletion request")
		}
	case err := <-errChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	return &pb.StatusResponse{
		Status: "Profile deletion request cancelled successfully",
	}, nil
}

// check if user is admin or not and return response. Used by admin only.
func (s *ProfileService) IsUserAdmin(ctx context.Context, req *pb.IdRequest) (*pb.IsUserAdminResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	isAdmin := s.db.Login(tenant).IsAdmin(userId)

	return &pb.IsUserAdminResponse{
		IsAdmin: isAdmin,
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
	return &pb.ProfileImageUploadURL{
		UploadUrl:    preSignedUrl,
		MediaUrl:     downloadUrl,
		Instructions: uploadInstructions,
	}, nil
}

// UploadProfileImage uploads profile image to cloud bucket with max size of 5mb.
func (s *ProfileService) UploadProfileImage(stream pb.Profile_UploadProfileImageServer) error {
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
		stream.SendAndClose(&pb.UploadImageResponse{UploadPath: result})
		return nil
	case err := <-errorChan:
		logger.Error("Failed uploading image", zap.Error(err))
		return err
	}
}

func (s *ProfileService) FetchProfiles(ctx context.Context, req *pb.FetchProfilesRequest) (*pb.ProfileListResponse, error) {
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

	userProfileProto := []*pb.UserProfileProto{}
	for _, userModel := range profiles {
		userProfileProto = append(userProfileProto, getProfileProto(&userModel))
	}

	// populate phone number field in profile proto
	var loginInfo []models.LoginModel
	select {
	case loginInfo = <-loginInfoChan:
	case <-errChan:
		logger.Error("Failed getting login info")
	}

	if len(loginInfo) > 0 {
		populateLoginInfo(userProfileProto, loginInfo)
	}

	response := &pb.ProfileListResponse{Profiles: userProfileProto, TotalUsers: int64(totalCount)}
	return response, nil
}

func (s *ProfileService) ChangeUserType(ctx context.Context, req *pb.ChangeUserTypeRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// fetch login info
	loginModel := <-s.db.Login(tenant).FindOneByPhoneOrEmail(req.Phone, req.Email)
	if loginModel == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// change user type
	loginModel.UserType = req.UserType.String()

	// save login info
	err := <-s.db.Login(tenant).Save(loginModel)
	if err != nil {
		logger.Error("Failed changing user type", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed changing user type")
	}

	return &pb.StatusResponse{
		Status: "User type changed successfully",
	}, nil
}

func (s *ProfileService) BlockUser(ctx context.Context, req *pb.IdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// fetch profile info
	profileResChan, errChan := s.db.Profile(tenant).FindOneById(req.UserId)

	select {
	case profileRes := <-profileResChan:
		// block user
		profileRes.IsBlocked = true
		err := <-s.db.Profile(tenant).Save(profileRes)

		if err != nil {
			logger.Error("Failed blocking user", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed blocking user")
		}
	case err := <-errChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	return &pb.StatusResponse{
		Status: "User blocked successfully",
	}, nil
}

func (s *ProfileService) UnblockUser(ctx context.Context, req *pb.IdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// fetch profile info
	profileResChan, errChan := s.db.Profile(tenant).FindOneById(req.UserId)

	select {
	case profileRes := <-profileResChan:
		// unblock user
		profileRes.IsBlocked = false
		err := <-s.db.Profile(tenant).Save(profileRes)

		if err != nil {
			logger.Error("Failed unblocking user", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed unblocking user")
		}
	case err := <-errChan:
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	return &pb.StatusResponse{
		Status: "User unblocked successfully",
	}, nil
}

// get profile proto from profile model
func getProfileProto(profileModel *models.ProfileModel) *pb.UserProfileProto {
	result := &pb.UserProfileProto{}

	if profileModel == nil {
		return result
	}

	copier.CopyWithOption(result, profileModel, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})

	// copy gender value
	value, ok := pb.Gender_value[profileModel.Gender]
	if !ok {
		value = int32(pb.Gender_Unspecified)
	}
	result.Gender = pb.Gender(value)

	// copy farming type
	value, ok = pb.FarmingType_value[profileModel.FarmingType]
	if !ok {
		value = int32(pb.FarmingType_UnspecifiedFarming)
	}
	result.FarmingType = pb.FarmingType(value)

	// copy land size
	value, ok = pb.LandSizeInAcres_value[profileModel.LandSizeInAcres]
	if !ok {
		value = int32(pb.LandSizeInAcres_UnspecifiedLandSize)
	}
	result.LandSizeInAcres = pb.LandSizeInAcres(value)

	return result
}

// get profile model from profile proto
func getProfileModel(profileProto *pb.CreateProfileRequest, profileModel *models.ProfileModel) *models.ProfileModel {

	if profileModel == nil {
		profileModel = &models.ProfileModel{}
	}

	copier.CopyWithOption(profileModel, profileProto, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	//copy gender if not unspecified
	if profileProto.Gender != pb.Gender_Unspecified {
		value, ok := pb.Gender_name[int32(profileProto.Gender)]
		if !ok {
			value = pb.Gender_name[int32(pb.Gender_Unspecified)]
		}
		profileModel.Gender = value
	}

	//copy farming type if not unspecified
	if profileProto.FarmingType != pb.FarmingType_UnspecifiedFarming {
		value, ok := pb.FarmingType_name[int32(profileProto.FarmingType)]
		if !ok {
			value = pb.FarmingType_name[int32(pb.FarmingType_UnspecifiedFarming)]
		}
		profileModel.FarmingType = value
	}

	//copy land size if not unspecified
	if profileProto.LandSizeInAcres != pb.LandSizeInAcres_UnspecifiedLandSize {
		value, ok := pb.LandSizeInAcres_name[int32(profileProto.LandSizeInAcres)]
		if !ok {
			value = pb.LandSizeInAcres_name[int32(pb.LandSizeInAcres_UnspecifiedLandSize)]
		}
		profileModel.LandSizeInAcres = value
	}
	return profileModel
}

// gets profile for userId or return empty model if doesn't exist.
func getExistingOrEmptyProfile(db db.AuthDbInterface, tenant, userId string) *models.ProfileModel {
	profile := &models.ProfileModel{}

	profileResChan, profileErrorChan := db.Profile(tenant).FindOneById(userId)

	// in case of error, return empty profile.
	select {
	case profileRes := <-profileResChan:
		profile = profileRes
	case <-profileErrorChan:
		logger.Error("Failed getting profile", zap.String("userId", userId), zap.String("tenant", tenant))
	}

	return profile
}

func populateLoginInfo(userProfileProto []*pb.UserProfileProto, loginInfo []models.LoginModel) {
	for i, profile := range userProfileProto {
		for _, loginModel := range loginInfo {
			if profile.UserId == loginModel.UserId {
				userProfileProto[i].PhoneNumber = loginModel.Phone
				copier.Copy(&userProfileProto[i], &loginModel)
				break
			}
		}
	}
}
