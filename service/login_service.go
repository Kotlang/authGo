package service

import (
	"context"
	"unicode"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
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
	db  db.AuthDbInterface
	otp otp.OtpClientInterface
}

func NewLoginService(
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

func (s *LoginService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.StatusResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid domain token")
	}

	if !(req.RestoreAccountRequest) {
		// check if user is marked for deletion and return status as failure.
		isPhone := isPhoneNumber(req.EmailOrPhone)
		phoneNumber, email := "", ""
		if isPhone {
			phoneNumber = req.EmailOrPhone
		} else {
			email = req.EmailOrPhone
		}

		loginDetails := <-s.db.Login(tenantDetails.Name).FindOneByPhoneOrEmail(phoneNumber, email)
		if loginDetails != nil {
			profileResChan, profileErrChan := s.db.Profile(tenantDetails.Name).FindOneById(loginDetails.Id())
			select {
			case profile := <-profileResChan:
				if profile.DeletionInfo.MarkedForDeletion {
					return &pb.StatusResponse{Status: "Marked for deletion"}, nil
				}
			case err := <-profileErrChan:
				logger.Error("Error fetching profile", zap.Error(err))
			}
		}

		if req.BlockUnknown && loginDetails == nil {
			return nil, status.Error(codes.NotFound, "User does not exist")
		}

	}

	err := s.otp.SendOtp(tenantDetails.Name, req.EmailOrPhone)
	if err != nil {
		return nil, err
	}

	return &pb.StatusResponse{Status: "success"}, nil
}

func (s *LoginService) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.AuthResponse, error) {
	if len(req.Domain) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Domain Token")
	}

	tenantDetails := <-s.db.Tenant().FindOneByToken(req.Domain)
	if tenantDetails == nil {
		return nil, status.Error(codes.PermissionDenied, "Invalid Domain token")
	}

	loginInfo := s.otp.GetLoginInfo(tenantDetails.Name, req.EmailOrPhone)
	if loginInfo == nil || !s.otp.ValidateOtp(tenantDetails.Name, req.EmailOrPhone, req.Otp) {
		return nil, status.Error(codes.PermissionDenied, "Wrong OTP")
	}

	// fetch profile for user.
	profileProto := &pb.UserProfileProto{}
	print("loginInfo.Id() ", loginInfo.UserId)
	resultChan, errorChan := s.db.Profile(tenantDetails.Name).FindOneById(loginInfo.Id())
	select {
	case profile := <-resultChan:

		if profile.DeletionInfo.MarkedForDeletion {
			profile.DeletionInfo = models.DeletionInfo{MarkedForDeletion: false}
			err := <-s.db.Profile(tenantDetails.Name).Save(profile)

			if err != nil {
				logger.Error("Internal error when saving Profile with id: "+profile.Id(), zap.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

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

func isPhoneNumber(emailOrPhone string) bool {
	for _, c := range emailOrPhone {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return len(emailOrPhone) == 10
}
