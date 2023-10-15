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

// ProfileMasterService is a struct that represents the service for managing user profiles.
// It implements the ProfileMasterServer interface from the pb package.
// It contains a pointer to an instance of AuthDb, which is used to interact with the database.
type ProfileMasterService struct {
	pb.UnimplementedProfileMasterServer
	db *db.AuthDb
}

// NewProfileMasterService creates a new instance of the ProfileMasterService struct.
// It takes an instance of the AuthDb struct as a parameter and returns a pointer to the ProfileMasterService struct.
// This service is responsible for handling profile-related operations.
func NewProfileMasterService(authDB *db.AuthDb) *ProfileMasterService {
	return &ProfileMasterService{
		db: authDB,
	}
}

// GetProfileMaster retrieves a list of profile masters based on the given language.
// If the language is not provided, it defaults to English.
// It returns a list of profile masters and an error if the operation fails.
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

// GetLanguages returns a list of distinct languages available in the profile master database for the given tenant.
// The function takes a context and a GetLanguagesRequest as input and returns a LanguagesResponse and an error.
// The LanguagesResponse contains a list of languages available in the database.
// If the function encounters an error, it returns an error with an appropriate error message.
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

// This is ADMIN PORTAL API,
// BulkGetProfileMaster returns a list of all profile masters if the user is an admin.
// It takes a context and a BulkGetProfileMasterRequest as input and returns a ProfileMasterResponse and an error.
// The function checks if the user is an admin and returns an error if not.
// It then retrieves a list of all profile masters and returns it in a ProfileMasterResponse.
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

// This is ADMIN PORTAL API,
// DeleteProfileMaster deletes a profile master record from the database.
// It requires the user to be an admin to perform the operation.
// If the user is not an admin, it returns a permission denied error.
// If the profile master record is not found, it returns a not found error.
// If there is an internal error during the operation, it returns an internal error.
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

// This is ADMIN PORTAL API,
// AddProfileMaster adds a new profile master to the database. It requires the user to have admin permissions.
// The function takes a context and an AddProfileMasterRequest as input and returns a ProfileMasterProto and an error.
// The ProfileMasterProto contains the newly added profile master's information.
// If the user does not have admin permissions, the function returns a PermissionDenied error.
// If the language is not present in the request, the function returns an InvalidArgument error.
// If there is an internal error when saving the profile master, the function returns an Internal error.
// If the profile master is successfully saved, the function returns the newly added profile
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
// UpdateProfileMaster updates the profile master with the given request data for the user identified by the context.
// It returns the updated profile master data if successful, otherwise it returns an error.
// The user must have admin privileges to update the profile master.
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
