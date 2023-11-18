package service

import (
	"context"
	"fmt"
	"strings"
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

type ProfileMasterService struct {
	pb.UnimplementedProfileMasterServer
	db db.AuthDbInterface
}

func NewProfileMasterService(authDB db.AuthDbInterface) *ProfileMasterService {
	return &ProfileMasterService{
		db: authDB,
	}
}

func (s *ProfileMasterService) GetProfileMaster(ctx context.Context, req *pb.GetProfileMasterRequest) (*pb.ProfileMasterResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	language := req.Language
	if len(strings.TrimSpace(language)) == 0 {
		language = "english"
	}

	profileMasterListChan, profileMasterListErrorChan := s.db.ProfileMaster(tenant).FindByLanguage(language)
	list := make([]*pb.ProfileMasterProto, 0)

	select {
	case profileMasterList := <-profileMasterListChan:
		copier.CopyWithOption(&list, &profileMasterList, copier.Option{DeepCopy: true})
		return &pb.ProfileMasterResponse{
			ProfileMasterList: list,
		}, nil
	case err := <-profileMasterListErrorChan:
		logger.Error("Failed getting profile master list", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile master list")
	}
}

func (s *ProfileMasterService) GetLanguages(ctx context.Context, req *pb.GetLanguagesRequest) (*pb.LanguagesResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	distinctLanguagesChan, distinctLanguagesErrorChan := s.db.ProfileMaster(tenant).Distinct("language", bson.D{}, 2*time.Second)
	list := make([]string, 0)

	select {
	case distinctLanguages := <-distinctLanguagesChan:
		for _, value := range distinctLanguages {
			res, ok := value.(string)
			if ok {
				list = append(list, res)
			}
		}
		return &pb.LanguagesResponse{
			Languages: list,
		}, nil
	case err := <-distinctLanguagesErrorChan:
		logger.Error("Failed getting distinct languages", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting distinct languages")
	}
}

// ADMIN PORTAL API
func (s *ProfileMasterService) BulkGetProfileMaster(ctx context.Context, req *pb.BulkGetProfileMasterRequest) (*pb.ProfileMasterResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginModelChan, errChan := s.db.Login(tenant).FindOneById(userId)
	select {
	case loginModel := <-loginModelChan:
		fmt.Println(loginModel.Phone)
		if loginModel.UserType != "admin" {
			return nil, status.Error(codes.PermissionDenied, "User with id"+userId+" don't have permission")
		}
	case err := <-errChan:
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	profileMasterListChan, profileMasterListErrorChan := s.db.ProfileMaster(tenant).Find(bson.M{}, nil, 0, 0)
	list := make([]*pb.ProfileMasterProto, 0)

	select {
	case profileMasterList := <-profileMasterListChan:
		copier.CopyWithOption(&list, &profileMasterList, copier.Option{DeepCopy: true})
		return &pb.ProfileMasterResponse{
			ProfileMasterList: list,
		}, nil
	case err := <-profileMasterListErrorChan:
		logger.Error("Failed bulk getting profile master list", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed bulk getting profile master list")
	}
}

// ADMIN PORTAL API
// Delete Profile Master
func (s *ProfileMasterService) DeleteProfileMaster(ctx context.Context, req *pb.DeleteProfileMasterRequest) (*pb.DeleteProfileMasterResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginModelChan, errChan := s.db.Login(tenant).FindOneById(userId)
	select {
	case loginModel := <-loginModelChan:
		if loginModel.UserType != "admin" {
			return nil, status.Error(codes.PermissionDenied, "User with id"+userId+" don't have permission")
		}
	case err := <-errChan:
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	profileMasterChan, errChan := s.db.ProfileMaster(tenant).FindOneById(req.Id)

	select {
	case profileMaster := <-profileMasterChan:
		err := <-s.db.ProfileMaster(tenant).DeleteById(profileMaster.Id())

		if err != nil {
			logger.Error("Internal error when deleting Profile Master with id: "+req.Id, zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		} else {
			return &pb.DeleteProfileMasterResponse{
				Status: "success",
			}, nil
		}
	case err := <-errChan:
		logger.Error("Profile Master not found", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Profile Master not found")
	}
}

// ADMIN PORTAL API
// Add Profile Master
func (s *ProfileMasterService) AddProfileMaster(ctx context.Context, req *pb.AddProfileMasterRequest) (*pb.ProfileMasterProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginModelChan, errChan := s.db.Login(tenant).FindOneById(userId)
	select {
	case loginModel := <-loginModelChan:
		if loginModel.UserType != "admin" {
			return nil, status.Error(codes.PermissionDenied, "User with id"+userId+" don't have permission")
		}
	case err := <-errChan:
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	if len(strings.TrimSpace(req.Language)) == 0 {
		logger.Error("Language is not present")
		return nil, status.Error(codes.InvalidArgument, "Language is not present")
	}
	profileMaster := &models.ProfileMasterModel{}
	copier.CopyWithOption(profileMaster, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	err := <-s.db.ProfileMaster(tenant).Save(profileMaster)

	if err != nil {
		logger.Error("Internal error when saving Profile Master with id: "+profileMaster.Id(), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		profileMasterChan, errChan := s.db.ProfileMaster(tenant).FindOneById(profileMaster.Id())
		select {
		case profileMaster := <-profileMasterChan:
			profileMasterProto := &pb.ProfileMasterProto{}
			copier.Copy(profileMasterProto, profileMaster)
			return profileMasterProto, nil
		case err := <-errChan:
			logger.Error("After saving, the Profile Master not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, "After saving, the Profile Master not found")
		}
	}
}

// ADMIN PORTAL API
// Update Profile Master
func (s *ProfileMasterService) UpdateProfileMaster(ctx context.Context, req *pb.ProfileMasterProto) (*pb.ProfileMasterProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginModelChan, errChan := s.db.Login(tenant).FindOneById(userId)
	select {
	case loginModel := <-loginModelChan:
		if loginModel.UserType != "admin" {
			return nil, status.Error(codes.PermissionDenied, "User with id"+userId+" don't have permission")
		}
	case err := <-errChan:
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	profileMasterChan, errChain := s.db.ProfileMaster(tenant).FindOneById(req.Id)

	select {
	case profileMaster := <-profileMasterChan:
		copier.CopyWithOption(profileMaster, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		profileMaster.Options = req.Options
		err := <-s.db.ProfileMaster(tenant).Save(profileMaster)
		if err != nil {
			logger.Error("Internal error when saving Profile Master with id: "+profileMaster.Id(), zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		} else {
			profileMasterProto := &pb.ProfileMasterProto{}
			copier.Copy(profileMasterProto, profileMaster)
			return profileMasterProto, nil
		}
	case err := <-errChain:
		logger.Error("Can't update Profile Master not found with id: "+req.Id, zap.Error(err))
		return nil, status.Error(codes.NotFound, "Can't update Profile Master not found with id: "+req.Id)
	}
}
