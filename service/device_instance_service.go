package service

import (
	"context"
	"strings"

	"github.com/Kotlang/authGo/db"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Kotlang/authGo/generated"
)

type DeviceInstanceService struct {
	pb.UnimplementedDeviceInstanceServer
	db *db.AuthDb
}

func NewDeviceInstanceService(db *db.AuthDb) *DeviceInstanceService {
	return &DeviceInstanceService{
		db: db,
	}
}

func (s *DeviceInstanceService) RegisterDeviceInstance(ctx context.Context, req *pb.RegisterDeviceInstanceRequest) (*pb.RegisterDeviceInstanceResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(strings.TrimSpace(userId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Token is not present.")
	}

	deviceInstance := &models.DeviceInstanceModel{
		LoginId: userId,
		Token:   req.Token,
	}

	err := <-s.db.DeviceInstance(tenant).Save(deviceInstance)
	if err != nil {
		return nil, err
	} else {
		return &pb.RegisterDeviceInstanceResponse{}, nil
	}
}
