package storage

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
    s3Client   *s3.Client
    bucketName string
)

func InitS3() {
    endpoint := os.Getenv("S3_ENDPOINT")
    bucketName = os.Getenv("S3_BUCKET_NAME")
    region := os.Getenv("S3_REGION")

    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(region),
        config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
            func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                return aws.Endpoint{
                    URL:               endpoint,
                    SigningRegion:    region,
                    HostnameImmutable: true,
                }, nil
            })),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            os.Getenv("S3_ACCESS_KEY"),
            os.Getenv("S3_SECRET_KEY"),
            "",
        )),
    )
    if err != nil {
        log.Fatal("Unable to load SDK config:", err)
    }

    s3Client = s3.NewFromConfig(cfg)
}

func UploadFile(file io.Reader, filename string) (string, error) {
    // Generate unique path for the file
    storagePath := fmt.Sprintf("images/%s", filename)

    // Upload to S3 with public-read ACL
    _, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: &bucketName,
        Key:    &storagePath,
        Body:   file,
        ACL:    types.ObjectCannedACLPublicRead,
    })
    if err != nil {
        return "", fmt.Errorf("failed to upload file: %v", err)
    }

    return storagePath, nil
}

func GetFileURL(storagePath string) string {
    cdnDomain := os.Getenv("CDN_DOMAIN")
    if cdnDomain == "" {
        log.Fatal("CDN_DOMAIN environment variable is required")
    }
    cdnDomain = strings.TrimRight(cdnDomain, "/")
    return fmt.Sprintf("%s/%s", cdnDomain, storagePath)
}

func DeleteFile(storagePath string) error {
    _, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
        Bucket: &bucketName,
        Key:    &storagePath,
    })
    return err
}