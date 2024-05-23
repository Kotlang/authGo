package service

import (
	"context"
	"os"
	"unicode"

	"github.com/Kotlang/authGo/db"
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/Kotlang/authGo/otp"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginService struct {
	authPb.UnimplementedLoginServer
	db  db.AuthDbInterface
	otp otp.OtpClientInterface
}

func ProvideLoginService(
	authDb db.AuthDbInterface,
	otp otp.OtpClientInterface) *LoginService {

	return &LoginService{
		db:  authDb,
		otp: otp,
	}
}

// removing auth interceptor
func (u *LoginService) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}

// removing the check user existence interceptor
func (u *LoginService) CheckUserExistenceOverride(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (s *LoginService) Login(ctx context.Context, req *authPb.LoginRequest) (*authPb.StatusResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}

	// get login details by phone or email
	isPhone := isPhoneNumber(req.EmailOrPhone)
	var loginDetails *db.LoginModel
	if isPhone {
		loginDetails = <-s.db.Login(tenantDetails.Name).FindOneByPhoneOrEmail(req.EmailOrPhone, "")
	} else {
		loginDetails = <-s.db.Login(tenantDetails.Name).FindOneByPhoneOrEmail("", req.EmailOrPhone)
	}

	// check if user is blocked, if yes return error
	if loginDetails != nil && loginDetails.IsBlocked {
		return nil, status.Error(codes.PermissionDenied, "User is blocked")
	}

	// if user does not exist and block unknown is true, return error
	if req.BlockUnknown && loginDetails == nil {
		return nil, status.Error(codes.NotFound, "User does not exist")
	}

	// if request is not to restore account and account is marked for deletion, return error
	if loginDetails != nil && !req.RestoreAccountRequest {

		// if user is marked for deletion, return error
		if loginDetails.DeletionInfo.MarkedForDeletion {
			return nil, status.Error(codes.PermissionDenied, "User is marked for deletion")
		}
	}

	// if phone number is excluded from verification, donot send otp and return success
	phoneNumberToExcludeVerification := os.Getenv("PHONE_NUMBER_TO_EXCLUDE_VERIFICATION")
	if req.EmailOrPhone == phoneNumberToExcludeVerification {
		return &authPb.StatusResponse{Status: "success"}, nil
	}

	// send otp
	err := s.otp.SendOtp(tenantDetails.Name, req.EmailOrPhone)
	if err != nil {
		return nil, err
	}
	return &authPb.StatusResponse{Status: "success"}, nil
}

func (s *LoginService) Verify(ctx context.Context, req *authPb.VerifyRequest) (*authPb.AuthResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid Domain token")
	}

	loginInfo := s.otp.GetLoginInfo(tenantDetails.Name, req.EmailOrPhone)

	// if phone number is excluded from verification, donot send otp
	phoneNumberToExcludeVerification := os.Getenv("PHONE_NUMBER_TO_EXCLUDE_VERIFICATION")
	if req.EmailOrPhone != phoneNumberToExcludeVerification {
		if loginInfo == nil || !s.otp.ValidateOtp(tenantDetails.Name, req.EmailOrPhone, req.Otp) {
			return nil, status.Error(codes.PermissionDenied, "Wrong OTP")
		}
	}

	// if phone number is excluded from verification check if otp is 666666
	if req.EmailOrPhone == phoneNumberToExcludeVerification && req.Otp != "666666" {
		return nil, status.Error(codes.PermissionDenied, "Wrong OTP")
	}

	// if user is blocked return error
	if loginInfo != nil && loginInfo.IsBlocked {
		return nil, status.Error(codes.PermissionDenied, "User is blocked")
	}

	// if deletion info is marked for deletion, update the deletion info
	if loginInfo != nil && loginInfo.DeletionInfo.MarkedForDeletion {
		loginInfo.DeletionInfo.MarkedForDeletion = false
		loginInfo.DeletionInfo.Reason = ""
		loginInfo.DeletionInfo.DeletionTime = 0

		// save the login info
		err := <-s.db.Login(tenantDetails.Name).Save(loginInfo)

		if err != nil {
			logger.Error("Error saving login info", zap.Error(err))
		}
	}

	// fetch profile for user.
	profileProto := &authPb.UserProfileProto{}
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
	return &authPb.AuthResponse{
		Jwt:      jwtToken,
		UserType: loginInfo.UserType,
		Profile:  profileProto,
	}, nil
}

func isPhoneNumber(emailOrPhone string) bool {
	for _, c := range emailOrPhone {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return len(emailOrPhone) == 10
}
