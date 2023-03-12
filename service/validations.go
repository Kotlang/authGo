package service

import (
	pb "github.com/Kotlang/authGo/generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// all input validations will be added here.

func ValidateProfileRequest(profileReq *pb.CreateProfileRequest) error {
	if len(profileReq.Name) > 50 {
		return status.Error(codes.InvalidArgument, "Name exceeds length of 50 characters.")
	}

	return nil
}
