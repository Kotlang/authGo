package service

import (
	"context"

	"github.com/Kotlang/authGo/auth"
	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/logger"
	"github.com/Kotlang/authGo/models"
	"github.com/Kotlang/authGo/otp"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginService struct {
	pb.UnimplementedLoginServer
	db          *db.AuthDb
	emailClient *otp.EmailClient
}

func NewLoginService(
	authDb *db.AuthDb,
	emailClient *otp.EmailClient) *LoginService {

	return &LoginService{
		db:          authDb,
		emailClient: emailClient,
	}
}

//removing auth interceptor
func (u *LoginService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}

func (s *LoginService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.StatusResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant.FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}

	if s.emailClient.IsValidEmail(req.EmailOrPhone) {
		logger.Info("Sending otp to ", zap.String("email", req.EmailOrPhone))
		s.emailClient.SendOtp(tenantDetails.Name, req.EmailOrPhone)
	}
	return &pb.StatusResponse{Status: "success"}, nil
}

func (s *LoginService) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.AuthResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant.FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}

	var loginInfo *models.LoginModel
	var err error
	if s.emailClient.IsValidEmail(req.EmailOrPhone) {
		loginInfo, err = s.emailClient.ValidateOtpAndGetLoginInfo(tenantDetails.Name, req.EmailOrPhone, req.Otp)
		if err != nil {
			return nil, err
		}
	}

	profileData := &models.ProfileModel{
		LoginId: loginInfo.Id(),
		Tenant:  tenantDetails.Name,
	}
	err = <-s.db.Profile.FindOneById(profileData)
	if err != nil {
		logger.Error("Failed fetching profile", zap.Error(err))
	}

	logger.Info("Fetched profile as ", zap.Any("profile", profileData))
	profileProto := &pb.UserProfileProto{}
	copier.Copy(profileProto, profileData)

	jwtToken := auth.GetToken(tenantDetails.Name, loginInfo.Id(), loginInfo.UserType)
	return &pb.AuthResponse{
		Jwt:      jwtToken,
		UserType: loginInfo.UserType,
		Profile:  profileProto,
	}, nil
}
