package service

import (
	"context"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LeadServiceInterface interface {
	pb.LeadServiceServer
	db.AuthDbInterface
}

type LeadService struct {
	pb.UnimplementedLeadServiceServer
	db db.AuthDbInterface
}

func NewLeadService(db db.AuthDbInterface) *LeadService {
	return &LeadService{db: db}
}

// Admin only API
func (s *LeadService) CreateLead(ctx context.Context, req *pb.CreateOrUpdateLeadRequest) (*pb.LeadProto, error) {

	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}
	// get the lead model from the request
	lead := getLeadModel(req)

	// save to db
	err := <-s.db.Lead(tenant).Save(lead)

	if err != nil {
		logger.Error("Error saving lead", zap.Error(err))
		return nil, err
	}

	// return the lead
	leadProto := getLeadProto(lead)

	return leadProto, nil
}

// Admin only API
func (s *LeadService) GetLeadById(ctx context.Context, req *pb.LeadIdRequest) (*pb.LeadProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}

	// get the lead from db
	leadResChan, errChan := s.db.Lead(tenant).FindOneById(req.LeadId)
	select {
	case lead := <-leadResChan:
		leadProto := getLeadProto(lead)
		return leadProto, nil
	case err := <-errChan:
		logger.Error("Error getting lead", zap.Error(err))
		return nil, err
	}
}

// Admin only API
func (s *LeadService) BulkGetLeadsById(ctx context.Context, req *pb.BulkIdRequest) (*pb.LeadListResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}

	// get the leads from db
	leadResChan, errChan := s.db.Lead(tenant).FindByIds(req.LeadIds)
	select {
	case leads := <-leadResChan:
		leadProtos := make([]*pb.LeadProto, len(leads))
		for i, lead := range leads {
			leadProtos[i] = getLeadProto(&lead)
		}
		return &pb.LeadListResponse{Leads: leadProtos}, nil
	case err := <-errChan:
		logger.Error("Error getting leads", zap.Error(err))
		return nil, err

	}
}

// Admin only API
func (s *LeadService) UpdateLead(ctx context.Context, req *pb.CreateOrUpdateLeadRequest) (*pb.LeadProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}

	// get the lead model from the request
	lead := getLeadModel(req)

	// save to db
	err := <-s.db.Lead(tenant).Save(lead)

	if err != nil {
		logger.Error("Error saving lead", zap.Error(err))
		return nil, err
	}

	// return the lead
	leadProto := getLeadProto(lead)

	return leadProto, nil
}

// Admin only API
func (s *LeadService) DeleteLead(ctx context.Context, req *pb.LeadIdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}

	// delete the lead
	err := <-s.db.Lead(tenant).DeleteById(req.LeadId)
	if err != nil {
		logger.Error("Error deleting lead", zap.Error(err))
		return nil, err
	}

	return &pb.StatusResponse{
		Status: "Success",
	}, nil
}

// Admin only API
func (s *LeadService) FetchLeads(ctx context.Context, req *pb.FetchLeadsRequest) (*pb.LeadListResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}

	// get the leads from db
	leads, totalCount := s.db.Lead(tenant).GetLeads(req.LeadFilters, int64(req.PageNumber), int64(req.PageSize))

	leadProtos := make([]*pb.LeadProto, len(leads))
	for i, lead := range leads {
		leadProtos[i] = getLeadProto(&lead)
	}

	return &pb.LeadListResponse{
		Leads:      leadProtos,
		TotalLeads: int64(totalCount),
	}, nil

}

func getLeadModel(req *pb.CreateOrUpdateLeadRequest) *models.LeadModel {

	// copying the request to lead model
	lead := &models.LeadModel{}
	copier.CopyWithOption(lead, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	// Copy the operator type
	if req.OperatorType != pb.OperatorType_UNSPECIFIED_OPERATOR {
		value, ok := pb.OperatorType_name[int32(req.OperatorType)]
		if !ok {
			lead.OperatorType = pb.OperatorType_name[int32(pb.OperatorType_UNSPECIFIED_OPERATOR)]
		}
		lead.OperatorType = value
	}

	// Copy the lead channel
	if req.Channel != pb.LeadChannel_UNSPECIFIED_CHANNEL {
		value, ok := pb.LeadChannel_name[int32(req.Channel)]
		if !ok {
			lead.Channel = pb.LeadChannel_name[int32(pb.LeadChannel_UNSPECIFIED_CHANNEL)]
		}
		lead.Channel = value
	}

	// Copy the farming type
	if req.FarmingType != pb.FarmingType_UnspecifiedFarming {
		value, ok := pb.FarmingType_name[int32(req.FarmingType)]
		if !ok {
			value = pb.FarmingType_name[int32(pb.FarmingType_UnspecifiedFarming)]
		}
		lead.FarmingType = value
	}

	// Copy the land size
	if req.LandSizeInAcres != pb.LandSizeInAcres_UnspecifiedLandSize {
		value, ok := pb.LandSizeInAcres_name[int32(req.LandSizeInAcres)]
		if !ok {
			value = pb.LandSizeInAcres_name[int32(pb.LandSizeInAcres_UnspecifiedLandSize)]
		}
		lead.LandSizeInAcres = value
	}

	// Todo: Copy the lead status
	return lead
}

func getLeadProto(lead *models.LeadModel) *pb.LeadProto {
	leadProto := &pb.LeadProto{}
	copier.CopyWithOption(leadProto, lead, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	// Copy the operator type
	value, ok := pb.OperatorType_value[lead.OperatorType]
	if !ok {
		leadProto.OperatorType = pb.OperatorType_UNSPECIFIED_OPERATOR
	}
	leadProto.OperatorType = pb.OperatorType(value)

	// Copy the lead channel
	value, ok = pb.LeadChannel_value[lead.Channel]
	if !ok {
		leadProto.Channel = pb.LeadChannel_UNSPECIFIED_CHANNEL
	}
	leadProto.Channel = pb.LeadChannel(value)

	// Copy the farming type
	value, ok = pb.FarmingType_value[lead.FarmingType]
	if !ok {
		leadProto.FarmingType = pb.FarmingType_UnspecifiedFarming
	}
	leadProto.FarmingType = pb.FarmingType(value)

	// Copy the certification details
	value, ok = pb.LandSizeInAcres_value[lead.LandSizeInAcres]
	if !ok {
		leadProto.LandSizeInAcres = pb.LandSizeInAcres_UnspecifiedLandSize
	}
	leadProto.LandSizeInAcres = pb.LandSizeInAcres(value)

	// TODO: Copy the lead status

	return leadProto
}
