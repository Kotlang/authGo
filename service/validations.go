package service

import (
	pb "github.com/Kotlang/authGo/generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// all input validations will be added here.

// ValidateProfileRequest validates the CreateProfileRequest protobuf message.
// It checks if the name field is not empty and does not exceed 50 characters.
// If the validation fails, it returns an error with a corresponding status code.
// Otherwise, it returns nil.
func ValidateProfileRequest(profileReq *pb.CreateProfileRequest) error {
	nameLen := len(profileReq.Name)
	if nameLen == 0 {
		return status.Error(codes.InvalidArgument, "Name is required.")
	}
	if nameLen > 50 {
		return status.Error(codes.InvalidArgument, "Name exceeds length of 50 characters.")
	}

	return nil
}
