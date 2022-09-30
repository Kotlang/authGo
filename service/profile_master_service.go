package service

import (
	"context"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
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
	if language == "" {
		language = "english"
	}

	profileMasterListChan, profileMasterListErrorChan := s.db.ProfileMaster(tenant).FindByLanguage(req.Language)
	list := make([]*pb.ProfileMasterProto, 0)

	select {
	case profileMasterList := <-profileMasterListChan:
		for _, profileMaster := range profileMasterList {
			list = append(list, &pb.ProfileMasterProto{
				Field:   profileMaster.Field,
				Type:    profileMaster.Type,
				Options: profileMaster.Options,
			})
		}
		return &pb.ProfileMasterResponse{
			ProfileMasterList: list,
		}, nil
	case err := <-profileMasterListErrorChan:
		logger.Error("Failed getting profile master list", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting profile master list")
	}
}
