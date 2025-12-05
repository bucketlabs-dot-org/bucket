package storage

import (
    "context"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type B2Client struct {
    PresignClient *s3.PresignClient
    Bucket        string
}

func NewB2Client(bucket, endpoint, key, secret string) (*B2Client, error) {

    awsCfg, err := config.LoadDefaultConfig(
        context.Background(),
        config.WithRegion("us-west-1"),
        config.WithCredentialsProvider(
            credentials.NewStaticCredentialsProvider(key, secret, ""),
        ),
        config.WithEndpointResolverWithOptions(
            aws.EndpointResolverWithOptionsFunc(
                func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                    return aws.Endpoint{
                        URL:               endpoint,
                        SigningRegion:     "us-west-1",
                        HostnameImmutable: true,
                    }, nil
                }),
        ),
    )
    if err != nil {
        return nil, err
    }

    s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
        o.UsePathStyle = true
    })

    presigner := s3.NewPresignClient(s3Client)

    return &B2Client{
        PresignClient: presigner,
        Bucket:        bucket,
    }, nil
}

func (b *B2Client) GenerateUploadURL(objectPath string, expires time.Duration) (string, error) {
    resp, err := b.PresignClient.PresignPutObject(
        context.Background(),
        &s3.PutObjectInput{
            Bucket: aws.String(b.Bucket),
            Key:    aws.String(objectPath),
        },
        func(opts *s3.PresignOptions) {
            opts.Expires = expires
        },
    )
    if err != nil {
        return "", err
    }
    return resp.URL, nil
}

func (b *B2Client) GenerateDownloadURL(objectPath string, expires time.Duration) (string, error) {
    resp, err := b.PresignClient.PresignGetObject(
        context.Background(),
        &s3.GetObjectInput{
            Bucket: aws.String(b.Bucket),
            Key:    aws.String(objectPath),
        },
        func(opts *s3.PresignOptions) {
            opts.Expires = expires
        },
    )
    if err != nil {
        return "", err
    }
    return resp.URL, nil
}
