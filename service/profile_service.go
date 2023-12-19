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

	loginInfo, oldProfile := getExistingOrEmptyProfile(s.db, tenant, userId)

	isNewUser := false
	if len(oldProfile.LoginId) == 0 {
		isNewUser = true
		oldProfile.LoginId = userId
	}

	// merge old profile and new profile
	copier.CopyWithOption(oldProfile, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	value, ok := pb.Gender_name[int32(req.Gender)]
	if !ok {
		value = pb.Gender_name[int32(pb.Gender_Unspecified)]
	}
	oldProfile.Gender = value

	err = <-s.db.Profile(tenant).Save(oldProfile)

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

	userProfileProto := getProfileProto(loginInfo, oldProfile)
	return userProfileProto, err
}

func (s *ProfileService) GetProfileById(ctx context.Context, req *pb.IdRequest) (*pb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	loginInfo, profile := getExistingOrEmptyProfile(s.db, tenant, userId)
	profileProto := getProfileProto(loginInfo, profile)

	return profileProto, nil
}

func (s *ProfileService) GetProfileByPhoneOrEmail(ctx context.Context, req *pb.GetProfileByPhoneOrEmailRequest) (*pb.UserProfileProto, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "Email or Phone is required")
	}
	// get login info using email or phone
	loginModel := <-s.db.Login(tenant).FindOneByPhoneOrEmail(req.Phone, req.Email)

	// get profile using login id from login in
	profileResChan, errChan := s.db.Profile(tenant).FindOneById(loginModel.Id())
	profileProto := &pb.UserProfileProto{}
	select {
	case loginInfo := <-profileResChan:
		copier.CopyWithOption(profileProto, loginInfo, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	case err := <-errChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Profile not found")
		}
		logger.Error("Failed getting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}
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

func (s *ProfileService) IsUserAdmin(ctx context.Context, req *pb.IdRequest) (*pb.IsUserAdminResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	loginInfoChan, errResChan := s.db.Login(tenant).FindOneById(userId)

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
func getExistingOrEmptyProfile(db db.AuthDbInterface, tenant, userId string) (*models.LoginModel, *models.ProfileModel) {
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
	value, ok := pb.Gender_value[profileModel.Gender]
	if !ok {
		value = int32(pb.Gender_Unspecified)
	}
	result.Gender = pb.Gender(value)
	return result
}
