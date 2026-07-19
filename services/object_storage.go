package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	appconfig "finalproject/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

var ErrObjectStorageNotConfigured = errors.New("object storage is not configured")

type ObjectStorageError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (err *ObjectStorageError) Error() string {
	parts := []string{"object storage request failed"}
	if err.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("status=%d", err.StatusCode))
	}
	if err.Code != "" {
		parts = append(parts, "code="+err.Code)
	}
	if err.Message != "" {
		parts = append(parts, fmt.Sprintf("message=%q", err.Message))
	}
	if len(parts) > 1 {
		return strings.Join(parts, " ")
	}

	return fmt.Sprintf("object storage request failed: %v", err.Err)
}

func (err *ObjectStorageError) Unwrap() error {
	return err.Err
}

type ObjectUploadInput struct {
	Key         string
	ContentType string
	Body        io.Reader
	Size        int64
}

type ObjectUploadResult struct {
	URL         string
	Key         string
	Bucket      string
	ContentType string
	Size        int64
}

type ObjectDownloadResult struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	CacheControl  string
	ETag          string
}

func UploadObject(ctx context.Context, cfg appconfig.Config, input ObjectUploadInput) (ObjectUploadResult, error) {
	if !cfg.ObjectStorageConfigured() {
		return ObjectUploadResult{}, ErrObjectStorageNotConfigured
	}

	client, err := newObjectStorageClient(ctx, cfg)
	if err != nil {
		return ObjectUploadResult{}, fmt.Errorf("load s3 config: %w", err)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(cfg.S3Bucket),
		Key:           aws.String(input.Key),
		Body:          input.Body,
		ContentType:   aws.String(input.ContentType),
		ContentLength: aws.Int64(input.Size),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return ObjectUploadResult{}, objectStorageRequestError(err)
	}

	return ObjectUploadResult{
		URL:         objectURL(cfg, input.Key),
		Key:         input.Key,
		Bucket:      cfg.S3Bucket,
		ContentType: input.ContentType,
		Size:        input.Size,
	}, nil
}

func GetObject(ctx context.Context, cfg appconfig.Config, key string) (ObjectDownloadResult, error) {
	if !cfg.ObjectStorageConfigured() {
		return ObjectDownloadResult{}, ErrObjectStorageNotConfigured
	}

	client, err := newObjectStorageClient(ctx, cfg)
	if err != nil {
		return ObjectDownloadResult{}, fmt.Errorf("load s3 config: %w", err)
	}

	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.S3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectDownloadResult{}, objectStorageRequestError(err)
	}

	return ObjectDownloadResult{
		Body:          result.Body,
		ContentType:   aws.ToString(result.ContentType),
		ContentLength: aws.ToInt64(result.ContentLength),
		CacheControl:  aws.ToString(result.CacheControl),
		ETag:          aws.ToString(result.ETag),
	}, nil
}

func newObjectStorageClient(ctx context.Context, cfg appconfig.Config) (*s3.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKeyID,
			cfg.S3SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.S3Endpoint)
		options.UsePathStyle = cfg.S3ForcePathStyle
		options.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		options.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return client, nil
}

func objectStorageRequestError(err error) error {
	storageErr := &ObjectStorageError{
		Err: err,
	}

	var responseErr *smithyhttp.ResponseError
	if errors.As(err, &responseErr) {
		storageErr.StatusCode = responseErr.HTTPStatusCode()
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		storageErr.Code = apiErr.ErrorCode()
		storageErr.Message = apiErr.ErrorMessage()
	}

	return storageErr
}

func objectURL(cfg appconfig.Config, key string) string {
	escapedKey := escapeObjectKey(key)

	if cfg.S3PublicBaseURL != "" {
		return cfg.S3PublicBaseURL + "/" + escapedKey
	}

	endpoint := strings.TrimRight(cfg.S3Endpoint, "/")
	if cfg.S3ForcePathStyle {
		return endpoint + "/" + url.PathEscape(cfg.S3Bucket) + "/" + escapedKey
	}

	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" {
		return endpoint + "/" + url.PathEscape(cfg.S3Bucket) + "/" + escapedKey
	}

	parsed.Host = cfg.S3Bucket + "." + parsed.Host
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/" + key
	return parsed.String()
}

func escapeObjectKey(key string) string {
	segments := strings.Split(strings.TrimLeft(key, "/"), "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}

	return strings.Join(segments, "/")
}
