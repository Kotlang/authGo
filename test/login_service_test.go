package test

import (
	"context"
	"testing"

	"github.com/Kotlang/authGo/db"
	pb "github.com/Kotlang/authGo/generated"
	"github.com/Kotlang/authGo/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Kotlang/authGo/service"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {

	var loginTestCases = []struct {
		desc              string
		request           *pb.LoginRequest
		expectedResponse  *pb.StatusResponse
		expectedErrorCode codes.Code // Expected gRPC error code
		expectedError     error      // Expected error message
	}{
		{
			desc: "Valid request",
			request: &pb.LoginRequest{
				EmailOrPhone: "9970378011",
				Domain:       "valid_domain",
			},
			expectedResponse:  &pb.StatusResponse{Status: "success"},
			expectedErrorCode: codes.OK,
			expectedError:     nil,
		},
		{
			desc: "Empty Domain",
			request: &pb.LoginRequest{
				EmailOrPhone: "9970378011",
				Domain:       "", // Empty domain should trigger an error
			},
			expectedResponse:  nil,
			expectedErrorCode: codes.InvalidArgument,
			expectedError:     status.Error(codes.InvalidArgument, "Invalid Domain Token"),
		},
		{
			desc: "Invalid Domain Token",
			request: &pb.LoginRequest{
				EmailOrPhone: "9970378011",
				Domain:       "invalidDomain", // A non-existent domain should trigger an error
			},
			expectedResponse:  nil,
			expectedErrorCode: codes.PermissionDenied,
			expectedError:     status.Error(codes.PermissionDenied, "Invalid domain token"),
		},
		{
			desc: "Invalid Email or Phone",
			request: &pb.LoginRequest{
				EmailOrPhone: "", // Empty email or phone should trigger an error
				Domain:       "valid_domain",
			},
			expectedResponse:  nil,
			expectedErrorCode: codes.InvalidArgument,
			expectedError:     status.Error(codes.InvalidArgument, "Incorrect email or phone"),
		},
		// {
		// 	desc: "Valid request but within 60 sec",
		// 	request: &pb.LoginRequest{
		// 		EmailOrPhone: "9970378011",
		// 		Domain:       "valid_domain",
		// 	},
		// 	expectedResponse:  nil,
		// 	expectedErrorCode: codes.PermissionDenied,
		// 	expectedError:     status.Error(codes.PermissionDenied, "Exceeded threshold of OTPs in a minute."),
		// },
	}

	for _, testData := range loginTestCases {
		t.Run(testData.desc, func(t *testing.T) {
			ctx := context.TODO()
			loginService := service.NewLoginService(&db.AuthDb{}, mocks.NewMockOtpClient(&db.AuthDb{}))
			response, err := loginService.Login(ctx, testData.request)
			assert.Equal(t, testData.expectedResponse, response)

			if err != nil {
				assert.EqualError(t, err, testData.expectedError.Error())
			}
		})

	}

}

func TestLoginService_Verify(t *testing.T) {

	loginService := service.NewLoginService(&db.AuthDb{}, mocks.NewMockOtpClient(&db.AuthDb{}))
	//The sendOtp treats 123456 as validOtp and all other otp as invalid
	var verifyTestCases = []struct {
		name             string
		request          *pb.VerifyRequest
		expectedError    error
		expectedResponse *pb.AuthResponse
	}{
		{
			name: "Valid input",
			request: &pb.VerifyRequest{
				Domain:       "valid_domain",
				EmailOrPhone: "9970378011",
				Otp:          "123456",
			},
			expectedError:    nil,
			expectedResponse: &pb.AuthResponse{},
		},
		{
			name: "Invalid Domain",
			request: &pb.VerifyRequest{
				Domain:       "",
				EmailOrPhone: "9970378011",
				Otp:          "123456",
			},
			expectedError:    status.Error(codes.InvalidArgument, "Invalid Domain Token"),
			expectedResponse: nil,
		},
		{
			name: "Invalid domain token in the database",
			request: &pb.VerifyRequest{
				Domain:       "invalid_domain",
				EmailOrPhone: "9970378011",
				Otp:          "123456",
			},
			expectedError:    status.Error(codes.PermissionDenied, "Invalid domain token"),
			expectedResponse: nil,
		},
		{
			name: "Number on which otp is not sent",
			request: &pb.VerifyRequest{
				Domain:       "valid_domain",
				EmailOrPhone: "9970378012",
				Otp:          "123456",
			},
			expectedError:    status.Error(codes.PermissionDenied, "Wrong OTP"),
			expectedResponse: nil,
		},
		{
			name: "Invalid OTP",
			request: &pb.VerifyRequest{
				Domain:       "valid_domain",
				EmailOrPhone: "9970378011",
				Otp:          "12456",
			},
			expectedError:    status.Error(codes.PermissionDenied, "Wrong OTP"),
			expectedResponse: nil,
		},
	}

	//send otp to a number
	loginService.Login(context.TODO(), &pb.LoginRequest{
		EmailOrPhone: "9970378011",
		Domain:       "valid_domain",
	})

	ctx := context.TODO()
	for _, testData := range verifyTestCases {
		response, err := loginService.Verify(ctx, testData.request)

		if testData.expectedResponse != nil {
			assert.NotNil(t, response)
		} else {
			assert.Nil(t, response)
		}

		if err != nil {
			assert.EqualError(t, err, testData.expectedError.Error())
		}
	}
}
