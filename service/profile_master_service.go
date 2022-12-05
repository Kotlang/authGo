package service

import (
	"context"
	"strings"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
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
	db *db.AuthDb
}

func NewProfileMasterService(authDB *db.AuthDb) *ProfileMasterService {
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

func (s *ProfileMasterService) BulkGetProfileMaster(ctx context.Context, req *pb.BulkGetProfileMasterRequest) (*pb.ProfileMasterResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

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
