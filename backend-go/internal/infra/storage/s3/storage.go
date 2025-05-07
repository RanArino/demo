package s3

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/ran/demo/backend-go/internal/core/models/document"
)

// S3StorageService implements the StoragePort interface using AWS S3
type S3StorageService struct {
	client *S3Client
	logger *slog.Logger
}

// NewS3StorageService creates a new S3 storage service
func NewS3StorageService(client *S3Client, logger *slog.Logger) *S3StorageService {
	return &S3StorageService{
		client: client,
		logger: logger,
	}
}

// Upload stores document content and returns storage metadata
func (s *S3StorageService) Upload(
	ctx context.Context,
	content *document.DocumentContent,
	options *document.StorageOptions,
) (*document.StorageObject, error) {
	// Generate a unique key if not provided
	bucket := options.Bucket
	if bucket == "" {
		bucket = s.client.GetConfig().Bucket
	}

	// Generate a unique key for the object
	key := fmt.Sprintf("documents/%s/%s", time.Now().Format("2006-01-02"), uuid.New().String())
	if options.MetadataAttributes != nil {
		if docID, ok := options.MetadataAttributes["document_id"]; ok {
			key = fmt.Sprintf("documents/%s/%s", docID, content.OriginalFilename)
		}
	}

	// Create metadata for the object
	metadata := make(map[string]string)
	if options.MetadataAttributes != nil {
		for k, v := range options.MetadataAttributes {
			metadata[k] = v
		}
	}
	metadata["original_filename"] = content.OriginalFilename
	metadata["content_type"] = content.MIMEType
	metadata["upload_date"] = time.Now().Format(time.RFC3339)

	// Upload the object
	putObjectInput := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          content.Stream,
		ContentType:   aws.String(content.MIMEType),
		ContentLength: aws.Int64(content.Size),
		Metadata:      metadata,
	}

	if options.StorageClass != "" {
		putObjectInput.StorageClass = s3types.StorageClass(options.StorageClass)
	}

	if options.PubliclyAccessible {
		putObjectInput.ACL = s3types.ObjectCannedACLPublicRead
	}

	if options.ExpiresAfter != nil {
		expiresAt := time.Now().Add(*options.ExpiresAfter)
		putObjectInput.Expires = aws.Time(expiresAt)
	}

	result, err := s.client.GetClient().PutObject(ctx, putObjectInput)
	if err != nil {
		return nil, document.NewStorageError(err, "upload", key, "s3")
	}

	// Create storage object with metadata
	storageObj := &document.StorageObject{
		ID:           document.StorageObjectID(key),
		Key:          key,
		Bucket:       bucket,
		Size:         content.Size,
		ContentType:  content.MIMEType,
		ETag:         *result.ETag,
		LastModified: time.Now(),
		Provider:     document.StorageProviderS3,
		Metadata:     metadata,
	}

	return storageObj, nil
}

// Download retrieves document content from storage
func (s *S3StorageService) Download(
	ctx context.Context,
	objectID document.StorageObjectID,
) (*document.DocumentContent, error) {
	// Get object with a key
	key := string(objectID)
	bucket := s.client.GetConfig().Bucket

	result, err := s.client.GetClient().GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, document.NewStorageError(err, "download", key, "s3")
	}

	// Get original filename from metadata
	filename := key
	if filenameFromMeta, ok := result.Metadata["original_filename"]; ok {
		filename = filenameFromMeta
	}

	// Create document content
	content := &document.DocumentContent{
		OriginalFilename: filename,
		MIMEType:         *result.ContentType,
		Stream:           result.Body,
		Size:             *result.ContentLength,
	}

	return content, nil
}

// DownloadToWriter downloads object content directly to a writer
func (s *S3StorageService) DownloadToWriter(
	ctx context.Context,
	objectID document.StorageObjectID,
	writer io.Writer,
) error {
	// Download the content first
	content, err := s.Download(ctx, objectID)
	if err != nil {
		return err
	}

	// Check if we can close the stream using type assertion
	if closer, ok := content.Stream.(io.ReadCloser); ok {
		defer closer.Close()
	}

	// Copy from the content stream to the writer
	_, err = io.Copy(writer, content.Stream)
	if err != nil {
		return document.NewStorageError(err, "stream", string(objectID), "s3")
	}

	return nil
}

// GetPresignedUploadURL generates a pre-signed URL for client-direct uploads
func (s *S3StorageService) GetPresignedUploadURL(
	ctx context.Context,
	filename string,
	contentType string,
	options *document.StorageOptions,
) (*document.StorageUploadInfo, error) {
	bucket := options.Bucket
	if bucket == "" {
		bucket = s.client.GetConfig().Bucket
	}

	// Generate a unique key for the object
	key := fmt.Sprintf("documents/%s/%s", time.Now().Format("2006-01-02"), uuid.New().String())
	if options.MetadataAttributes != nil {
		if docID, ok := options.MetadataAttributes["document_id"]; ok {
			key = fmt.Sprintf("documents/%s/%s", docID, filename)
		}
	}

	// Create metadata for the object
	metadata := make(map[string]string)
	if options.MetadataAttributes != nil {
		for k, v := range options.MetadataAttributes {
			metadata[k] = v
		}
	}
	metadata["original_filename"] = filename
	metadata["content_type"] = contentType
	metadata["upload_date"] = time.Now().Format(time.RFC3339)

	// Create presigned URL input
	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	}

	if options.StorageClass != "" {
		putObjectInput.StorageClass = s3types.StorageClass(options.StorageClass)
	}

	if options.PubliclyAccessible {
		putObjectInput.ACL = s3types.ObjectCannedACLPublicRead
	}

	if options.ExpiresAfter != nil {
		expiresAt := time.Now().Add(*options.ExpiresAfter)
		putObjectInput.Expires = aws.Time(expiresAt)
	}

	// Create presigned URL
	presignDuration := time.Duration(s.client.GetConfig().PresignedURLDuration) * time.Second
	if presignDuration == 0 {
		presignDuration = 15 * time.Minute
	}

	presignResult, err := s.client.GetPresignClient().PresignPutObject(ctx, putObjectInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = presignDuration
		})
	if err != nil {
		return nil, document.NewStorageError(err, "presign-upload", key, "s3")
	}

	// Create storage upload info
	uploadInfo := &document.StorageUploadInfo{
		PresignedURL: presignResult.URL,
		ExpiresAt:    time.Now().Add(presignDuration),
		Key:          key,
		Bucket:       bucket,
	}

	// Add form fields for POST policy if needed
	// This is a simplification - a full implementation would support POST policy for S3 browser uploads
	return uploadInfo, nil
}

// GetPresignedDownloadURL generates a pre-signed URL for client-direct downloads
func (s *S3StorageService) GetPresignedDownloadURL(
	ctx context.Context,
	objectID document.StorageObjectID,
	filename string,
	expiresInSeconds int,
) (*document.StorageDownloadInfo, error) {
	key := string(objectID)
	bucket := s.client.GetConfig().Bucket

	// Create GetObject input
	getObjectInput := &s3.GetObjectInput{
		Bucket:                     aws.String(bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", filename)),
	}

	// Create presigned URL
	presignDuration := time.Duration(expiresInSeconds) * time.Second
	if presignDuration == 0 {
		presignDuration = 15 * time.Minute
	}

	presignResult, err := s.client.GetPresignClient().PresignGetObject(ctx, getObjectInput,
		func(opts *s3.PresignOptions) {
			opts.Expires = presignDuration
		})
	if err != nil {
		return nil, document.NewStorageError(err, "presign-download", key, "s3")
	}

	// Get object metadata to include in the response
	headResult, err := s.client.GetClient().HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Warn("Failed to get object metadata", "error", err, "key", key)
		// Continue anyway, as the presigned URL is still valid
	}

	// Create storage download info
	contentType := "application/octet-stream"
	var size int64 = 0
	if headResult != nil {
		if headResult.ContentType != nil {
			contentType = *headResult.ContentType
		}
		if headResult.ContentLength != nil {
			size = *headResult.ContentLength
		}
	}

	downloadInfo := &document.StorageDownloadInfo{
		PresignedURL: presignResult.URL,
		ExpiresAt:    time.Now().Add(presignDuration),
		ContentType:  contentType,
		Size:         size,
		Filename:     filename,
	}

	return downloadInfo, nil
}

// Delete removes an object from storage
func (s *S3StorageService) Delete(ctx context.Context, objectID document.StorageObjectID) error {
	key := string(objectID)
	bucket := s.client.GetConfig().Bucket

	_, err := s.client.GetClient().DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return document.NewStorageError(err, "delete", key, "s3")
	}

	return nil
}

// Copy duplicates an object to a new location/key
func (s *S3StorageService) Copy(
	ctx context.Context,
	sourceID document.StorageObjectID,
	destOptions *document.StorageOptions,
) (*document.StorageObject, error) {
	sourceBucket := s.client.GetConfig().Bucket
	sourceKey := string(sourceID)

	destBucket := sourceBucket
	if destOptions.Bucket != "" {
		destBucket = destOptions.Bucket
	}

	// Generate a destination key
	destKey := fmt.Sprintf("documents/copy_%s", uuid.New().String())
	if destOptions.MetadataAttributes != nil {
		if docID, ok := destOptions.MetadataAttributes["document_id"]; ok {
			if filename, ok := destOptions.MetadataAttributes["filename"]; ok {
				destKey = fmt.Sprintf("documents/%s/%s", docID, filename)
			} else {
				destKey = fmt.Sprintf("documents/%s/copy_%s", docID, uuid.New().String())
			}
		}
	}

	// Create copy input
	copySource := fmt.Sprintf("%s/%s", sourceBucket, sourceKey)
	copyObjectInput := &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
	}

	if destOptions.ContentType != "" {
		copyObjectInput.ContentType = aws.String(destOptions.ContentType)
	}

	if destOptions.StorageClass != "" {
		copyObjectInput.StorageClass = s3types.StorageClass(destOptions.StorageClass)
	}

	if destOptions.PubliclyAccessible {
		copyObjectInput.ACL = s3types.ObjectCannedACLPublicRead
	}

	if destOptions.MetadataAttributes != nil {
		metadata := make(map[string]string)
		for k, v := range destOptions.MetadataAttributes {
			metadata[k] = v
		}
		copyObjectInput.Metadata = metadata
		copyObjectInput.MetadataDirective = s3types.MetadataDirectiveReplace
	}

	// Copy the object
	result, err := s.client.GetClient().CopyObject(ctx, copyObjectInput)
	if err != nil {
		return nil, document.NewStorageError(err, "copy", sourceKey, "s3")
	}

	// Get metadata from HeadObject to include in the response
	headResult, err := s.client.GetClient().HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(destBucket),
		Key:    aws.String(destKey),
	})
	if err != nil {
		s.logger.Warn("Failed to get copied object metadata", "error", err, "key", destKey)
		// Continue anyway and return partial info
	}

	// Create storage object with metadata
	storageObj := &document.StorageObject{
		ID:           document.StorageObjectID(destKey),
		Key:          destKey,
		Bucket:       destBucket,
		ETag:         *result.CopyObjectResult.ETag,
		LastModified: time.Now(),
		Provider:     document.StorageProviderS3,
	}

	if headResult != nil {
		if headResult.ContentType != nil {
			storageObj.ContentType = *headResult.ContentType
		}
		if headResult.ContentLength != nil {
			storageObj.Size = *headResult.ContentLength
		}
		if headResult.Metadata != nil {
			storageObj.Metadata = headResult.Metadata
		}
	}

	return storageObj, nil
}

// GetObjectMetadata retrieves metadata without downloading content
func (s *S3StorageService) GetObjectMetadata(
	ctx context.Context,
	objectID document.StorageObjectID,
) (*document.StorageObject, error) {
	key := string(objectID)
	bucket := s.client.GetConfig().Bucket

	headResult, err := s.client.GetClient().HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, document.NewStorageError(err, "get-metadata", key, "s3")
	}

	// Create storage object with metadata
	storageObj := &document.StorageObject{
		ID:           objectID,
		Key:          key,
		Bucket:       bucket,
		Provider:     document.StorageProviderS3,
		LastModified: *headResult.LastModified,
		Metadata:     headResult.Metadata,
	}

	if headResult.ContentType != nil {
		storageObj.ContentType = *headResult.ContentType
	}
	if headResult.ContentLength != nil {
		storageObj.Size = *headResult.ContentLength
	}
	if headResult.ETag != nil {
		storageObj.ETag = *headResult.ETag
	}

	return storageObj, nil
}

// ListObjectsByPrefix lists objects with a specific prefix (path)
func (s *S3StorageService) ListObjectsByPrefix(
	ctx context.Context,
	prefix string,
	maxItems int,
) ([]*document.StorageObject, error) {
	bucket := s.client.GetConfig().Bucket

	if maxItems <= 0 || maxItems > 1000 {
		maxItems = 100 // Default and max limit for S3 ListObjects
	}

	// List objects
	listResult, err := s.client.GetClient().ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(int32(maxItems)),
	})
	if err != nil {
		return nil, document.NewStorageError(err, "list", prefix, "s3")
	}

	// Create storage objects from results
	objects := make([]*document.StorageObject, 0, len(listResult.Contents))
	for _, obj := range listResult.Contents {
		size := int64(0)
		if obj.Size != nil {
			size = *obj.Size
		}

		// Create storage object
		storageObj := &document.StorageObject{
			ID:           document.StorageObjectID(*obj.Key),
			Key:          *obj.Key,
			Bucket:       bucket,
			Size:         size,
			ETag:         *obj.ETag,
			LastModified: *obj.LastModified,
			Provider:     document.StorageProviderS3,
		}

		objects = append(objects, storageObj)
	}

	return objects, nil
}

// ConvertFileFormat attempts to convert a file from one format to another
func (s *S3StorageService) ConvertFileFormat(
	ctx context.Context,
	sourceID document.StorageObjectID,
	targetFormat string,
) (document.StorageObjectID, error) {
	// For the initial implementation, file format conversion would require additional services
	// For simplicity, we'll just return an error indicating it's not implemented yet
	return "", fmt.Errorf("file format conversion not implemented yet")
}
