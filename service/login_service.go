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

// LoginService provides methods for authenticating and verifying users using OTP.
// LoginService is a service that handles user authentication and login.
// It provides methods to authenticate users, generate and verify OTPs, and manage user sessions.
type LoginService struct {
	pb.UnimplementedLoginServer
	db  *db.AuthDb     // A database client for user authentication data.
	otp *otp.OtpClient // An OTP client for generating and verifying OTPs.
}

// NewLoginService creates a new instance of LoginService.
func NewLoginService(
	authDb *db.AuthDb,
	otp *otp.OtpClient) *LoginService {

	return &LoginService{
		db:  authDb,
		otp: otp,
	}
}

// AuthFuncOverride is a function that overrides the default authentication function for the LoginService.
// This function is used to remove the authentication interceptor for the LoginService.
func (u *LoginService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}

// Login authenticates the user by sending an OTP to their email or phone number.
// It first validates the domain token and then sends the OTP using the OtpClient.
// If the OTP is sent successfully, it returns a StatusResponse with status "success".
// If there is an error, it returns an error with a corresponding status code.
func (s *LoginService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.StatusResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}

	err := s.otp.SendOtp(tenantDetails.Name, req.EmailOrPhone)
	if err != nil {
		return nil, err
	}

	return &pb.StatusResponse{Status: "success"}, nil
}

// Verify verifies the OTP sent to the user's email or phone number.
// It first validates the domain token and then verifies the OTP using the OtpClient.
// If the OTP is verified successfully, it returns an AuthResponse with a JWT token, user type and user profile.
// If there is an error, it returns an error with a corresponding status code.
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
