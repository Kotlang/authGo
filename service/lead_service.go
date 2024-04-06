package service

import (
	"context"

	"github.com/Kotlang/authGo/db"
	authPb "github.com/Kotlang/authGo/generated/auth"
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LeadServiceInterface interface {
	authPb.LeadServiceServer
	db.AuthDbInterface
}

type LeadService struct {
	authPb.UnimplementedLeadServiceServer
	db db.AuthDbInterface
}

func NewLeadService(db db.AuthDbInterface) *LeadService {
	return &LeadService{db: db}
}

// Admin only API
func (s *LeadService) CreateLead(ctx context.Context, req *authPb.CreateOrUpdateLeadRequest) (*authPb.LeadProto, error) {

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
func (s *LeadService) GetLeadById(ctx context.Context, req *authPb.LeadIdRequest) (*authPb.LeadProto, error) {
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
func (s *LeadService) BulkGetLeadsById(ctx context.Context, req *authPb.BulkIdRequest) (*authPb.LeadListResponse, error) {
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
		leadProtos := make([]*authPb.LeadProto, len(leads))
		for i, lead := range leads {
			leadProtos[i] = getLeadProto(&lead)
		}
		return &authPb.LeadListResponse{Leads: leadProtos}, nil
	case err := <-errChan:
		logger.Error("Error getting leads", zap.Error(err))
		return nil, err

	}
}

// Admin only API
func (s *LeadService) UpdateLead(ctx context.Context, req *authPb.CreateOrUpdateLeadRequest) (*authPb.LeadProto, error) {
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
func (s *LeadService) DeleteLead(ctx context.Context, req *authPb.LeadIdRequest) (*authPb.StatusResponse, error) {
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

	return &authPb.StatusResponse{
		Status: "Success",
	}, nil
}

// Admin only API
func (s *LeadService) FetchLeads(ctx context.Context, req *authPb.FetchLeadsRequest) (*authPb.LeadListResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// check if user is admin
	if !s.db.Login(tenant).IsAdmin(userId) {
		logger.Error("User is not admin", zap.String("userId", userId))
		return nil, status.Error(codes.PermissionDenied, "User is not admin")
	}

	if req.PageNumber < 0 {
		req.PageNumber = 0
	}

	if req.PageSize < 0 {
		req.PageSize = 10
	}

	// get the leads from db
	leads, totalCount := s.db.Lead(tenant).GetLeads(req.LeadFilters, int64(req.PageSize), int64(req.PageNumber))

	leadProtos := make([]*authPb.LeadProto, len(leads))
	for i, lead := range leads {
		leadProtos[i] = getLeadProto(&lead)
	}

	return &authPb.LeadListResponse{
		Leads:      leadProtos,
		TotalLeads: int64(totalCount),
	}, nil

}

func getLeadModel(req *authPb.CreateOrUpdateLeadRequest) *models.LeadModel {

	// copying the request to lead model
	lead := &models.LeadModel{}
	copier.CopyWithOption(lead, req, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	// Copy the operator type
	if req.OperatorType != authPb.OperatorType_UNSPECIFIED_OPERATOR {
		value, ok := authPb.OperatorType_name[int32(req.OperatorType)]
		if !ok {
			lead.OperatorType = authPb.OperatorType_name[int32(authPb.OperatorType_UNSPECIFIED_OPERATOR)]
		}
		lead.OperatorType = value
	}

	// Copy the lead channel
	if req.Channel != authPb.LeadChannel_UNSPECIFIED_CHANNEL {
		value, ok := authPb.LeadChannel_name[int32(req.Channel)]
		if !ok {
			lead.Channel = authPb.LeadChannel_name[int32(authPb.LeadChannel_UNSPECIFIED_CHANNEL)]
		}
		lead.Channel = value
	}

	// Copy the farming type
	if req.FarmingType != authPb.FarmingType_UnspecifiedFarming {
		value, ok := authPb.FarmingType_name[int32(req.FarmingType)]
		if !ok {
			value = authPb.FarmingType_name[int32(authPb.FarmingType_UnspecifiedFarming)]
		}
		lead.FarmingType = value
	}

	// Copy the land size
	if req.LandSizeInAcres != authPb.LandSizeInAcres_UnspecifiedLandSize {
		value, ok := authPb.LandSizeInAcres_name[int32(req.LandSizeInAcres)]
		if !ok {
			value = authPb.LandSizeInAcres_name[int32(authPb.LandSizeInAcres_UnspecifiedLandSize)]
		}
		lead.LandSizeInAcres = value
	}

	// Todo: Copy the lead status
	return lead
}

func getLeadProto(lead *models.LeadModel) *authPb.LeadProto {
	leadProto := &authPb.LeadProto{}
	copier.CopyWithOption(leadProto, lead, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	// Copy the operator type
	value, ok := authPb.OperatorType_value[lead.OperatorType]
	if !ok {
		leadProto.OperatorType = authPb.OperatorType_UNSPECIFIED_OPERATOR
	}
	leadProto.OperatorType = authPb.OperatorType(value)

	// Copy the lead channel
	value, ok = authPb.LeadChannel_value[lead.Channel]
	if !ok {
		leadProto.Channel = authPb.LeadChannel_UNSPECIFIED_CHANNEL
	}
	leadProto.Channel = authPb.LeadChannel(value)

	// Copy the farming type
	value, ok = authPb.FarmingType_value[lead.FarmingType]
	if !ok {
		leadProto.FarmingType = authPb.FarmingType_UnspecifiedFarming
	}
	leadProto.FarmingType = authPb.FarmingType(value)

	// Copy the certification details
	value, ok = authPb.LandSizeInAcres_value[lead.LandSizeInAcres]
	if !ok {
		leadProto.LandSizeInAcres = authPb.LandSizeInAcres_UnspecifiedLandSize
	}
	leadProto.LandSizeInAcres = authPb.LandSizeInAcres(value)

	// TODO: Copy the lead status

	return leadProto
}
