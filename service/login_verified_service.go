package service

import (
	"context"
	"time"

	"github.com/Kotlang/authGo/db"
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/SaiNageswarS/go-api-boot/async"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginVerifiedService struct {
	authPb.UnimplementedLoginVerifiedServer
	mongo odm.MongoClient
}

func ProvideLoginVerifiedService(
	mongo odm.MongoClient) *LoginVerifiedService {

	return &LoginVerifiedService{
		mongo: mongo,
	}
}

// RequestProfileDeletion marks profile for deletion and is used by the client
func (s *LoginVerifiedService) RequestProfileDeletion(ctx context.Context, req *authPb.ProfileDeletionRequest) (*authPb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Fetch profile info
	loginRes, err := async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		logger.Error("Failed getting login", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	loginRes.DeletionInfo = db.DeletionInfo{
		MarkedForDeletion: true,
		DeletionTime:      time.Now().Unix(),
		Reason:            req.Reason,
	}
	_, err = async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).Save(ctx, *loginRes))

	return &authPb.StatusResponse{
		Status: "Profile deletion request sent successfully",
	}, nil
}

// Admin only API
// CancelProfileDeletionRequest cancels profile deletion request and is used by admin only.
func (s *LoginVerifiedService) CancelProfileDeletionRequest(ctx context.Context, req *authPb.IdRequest) (*authPb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !db.IsAdmin(s.mongo, tenant, userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// Fetch profile info
	loginRes, err := async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		logger.Error("Failed saving profile deletion request", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed saving profile deletion request")
	}

	loginRes.DeletionInfo = db.DeletionInfo{
		MarkedForDeletion: false,
		DeletionTime:      0,
		Reason:            "",
	}
	async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).Save(ctx, *loginRes))

	return &authPb.StatusResponse{
		Status: "Profile deletion request cancelled successfully",
	}, nil
}

// Admin only API
// GetPendingProfileDeletionRequests returns all profiles marked for deletion and is used by admin only.
func (s *LoginVerifiedService) GetPendingProfileDeletionRequests(ctx context.Context, req *authPb.GetProfileDeletionRequest) (*authPb.ProfileListResponse, error) {
	userID, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !db.IsAdmin(s.mongo, tenant, userID) {
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
	loginResChan := odm.CollectionOf[db.LoginModel](s.mongo, tenant).Find(ctx, filter, nil, int64(req.PageSize), skip)
	totalCountResChan := odm.CollectionOf[db.LoginModel](s.mongo, tenant).Count(ctx, filter)

	// get total count of pending profile deletion requests
	totalCount := 0
	totalCountRes, err := async.Await(totalCountResChan)
	if err != nil {
		logger.Error("Error fetching total count of pending profile deletion requests", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching total count of pending profile deletion requests")
	}

	totalCount = int(totalCountRes)

	var login []db.LoginModel
	userIds := []string{}

	login, err = async.Await(loginResChan)
	if err != nil {
		logger.Error("Error fetching login info", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching login info")
	}

	// Extract user IDs from login info
	for _, l := range login {
		userIds = append(userIds, l.Id())
	}

	// Fetch profiles for pending profile deletion requests
	profiles, err := async.Await(db.FindProfilesByIds(ctx, s.mongo, tenant, userIds))
	if err != nil {
		logger.Error("Error fetching profiles", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching profiles")
	}

	profileProto := []*authPb.UserProfileProto{}
	for _, profile := range profiles {
		profileProto = append(profileProto, getProfileProto(&profile))
	}
	populateLoginInfo(profileProto, login)

	return &authPb.ProfileListResponse{
		Profiles:   profileProto,
		TotalUsers: int64(totalCount),
	}, nil
}

// Admin only API
// DeleteProfile deletes profile and login from db and is used by admin only.
func (s *LoginVerifiedService) DeleteProfile(ctx context.Context, req *authPb.IdRequest) (*authPb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !db.IsAdmin(s.mongo, tenant, userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// Check if profile exists
	isExists, _ := async.Await(odm.CollectionOf[db.ProfileModel](s.mongo, tenant).Exists(ctx, req.UserId))

	if !isExists {
		return &authPb.StatusResponse{
			Status: "Profile not found",
		}, nil
	}

	// Delete profile from db
	_, err := async.Await(odm.CollectionOf[db.ProfileModel](s.mongo, tenant).DeleteByID(ctx, req.UserId))
	if err != nil {
		logger.Error("Failed deleting profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed deleting profile")
	}

	// Delete login from db
	_, err = async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).DeleteByID(ctx, req.UserId))
	if err != nil {
		logger.Error("Failed deleting login", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed deleting login")
	}

	// TODO: Delete all posts, comments, notifications, etc. related to this user.

	return &authPb.StatusResponse{
		Status: "Profile deleted successfully",
	}, nil
}

// Admin only API
// check if user is admin or not and return response.
func (s *LoginVerifiedService) IsUserAdmin(ctx context.Context, req *authPb.IdRequest) (*authPb.IsUserAdminResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	// Check if user is admin
	if !db.IsAdmin(s.mongo, tenant, userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	isAdmin := db.IsAdmin(s.mongo, tenant, userId)

	return &authPb.IsUserAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

// Admin only API
func (s *LoginVerifiedService) ChangeUserType(ctx context.Context, req *authPb.ChangeUserTypeRequest) (*authPb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !db.IsAdmin(s.mongo, tenant, userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// fetch login info
	loginModel := <-db.FindOneByPhoneOrEmail(ctx, s.mongo, tenant, req.Phone, req.Email)
	if loginModel == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// change user type
	loginModel.UserType = req.UserType.String()

	// save login info
	_, err := async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).Save(ctx, *loginModel))
	if err != nil {
		logger.Error("Failed changing user type", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed changing user type")
	}

	return &authPb.StatusResponse{
		Status: "User type changed successfully",
	}, nil
}

// Admin only API
// BlockUser blocks user.
func (s *LoginVerifiedService) BlockUser(ctx context.Context, req *authPb.IdRequest) (*authPb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !db.IsAdmin(s.mongo, tenant, userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// fetch login info
	loginRes, err := async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		logger.Error("Failed getting login", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login")
	}

	loginRes.IsBlocked = true
	_, err = async.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).Save(ctx, *loginRes))
	if err != nil {
		logger.Error("Failed blocking user", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed blocking user")
	}

	return &authPb.StatusResponse{
		Status: "User blocked successfully",
	}, nil
}

func populateLoginInfo(userProfileProto []*authPb.UserProfileProto, loginInfo []db.LoginModel) {
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
