package service

import (
	"context"
	"time"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginVerifiedService struct {
	pb.UnimplementedLoginVerifiedServer
	db db.AuthDbInterface
}

func NewLoginVerifiedService(
	authDb db.AuthDbInterface) *LoginVerifiedService {

	return &LoginVerifiedService{
		db: authDb,
	}
}

func (s *LoginVerifiedService) RequestProfileDeletion(ctx context.Context, req *pb.ProfileDeletionRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Fetch profile info
	loginResChan, errChan := s.db.Login(tenant).FindOneById(userId)

	select {
	case loginRes := <-loginResChan:
		// Mark profile for deletion
		loginRes.DeletionInfo = models.DeletionInfo{
			MarkedForDeletion: true,
			DeletionTime:      time.Now().Unix(),
			Reason:            req.Reason,
		}
		err := <-s.db.Login(tenant).Save(loginRes)
		if err != nil {
			logger.Error("Failed saving profile deletion request", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed saving profile deletion request")
		}
	case err := <-errChan:
		logger.Error("Failed getting login", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	return &pb.StatusResponse{
		Status: "Profile deletion request sent successfully",
	}, nil
}

// GetPendingProfileDeletionRequests returns all profiles marked for deletion and is used by admin only.
func (s *LoginVerifiedService) GetPendingProfileDeletionRequests(ctx context.Context, req *pb.GetProfileDeletionRequest) (*pb.ProfileListResponse, error) {
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
	loginResChan, errChan := s.db.Login(tenant).Find(filter, nil, int64(req.PageSize), skip)
	totalCountResChan, countErrChan := s.db.Login(tenant).CountDocuments(filter)

	// get total count of pending profile deletion requests
	totalCount := 0
	select {
	case count := <-totalCountResChan:
		totalCount = int(count)
	case err := <-countErrChan:
		logger.Error("Error fetching user count", zap.Error(err))
	}

	login := []models.LoginModel{}
	userIds := []string{}
	select {
	case res := <-loginResChan:
		login = res
		for _, l := range login {
			userIds = append(userIds, l.Id())
		}
	case err := <-errChan:
		logger.Error("Error fetching user IDs", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching user IDs")
	}

	// Fetch profiles for pending profile deletion requests
	profileResChan, errChan := s.db.Profile(tenant).FindByIds(userIds)
	select {
	case profiles := <-profileResChan:
		profileProto := []*pb.UserProfileProto{}
		copier.Copy(&profileProto, &profiles)
		populateLoginInfo(profileProto, login)

		return &pb.ProfileListResponse{
			Profiles:   profileProto,
			TotalUsers: int64(totalCount),
		}, nil
	case err := <-errChan:
		logger.Error("Error fetching profiles", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error fetching profiles")
	}

}

// DeleteProfile deletes profile and login from db and is used by admin only.
func (s *LoginVerifiedService) DeleteProfile(ctx context.Context, req *pb.IdRequest) (*pb.StatusResponse, error) {
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
func (s *LoginVerifiedService) CancelProfileDeletionRequest(ctx context.Context, req *pb.IdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// Check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		return nil, status.Error(codes.PermissionDenied, "User with id "+userId+" don't have permission")
	}

	// Fetch profile info
	loginResChan, errChan := s.db.Login(tenant).FindOneById(req.UserId)

	select {
	case loginRes := <-loginResChan:
		// Cancel profile deletion request
		loginRes.DeletionInfo = models.DeletionInfo{
			MarkedForDeletion: false,
			DeletionTime:      0,
			Reason:            "",
		}
		err := <-s.db.Login(tenant).Save(loginRes)
		if err != nil {
			logger.Error("Failed saving profile deletion request", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed saving profile deletion request")
		}
	case err := <-errChan:
		logger.Error("Failed getting login", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile")
	}

	return &pb.StatusResponse{
		Status: "Profile deletion request cancelled successfully",
	}, nil
}

// check if user is admin or not and return response. Used by admin only.
func (s *LoginVerifiedService) IsUserAdmin(ctx context.Context, req *pb.IdRequest) (*pb.IsUserAdminResponse, error) {
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