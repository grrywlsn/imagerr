package storage

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"

	"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
    s3Client   *s3.Client
    bucketName string
)

func InitS3() {
    endpoint := os.Getenv("S3_ENDPOINT")
    bucketName = os.Getenv("S3_BUCKET")
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

    // Upload to S3
    _, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket: &bucketName,
        Key:    &storagePath,
        Body:   file,
    })
    if err != nil {
        return "", fmt.Errorf("failed to upload file: %v", err)
    }

    return storagePath, nil
}

func GetFileURL(storagePath string) string {
    return fmt.Sprintf("%s/%s/%s", os.Getenv("S3_ENDPOINT"), bucketName, storagePath)
}

func DeleteFile(storagePath string) error {
    _, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
        Bucket: &bucketName,
        Key:    &storagePath,
    })
    return err
}