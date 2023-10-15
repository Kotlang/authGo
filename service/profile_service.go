package service

import (
	"context"
	"encoding/json"
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
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProfileService represents a service that handles user profile related operations.
// It is responsible for managing user profile data and exposing it to the client.
type ProfileService struct {
	pb.UnimplementedProfileServer
	db *db.AuthDb
}

// NewProfileService creates a new instance of the ProfileService struct.
// It takes a pointer to an AuthDb instance as its argument and returns a pointer to a ProfileService instance.
// This function is used to initialize a new ProfileService object.
func NewProfileService(db *db.AuthDb) *ProfileService {
	return &ProfileService{
		db: db,
	}
}

// CreateOrUpdateProfile creates or updates a user profile with the given request.
// It validates the request, gets the user ID and tenant from the context, and retrieves the existing profile if it exists.
// If the user is new, it sets the login ID to the user ID.
// It then merges the old profile and the new profile, sets the gender, and saves the profile to the database.
// If the user is new, it registers a "user.created" event with the user ID and name.
// Finally, it returns the user profile and any errors that occurred during the process.
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
	newMetadata := copyAll(oldProfile.MetadataMap, getMapFromJson(req.MetaDataMap))
	copier.CopyWithOption(oldProfile, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	value, ok := pb.Gender_name[int32(req.Gender)]
	if !ok {
		value = pb.Gender_name[int32(pb.Gender_Unspecified)]
	}
	oldProfile.Gender = value
	oldProfile.MetadataMap = newMetadata

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

// GetProfileById retrieves the user profile for the given user ID.
// If the user ID is not provided, it retrieves the profile for the authenticated user.
// Returns the user profile in the form of a protobuf message.
func (s *ProfileService) GetProfileById(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserProfileProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	loginInfo, profile := getExistingOrEmptyProfile(s.db, tenant, userId)
	profileProto := getProfileProto(loginInfo, profile)

	return profileProto, nil
}

// BulkGetProfileByIds retrieves multiple user profiles by their IDs.
// It takes a context and a BulkGetProfileRequest as input, and returns a BulkGetProfileResponse and an error.
// The function retrieves the user ID and tenant from the context, and uses them to find the profiles and login information in the database.
// It then creates a list of UserProfileProto objects and returns them in a BulkGetProfileResponse.
// If there is an error retrieving the profiles or login information, the function returns an error with an appropriate status code.
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

// GetProfileImageUploadUrl generates a pre-signed URL for uploading a user's profile image to an S3 bucket.
// The function returns a ProfileImageUploadURL object containing the pre-signed URL, the download URL, and upload instructions.
// To upload the image, send a PUT request to the pre-signed URL with the image file as binary data.
// After uploading the image, send the mediaUrl in a createOrUpdateProfile request.
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

// UploadProfileImage handles the uploading of a user's profile image.
// It receives a stream of data chunks from the client, validates the image data,
// and uploads it to an Azure bucket. The function returns an error if the upload fails.
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

// copyAll copies all key-value pairs from the newMap to the oldMap.
// If oldMap is nil, it creates a new map.
// Returns the updated oldMap.
func copyAll(oldMap, newMap map[string]interface{}) map[string]interface{} {
	if oldMap == nil {
		oldMap = make(map[string]interface{})
	}

	for k, v := range newMap {
		oldMap[k] = v
	}

	return oldMap
}

// getMapFromJson function takes a JSON string as input and returns a map of string keys and interface{} values.
// This function is used to convert a JSON string to a map in Go.
func getMapFromJson(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

// gets profile for userId or return empty model if doesn't exist.
// getExistingOrEmptyProfile retrieves the profile and login information of a user from the database.
// If the user does not exist in the database, it returns an empty profile and login information.
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

// getProfileProto returns a protobuf representation of a user profile, given a LoginModel and a ProfileModel.
// If the ProfileModel is nil, an empty UserProfileProto is returned.
// The function copies the fields from the ProfileModel and LoginModel to the resulting UserProfileProto.
// The gender field is converted to the corresponding protobuf value, or set to Gender_Unspecified if the value is not recognized.
// The metadata map is serialized to a JSON string and stored in the MetaDataMap field of the resulting UserProfileProto.
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
	// serialize metadata map.
	metadataString, err := json.Marshal(profileModel.MetadataMap)
	if err != nil {
		logger.Error("Failed serializing metadata json", zap.Any("MetadataMap", profileModel.MetadataMap))
	}

	result.MetaDataMap = string(metadataString)
	return result
}
