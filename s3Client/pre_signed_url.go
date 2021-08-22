package s3client

import (
	"fmt"
	"time"

	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

// Returns pre-signed Upload Url and download Url.
func GetPresignedUrlForProfilePic(tenant, userId, extension string) (string, string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1")},
	)
	if err != nil {
		logger.Error("Error getting aws session", zap.Error(err))
		return "", ""
	}

	// Create S3 service client
	svc := s3.New(sess)

	imagePath := fmt.Sprintf("profile-images/%s/%s/%d.%s", tenant, userId, time.Now().Unix(), extension)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("kotlang-assets"),
		Key:    aws.String(imagePath),
	})
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		logger.Error("Error signing s3 url", zap.Error(err))
		return "", ""
	}

	downloadUrl := fmt.Sprintf("https://kotlang-assets.s3.ap-south-1.amazonaws.com/%s", imagePath)
	return urlStr, downloadUrl
}
