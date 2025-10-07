package server

import (
	"context"
	"strings"
	"time"

	knowledgev1 "demo/ms_knowledge/api/proto/v1"
	"demo/ms_knowledge/internal/domain"
	"demo/ms_knowledge/internal/service"
	userv1 "demo/ms_user/api/proto/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCServer struct {
	knowledgev1.UnimplementedKnowledgeServiceServer

	spaceService   *service.SpaceService
	contentService *service.ContentService
	userClient     userv1.UserServiceClient
}

func NewGRPCServer(spaceService *service.SpaceService, contentService *service.ContentService) *GRPCServer {
	return &GRPCServer{
		spaceService:   spaceService,
		contentService: contentService,
	}
}

// WithUserClient allows injecting a user service client (useful for startup wiring and tests).
func (s *GRPCServer) WithUserClient(c userv1.UserServiceClient) *GRPCServer {
	s.userClient = c
	return s
}

// Space Management
func (s *GRPCServer) CreateSpace(ctx context.Context, req *knowledgev1.CreateSpaceRequest) (*knowledgev1.Space, error) {
	// Auth interceptor has already validated the user and injected context values
	space, err := s.spaceService.CreateSpace(ctx, req.Title, req.Description, req.Keywords, req.Icon)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create space: %v", err)
	}

	return s.domainSpaceToProto(space), nil
}

func (s *GRPCServer) GetSpace(ctx context.Context, req *knowledgev1.GetSpaceRequest) (*knowledgev1.Space, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid space id: %v", err)
	}

	space, err := s.spaceService.GetSpace(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get space: %v", err)
	}

	return s.domainSpaceWithStatsToProto(space), nil
}

func (s *GRPCServer) ListSpaces(ctx context.Context, req *knowledgev1.ListSpacesRequest) (*knowledgev1.ListSpacesResponse, error) {
	var ownerID uuid.UUID
	if req.OwnerId != "" {
		var err error
		ownerID, err = uuid.Parse(req.OwnerId)
		// Tolerate non-UUID owner ids (e.g., external auth IDs) by ignoring the filter
		if err != nil {
			ownerID = uuid.Nil
		}
	}

	pageSize := 0
	if req.Page != nil && req.Page.PageSize > 0 {
		pageSize = int(req.Page.PageSize)
	}

	filter := domain.SpaceFilter{
		OwnerID: ownerID,
		Query:   req.Q,
		Limit:   pageSize,
		Offset:  0, // TODO: Implement pagination with page token
	}

	spaces, err := s.spaceService.ListSpaces(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list spaces: %v", err)
	}

	protoSpaces := make([]*knowledgev1.Space, len(spaces))
	for i, space := range spaces {
		protoSpaces[i] = s.domainSpaceWithStatsToProto(space)
	}

	return &knowledgev1.ListSpacesResponse{
		Items: protoSpaces,
	}, nil
}

func (s *GRPCServer) UpdateSpace(ctx context.Context, req *knowledgev1.UpdateSpaceRequest) (*knowledgev1.Space, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid space id: %v", err)
	}
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Keywords != nil {
		updates["keywords"] = req.Keywords
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.AccessLevel != "" {
		updates["access_level"] = req.AccessLevel
	}
	// SECURITY: Do not allow changing owner_id via this endpoint to prevent privilege escalation.
	// If ownership transfer is needed, implement a dedicated admin-authorized flow.

	space, err := s.spaceService.UpdateSpace(ctx, id, updates)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update space: %v", err)
	}

	return s.domainSpaceToProto(space), nil
}

func (s *GRPCServer) DeleteSpace(ctx context.Context, req *knowledgev1.DeleteSpaceRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid space id: %v", err)
	}
	err = s.spaceService.DeleteSpace(ctx, id, req.HardDelete, req.Force)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete space: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// Content Source Management
func (s *GRPCServer) CreateUploadURL(ctx context.Context, req *knowledgev1.CreateUploadURLRequest) (*knowledgev1.CreateUploadURLResponse, error) {
	spaceID, err := uuid.Parse(req.SpaceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid space id: %v", err)
	}

	content, uploadURL, objectKey, expiresAt, err := s.contentService.CreateUploadURL(ctx, spaceID, req.Filename, req.MimeType, req.SizeBytes, req.Title, req.ObjectKind)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create upload URL: %v", err)
	}

	return &knowledgev1.CreateUploadURLResponse{
		UploadUrl:     uploadURL,
		ObjectKey:     objectKey,
		ExpiresAt:     timestamppb.New(expiresAt),
		ContentSource: s.domainContentSourceToProto(content),
	}, nil
}

func (s *GRPCServer) ConfirmUpload(ctx context.Context, req *knowledgev1.ConfirmUploadRequest) (*knowledgev1.ContentSource, error) {
	contentID, err := uuid.Parse(req.ContentSourceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content source id: %v", err)
	}

	if strings.TrimSpace(req.BlobHash) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "blob_hash is required")
	}

	// Use the standard ConfirmUpload method
	content, err := s.contentService.ConfirmUpload(ctx, contentID, req.BlobHash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to confirm upload: %v", err)
	}

	return s.domainContentSourceToProto(content), nil
}

func (s *GRPCServer) GetContentSource(ctx context.Context, req *knowledgev1.GetContentSourceRequest) (*knowledgev1.ContentSource, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content source id: %v", err)
	}

	content, err := s.contentService.GetContentSource(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get content source: %v", err)
	}

	return s.domainContentSourceToProto(content), nil
}

func (s *GRPCServer) DeleteContentSource(ctx context.Context, req *knowledgev1.DeleteContentSourceRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content source id: %v", err)
	}

	if err := s.contentService.DeleteContentSource(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete content source: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *GRPCServer) ListContentSources(ctx context.Context, req *knowledgev1.ListContentSourcesRequest) (*knowledgev1.ListContentSourcesResponse, error) {
	spaceID, err := uuid.Parse(req.SpaceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid space id: %v", err)
	}

	pageSize := 0
	if req.Page != nil && req.Page.PageSize > 0 {
		pageSize = int(req.Page.PageSize)
	}

	// Build status string but keep assignment inside struct literal to avoid
	// mutating the filter after construction (keeps initialization atomic).
	statusStr := ""
	if req.Status != knowledgev1.ContentStatus_CONTENT_STATUS_UNSPECIFIED {
		statusStr = req.Status.String()
	}

	filter := domain.ContentSourceFilter{
		SpaceID: spaceID,
		Status:  domain.ContentStatus(statusStr),
		Limit:   pageSize,
		Offset:  0, // TODO: Implement pagination with page token
	}

	contents, err := s.contentService.ListContentSources(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list content sources: %v", err)
	}

	protoContents := make([]*knowledgev1.ContentSource, len(contents))
	for i, content := range contents {
		protoContents[i] = s.domainContentSourceToProto(content)
	}

	return &knowledgev1.ListContentSourcesResponse{
		Items: protoContents,
	}, nil
}

func (s *GRPCServer) UpdateContentSourceStatus(ctx context.Context, req *knowledgev1.UpdateContentSourceStatusRequest) (*knowledgev1.ContentSource, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content source id: %v", err)
	}

	content, err := s.contentService.UpdateContentSourceStatus(
		ctx,
		id,
		domain.ContentStatus(req.Status.String()),
		req.ProcessedBlobHash,
		req.ErrorMessage,
		nil,
		nil,
		req.Title,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update content source status: %v", err)
	}

	return s.domainContentSourceToProto(content), nil
}

func (s *GRPCServer) UpdateContentSource(ctx context.Context, req *knowledgev1.UpdateContentSourceRequest) (*knowledgev1.ContentSource, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content source id: %v", err)
	}

	// Determine which fields to update based on update_mask
	var titlePtr *string
	var keywordsPtr *[]string

	if req.UpdateMask != nil && len(req.UpdateMask.Paths) > 0 {
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "title":
				t := strings.TrimSpace(req.Content.GetTitle())
				titlePtr = &t
			case "keywords":
				ks := make([]string, len(req.Content.GetKeywords()))
				copy(ks, req.Content.GetKeywords())
				keywordsPtr = &ks
			default:
				return nil, status.Errorf(codes.InvalidArgument, "unsupported path in update_mask: %s", path)
			}
		}
	} else {
		// No mask: allow partial semantics based on presence in payload
		if req.Content != nil {
			if v := strings.TrimSpace(req.Content.GetTitle()); v != "" {
				titlePtr = &v
			}
			// For keywords, use presence: if provided (even empty), apply
			if req.Content.Keywords != nil {
				ks := make([]string, len(req.Content.GetKeywords()))
				copy(ks, req.Content.GetKeywords())
				keywordsPtr = &ks
			}
		}
	}

	content, err := s.contentService.UpdateContentSource(ctx, id, titlePtr, keywordsPtr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update content source: %v", err)
	}
	return s.domainContentSourceToProto(content), nil
}

// GenerateDownloadURL returns a presigned URL for downloading the requested object.
func (s *GRPCServer) GenerateDownloadURL(ctx context.Context, req *knowledgev1.GenerateDownloadURLRequest) (*knowledgev1.GenerateDownloadURLResponse, error) {
	contentID, err := uuid.Parse(req.ContentSourceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content source id: %v", err)
	}

	// TTL bounds
	ttl := 15 * time.Minute
	if req.ExpiresSeconds > 0 {
		if req.ExpiresSeconds < 30 {
			req.ExpiresSeconds = 30
		}
		if req.ExpiresSeconds > 3600 {
			req.ExpiresSeconds = 3600
		}
		ttl = time.Duration(req.ExpiresSeconds) * time.Second
	}

	// Use the simplified method that accepts kind directly
	url, expiresAt, err := s.contentService.GenerateDownloadURL(ctx, contentID, req.ObjectKind, ttl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate download URL: %v", err)
	}

	// Fetch content to compute object key; RLS is enforced in service layer
	content, err := s.contentService.GetContentSource(ctx, contentID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get content source: %v", err)
	}

	// Generate object key using consistent path format
	objectKey := content.OwnerID.String() + "/spaces/" + content.SpaceID.String() + "/content/" + content.ID.String() + "/" + content.Source

	return &knowledgev1.GenerateDownloadURLResponse{
		Url:       url,
		ExpiresAt: timestamppb.New(expiresAt),
		ObjectKey: objectKey,
	}, nil
}

// Utilities
func (s *GRPCServer) Healthz(ctx context.Context, req *emptypb.Empty) (*knowledgev1.HealthStatus, error) {
	components := make(map[string]string)
	overallStatus := "OK"
	// Check database health
	if err := s.checkDatabase(ctx); err != nil {
		components["database"] = "FAIL: " + err.Error()
		overallStatus = "FAIL"
	} else {
		components["database"] = "OK"
	}
	// Check storage health
	if err := s.checkStorage(ctx); err != nil {
		components["storage"] = "FAIL: " + err.Error()
		overallStatus = "FAIL"
	} else {
		components["storage"] = "OK"
	}
	return &knowledgev1.HealthStatus{
		Status:     overallStatus,
		Components: components,
	}, nil
}

// checkDatabase performs a simple health check for the database.
func (s *GRPCServer) checkDatabase(ctx context.Context) error {
	// Try a simple operation, e.g., list spaces with a limit of 1
	filter := domain.SpaceFilter{
		Limit: 1,
	}
	_, err := s.spaceService.ListSpaces(ctx, filter)
	return err
}

// checkStorage performs a simple health check for the storage service.
func (s *GRPCServer) checkStorage(ctx context.Context) error {
	// Try a simple operation, e.g., list contents with a limit of 1
	filter := domain.ContentSourceFilter{
		Limit: 1,
	}
	_, err := s.contentService.ListContentSources(ctx, filter)
	return err
}

// Helper methods for converting between domain and proto types
func (s *GRPCServer) domainSpaceToProto(space *domain.Space) *knowledgev1.Space {
	return &knowledgev1.Space{
		Id:          space.ID.String(),
		Title:       space.Title,
		Description: space.Description,
		OwnerId:     space.OwnerID.String(),
		Icon:        space.Icon,
		Keywords:    space.Keywords,
		AccessLevel: space.AccessLevel,
		CreatedAt:   timestamppb.New(space.CreatedAt),
		UpdatedAt:   timestamppb.New(space.LastUpdatedAt),
	}
}

func (s *GRPCServer) domainSpaceWithStatsToProto(space *domain.SpaceWithStats) *knowledgev1.Space {
	protoSpace := s.domainSpaceToProto(&space.Space)
	// Ensure denormalized counters are exposed to clients
	protoSpace.DocumentCount = space.Stats.ContentCount
	protoSpace.TotalSizeBytes = space.Stats.TotalSizeBytes
	protoSpace.Stats = &knowledgev1.SpaceStats{
		ContentCount:   space.Stats.ContentCount,
		LinkCount:      space.Stats.LinkCount,
		LastActivityAt: timestamppb.New(space.Stats.LastActivityAt),
	}
	return protoSpace
}

func (s *GRPCServer) domainContentSourceToProto(content *domain.ContentSource) *knowledgev1.ContentSource {
	var processedBlobHash string
	if content.ProcessedBlobHash != nil {
		processedBlobHash = *content.ProcessedBlobHash
	}

	var contentSummary string
	if content.ContentSummary != nil {
		contentSummary = *content.ContentSummary
	}

	var keywords []string
	if content.Keywords != nil {
		keywords = content.Keywords
	}

	return &knowledgev1.ContentSource{
		Id:                content.ID.String(),
		SpaceId:           content.SpaceID.String(),
		Status:            knowledgev1.ContentStatus(knowledgev1.ContentStatus_value[string(content.Status)]),
		OwnerId:           content.OwnerID.String(),
		Source:            content.Source,
		MimeType:          content.MediaType,
		SizeBytes:         content.SizeBytes,
		Title:             content.Title,
		OriginalBlobHash:  content.OriginalBlobHash,
		ProcessedBlobHash: processedBlobHash,
		ContentSummary:    contentSummary,
		Keywords:          keywords,
		CreatedAt:         timestamppb.New(content.CreatedAt),
		UpdatedAt:         timestamppb.New(content.UpdatedAt),
	}
}
