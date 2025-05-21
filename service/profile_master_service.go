package service

import (
	"context"
	"strings"
	"time"

	"github.com/Kotlang/authGo/db"
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileMasterService struct {
	authPb.UnimplementedProfileMasterServer
	mongo odm.MongoClient
}

func ProvideProfileMasterService(mongo odm.MongoClient) *ProfileMasterService {
	return &ProfileMasterService{
		mongo: mongo,
	}
}

func (s *ProfileMasterService) GetProfileMaster(ctx context.Context, req *authPb.GetProfileMasterRequest) (*authPb.ProfileMasterResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	language := req.Language
	if len(strings.TrimSpace(language)) == 0 {
		language = "english"
	}

	profileMasterList, err := odm.Await(db.FindByLanguage(ctx, s.mongo, tenant, language))
	if err != nil {
		logger.Error("Failed getting profile master list", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile master list")
	}

	list := make([]*authPb.ProfileMasterProto, 0)
	copier.CopyWithOption(&list, &profileMasterList, copier.Option{DeepCopy: true})
	return &authPb.ProfileMasterResponse{
		ProfileMasterList: list,
	}, nil
}

func (s *ProfileMasterService) GetLanguages(ctx context.Context, req *authPb.GetLanguagesRequest) (*authPb.LanguagesResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	distinctLanguages, err := odm.Await(odm.CollectionOf[db.ProfileMasterModel](s.mongo, tenant).Distinct(ctx, "language", bson.D{}, 2*time.Second))
	if err != nil {
		logger.Error("Failed getting distinct languages", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting distinct languages")
	}

	list := make([]string, 0)
	for _, value := range distinctLanguages {
		res, ok := value.(string)
		if ok {
			list = append(list, res)
		}
	}
	return &authPb.LanguagesResponse{
		Languages: list,
	}, nil
}

// ADMIN PORTAL API
func (s *ProfileMasterService) BulkGetProfileMaster(ctx context.Context, req *authPb.BulkGetProfileMasterRequest) (*authPb.ProfileMasterResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginModel, err := odm.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	if loginModel.UserType != "admin" {
		return nil, status.Error(codes.PermissionDenied, "User with id"+userId+" don't have permission")
	}

	profileMasterList, err := odm.Await(odm.CollectionOf[db.ProfileMasterModel](s.mongo, tenant).Find(ctx, bson.M{}, nil, 0, 0))
	if err != nil {
		logger.Error("Failed getting profile master list", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile master list")
	}

	list := make([]*authPb.ProfileMasterProto, 0)
	copier.CopyWithOption(&list, &profileMasterList, copier.Option{DeepCopy: true})
	return &authPb.ProfileMasterResponse{
		ProfileMasterList: list,
	}, nil
}

// ADMIN PORTAL API
// Add Profile Master
func (s *ProfileMasterService) AddProfileMaster(ctx context.Context, req *authPb.AddProfileMasterRequest) (*authPb.ProfileMasterProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	loginModel, err := odm.Await(odm.CollectionOf[db.LoginModel](s.mongo, tenant).FindOneByID(ctx, userId))
	if err != nil {
		logger.Error("Failed getting login info using id: "+userId, zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting login info using id: "+userId)
	}

	if loginModel.UserType != "admin" {
		return nil, status.Error(codes.PermissionDenied, "User with id"+userId+" don't have permission")
	}

	if len(strings.TrimSpace(req.Language)) == 0 {
		logger.Error("Language is not present")
		return nil, status.Error(codes.InvalidArgument, "Language is not present")
	}
	profileMaster := &db.ProfileMasterModel{}
	copier.CopyWithOption(profileMaster, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	_, err = odm.Await(odm.CollectionOf[db.ProfileMasterModel](s.mongo, tenant).Save(ctx, *profileMaster))

	if err != nil {
		logger.Error("Internal error when saving Profile Master with id: "+profileMaster.Id(), zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		profileMaster, err = odm.Await(odm.CollectionOf[db.ProfileMasterModel](s.mongo, tenant).FindOneByID(ctx, profileMaster.Id()))

		if err != nil {
			logger.Error("Failed getting profile master list", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed getting profile master list")
		}

		profileMasterProto := &authPb.ProfileMasterProto{}
		copier.Copy(profileMasterProto, profileMaster)
		return profileMasterProto, nil
	}
}
