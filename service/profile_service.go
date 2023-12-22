package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/extensions"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/aws"
	"github.com/SaiNageswarS/go-api-boot/azure"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileService struct {
	pb.UnimplementedProfileServer
	db db.AuthDbInterface
}

func NewProfileService(db db.AuthDbInterface) *ProfileService {
	return &ProfileService{
		db: db,
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

	isNewUser := false
	if len(oldProfile.LoginId) == 0 {
		isNewUser = true
		oldProfile.LoginId = userId
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
	profile := getExistingOrEmptyProfile(s.db, tenant, userId)
	profileProto := getProfileProto(profile)

	return profileProto, nil
}

func (s *ProfileService) GetProfileByPhoneOrEmail(ctx context.Context, req *pb.GetProfileByPhoneOrEmailRequest) (*pb.UserProfileProto, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "Email or Phone is required")
	}
	// get login info using email or phone
	loginModel := <-s.db.Login(tenant).FindOneByPhoneOrEmail(req.Phone, req.Email)

	if loginModel == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// get profile using login id from login in
	profileResChan, errChan := s.db.Profile(tenant).FindOneById(loginModel.Id())
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

// check if user is admin or not.
func (s *ProfileService) IsUserAdmin(ctx context.Context, req *pb.IdRequest) (*pb.IsUserAdminResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	loginInfoChan, errResChan := s.db.Login(tenant).FindOneById(userId)

	//get login info using userId
	select {
	case loginInfo := <-loginInfoChan:
		return &pb.IsUserAdminResponse{
			IsAdmin: loginInfo.UserType == "admin",
		}, nil
	case err := <-errResChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		logger.Error("Failed getting login info", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info")
	}
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

// UploadProfileImage uploads profile image to azure bucket with max size of 5mb.
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
