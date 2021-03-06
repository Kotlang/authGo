package service

import (
	"context"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/otp"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginService struct {
	pb.UnimplementedLoginServer
	db  *db.AuthDb
	otp *otp.OtpClient
}

func NewLoginService(
	authDb *db.AuthDb,
	otp *otp.OtpClient) *LoginService {

	return &LoginService{
		db:  authDb,
		otp: otp,
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

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}
	s.otp.SendOtp(tenantDetails.Name, req.EmailOrPhone)
	return &pb.StatusResponse{Status: "success"}, nil
}

func (s *LoginService) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.AuthResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}

	loginInfo := s.otp.GetLoginInfo(tenantDetails.Name, req.EmailOrPhone)
	if loginInfo == nil || !s.otp.ValidateOtp(tenantDetails.Name, req.EmailOrPhone, req.Otp) {
		return nil, status.Error(codes.PermissionDenied, "Wrong OTP")
	}

	// fetch profile for user.
	profileProto := &pb.UserProfileProto{}
	resultChan, errorChan := s.db.Profile(tenantDetails.Name).FindOneById(loginInfo.Id())
	select {
	case profile := <-resultChan:
		copier.Copy(profileProto, profile)
	case err := <-errorChan:
		logger.Error("Error fetching profile", zap.Error(err))
	}

	// copy login info to profile even if profile is not present.
	copier.CopyWithOption(profileProto, loginInfo, copier.Option{IgnoreEmpty: true})

	jwtToken := auth.GetToken(tenantDetails.Name, loginInfo.Id(), loginInfo.UserType)
	return &pb.AuthResponse{
		Jwt:      jwtToken,
		UserType: loginInfo.UserType,
		Profile:  profileProto,
	}, nil
}
