package service

import (
	authPb "github.com/Kotlang/authGo/generated/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// all input validations will be added here.

func ValidateProfileRequest(profileReq *authPb.CreateProfileRequest) error {
	nameLen := len(profileReq.Name)
	if nameLen == 0 {
		return status.Error(codes.InvalidArgument, "Name is required.")
	}
	if nameLen > 50 {
		return status.Error(codes.InvalidArgument, "Name exceeds length of 50 characters.")
	}

	return nil
}
