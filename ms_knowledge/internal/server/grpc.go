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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCServer struct {
	knowledgev1.UnimplementedKnowledgeServiceServer

	spaceService         *service.SpaceService
	contentService       *service.ContentService
	knowledgeLinkService *service.KnowledgeLinkService
	userClient           userv1.UserServiceClient
}

func NewGRPCServer(spaceService *service.SpaceService, contentService *service.ContentService, knowledgeLinkService *service.KnowledgeLinkService) *GRPCServer {
	return &GRPCServer{
		spaceService:         spaceService,
		contentService:       contentService,
		knowledgeLinkService: knowledgeLinkService,
	}
}

// WithUserClient allows injecting a user service client (useful for startup wiring and tests).
func (s *GRPCServer) WithUserClient(c userv1.UserServiceClient) *GRPCServer {
	s.userClient = c
	return s
}

// Space Management
func (s *GRPCServer) CreateSpace(ctx context.Context, req *knowledgev1.CreateSpaceRequest) (*knowledgev1.Space, error) {
	// Derive owner_id server-side. Prefer resolving via ms_user using the incoming JWT.
	ctxOwner := ctx
	if s.userClient != nil {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				// Inject resolved internal user id into context for services
				ctxOwner = context.WithValue(ctx, domain.ContextKey("owner_id"), uResp.GetUser().GetId())
			}
		}
	}

	space, err := s.spaceService.CreateSpace(ctxOwner, req.Title, req.Description)
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

	// RBAC/Ownership: admins bypass, others must own
	if role, ok := ctx.Value(domain.RoleKey).(string); !(ok && role == "admin") {
		if md, ok := metadata.FromIncomingContext(ctx); ok && s.userClient != nil {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				ownerID := uResp.GetUser().GetId()
				sp, _ := s.spaceService.GetSpace(ctx, id)
				if sp != nil && sp.OwnerID.String() != ownerID {
					return nil, status.Errorf(codes.PermissionDenied, "not owner")
				}
			}
		}
	}

	updates := make(map[string]interface{})
	if req.Space.GetTitle() != "" {
		updates["title"] = req.Space.GetTitle()
	}
	if req.Space.GetDescription() != "" {
		updates["description"] = req.Space.GetDescription()
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

	// RBAC/Ownership: admins bypass, others must own
	if role, ok := ctx.Value(domain.RoleKey).(string); !(ok && role == "admin") {
		if md, ok := metadata.FromIncomingContext(ctx); ok && s.userClient != nil {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				ownerID := uResp.GetUser().GetId()
				sp, _ := s.spaceService.GetSpace(ctx, id)
				if sp != nil && sp.OwnerID.String() != ownerID {
					return nil, status.Errorf(codes.PermissionDenied, "not owner")
				}
			}
		}
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

	// RBAC/Ownership: admins bypass, others must own the space
	if role, ok := ctx.Value(domain.RoleKey).(string); !(ok && role == "admin") {
		if md, ok := metadata.FromIncomingContext(ctx); ok && s.userClient != nil {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				ownerID := uResp.GetUser().GetId()
				sp, _ := s.spaceService.GetSpace(ctx, spaceID)
				if sp != nil && sp.OwnerID.String() != ownerID {
					return nil, status.Errorf(codes.PermissionDenied, "not owner")
				}
			}
		}
	}

	// Map object_kind enum to kind string
	var kind string
	switch req.ObjectKind {
	case knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_ORIGINAL:
		kind = "source"
	case knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_PROCESSED:
		kind = "processed"
	default:
		return nil, status.Errorf(codes.InvalidArgument, "object_kind is required")
	}

	content, uploadURL, objectKey, expiresAt, err := s.contentService.CreateUploadURLWithKind(ctx, spaceID, req.Filename, req.MimeType, req.SizeBytes, req.Title, kind)
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

	// Validate kind and map
	var kind string
	switch req.ObjectKind {
	case knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_ORIGINAL:
		kind = "source"
	case knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_PROCESSED:
		kind = "processed"
	default:
		return nil, status.Errorf(codes.InvalidArgument, "object_kind is required")
	}
	if strings.TrimSpace(req.BlobHash) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "blob_hash is required")
	}

	content, err := s.contentService.ConfirmUploadWithKind(ctx, contentID, kind, req.BlobHash)
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

	// RBAC/Ownership: admins bypass, others must own the content source
	if role, ok := ctx.Value(domain.RoleKey).(string); !(ok && role == "admin") {
		if md, ok := metadata.FromIncomingContext(ctx); ok && s.userClient != nil {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				ownerID := uResp.GetUser().GetId()
				cs, _ := s.contentService.GetContentSource(ctx, id)
				if cs != nil && cs.OwnerID.String() != ownerID {
					return nil, status.Errorf(codes.PermissionDenied, "not owner")
				}
			}
		}
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

	content, err := s.contentService.UpdateContentSourceStatus(ctx, id, domain.ContentStatus(req.Status.String()), req.ProcessedBlobHash, req.ErrorMessage)
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

	// RBAC/Ownership: admins bypass, others must own the content
	if role, ok := ctx.Value(domain.RoleKey).(string); !(ok && role == "admin") {
		if md, ok := metadata.FromIncomingContext(ctx); ok && s.userClient != nil {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				callerID := uResp.GetUser().GetId()
				cs, _ := s.contentService.GetContentSource(ctx, id)
				if cs != nil && cs.OwnerID.String() != callerID {
					return nil, status.Errorf(codes.PermissionDenied, "not owner")
				}
			}
		}
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

	// Validate object_kind early
	var kind string
	switch req.ObjectKind {
	case knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_ORIGINAL:
		kind = "source"
	case knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_PROCESSED:
		kind = "processed"
	default:
		return nil, status.Errorf(codes.InvalidArgument, "object_kind is required")
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

	// Fetch content for RBAC checks and to compute object key
	content, err := s.contentService.GetContentSource(ctx, contentID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get content source: %v", err)
	}

	// RBAC/Ownership: admins bypass, others must own the content (defense in depth)
	if role, ok := ctx.Value(domain.RoleKey).(string); !(ok && role == "admin") {
		if md, ok := metadata.FromIncomingContext(ctx); ok && s.userClient != nil {
			outCtx := metadata.NewOutgoingContext(ctx, md)
			uReq := &userv1.GetUserRequest{}
			if uResp, err := s.userClient.GetUser(outCtx, uReq); err == nil && uResp.GetUser() != nil {
				callerID := uResp.GetUser().GetId()
				if content.OwnerID.String() != callerID {
					return nil, status.Errorf(codes.PermissionDenied, "not owner")
				}
			}
		}
	}

	url, expiresAt, err := s.contentService.GenerateDownloadURLWithKind(ctx, contentID, ttl, kind)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate download URL: %v", err)
	}

	objectKey := "spaces/" + content.SpaceID.String() + "/content/" + content.ID.String() + "/" + content.Source

	return &knowledgev1.GenerateDownloadURLResponse{
		Url:       url,
		ExpiresAt: timestamppb.New(expiresAt),
		ObjectKey: objectKey,
	}, nil
}

// Knowledge Link Management
func (s *GRPCServer) CreateKnowledgeLink(ctx context.Context, req *knowledgev1.CreateKnowledgeLinkRequest) (*knowledgev1.KnowledgeLink, error) {
	fromID, err := uuid.Parse(req.FromContentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from content id: %v", err)
	}

	toID, err := uuid.Parse(req.ToContentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to content id: %v", err)
	}

	link, err := s.knowledgeLinkService.CreateKnowledgeLink(ctx, fromID, toID, domain.RelationType(req.RelationType.String()), req.Weight)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create knowledge link: %v", err)
	}

	return s.domainKnowledgeLinkToProto(link), nil
}

// Get a single enriched link (includes previews)
func (s *GRPCServer) GetKnowledgeLink(ctx context.Context, req *knowledgev1.GetKnowledgeLinkRequest) (*knowledgev1.EnrichedKnowledgeLink, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid link id: %v", err)
	}

	el, err := s.knowledgeLinkService.GetKnowledgeLink(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get knowledge link: %v", err)
	}

	return &knowledgev1.EnrichedKnowledgeLink{
		Id:           el.ID.String(),
		From:         &knowledgev1.ContentPreview{Id: el.From.ID.String(), Title: el.From.Title, ContentSummary: stringOrEmpty(el.From.ContentSummary)},
		To:           &knowledgev1.ContentPreview{Id: el.To.ID.String(), Title: el.To.Title, ContentSummary: stringOrEmpty(el.To.ContentSummary)},
		RelationType: knowledgev1.RelationType(knowledgev1.RelationType_value[string(el.RelationType)]),
		Weight:       derefOrZero(el.Weight),
		CreatedAt:    timestamppb.New(el.CreatedAt),
		UpdatedAt:    timestamppb.New(el.UpdatedAt),
	}, nil
}

func (s *GRPCServer) ListKnowledgeLinks(ctx context.Context, req *knowledgev1.ListKnowledgeLinksRequest) (*knowledgev1.ListKnowledgeLinksResponse, error) {
	contentID, err := uuid.Parse(req.ContentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content id: %v", err)
	}

	pageSize := 0
	if req.Page != nil && req.Page.PageSize > 0 {
		pageSize = int(req.Page.PageSize)
	}

	filter := domain.LinkFilter{
		ContentID:    contentID,
		Direction:    domain.LinkDirection(req.Direction.String()),
		RelationType: domain.RelationType(req.RelationType.String()),
		Limit:        pageSize,
		Offset:       0, // TODO: Implement pagination with page token
	}

	els, err := s.knowledgeLinkService.ListKnowledgeLinks(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list knowledge links: %v", err)
	}

	out := make([]*knowledgev1.EnrichedKnowledgeLink, len(els))
	for i, el := range els {
		out[i] = &knowledgev1.EnrichedKnowledgeLink{
			Id:           el.ID.String(),
			From:         &knowledgev1.ContentPreview{Id: el.From.ID.String(), Title: el.From.Title, ContentSummary: stringOrEmpty(el.From.ContentSummary)},
			To:           &knowledgev1.ContentPreview{Id: el.To.ID.String(), Title: el.To.Title, ContentSummary: stringOrEmpty(el.To.ContentSummary)},
			RelationType: knowledgev1.RelationType(knowledgev1.RelationType_value[string(el.RelationType)]),
			Weight:       derefOrZero(el.Weight),
			CreatedAt:    timestamppb.New(el.CreatedAt),
			UpdatedAt:    timestamppb.New(el.UpdatedAt),
		}
	}

	return &knowledgev1.ListKnowledgeLinksResponse{Items: out}, nil
}

func (s *GRPCServer) ListAllSpaceLinks(ctx context.Context, req *knowledgev1.ListAllSpaceLinksRequest) (*knowledgev1.ListAllSpaceLinksResponse, error) {
	spaceID, err := uuid.Parse(req.SpaceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid space id: %v", err)
	}

	pageSize := 0
	if req.Page != nil && req.Page.PageSize > 0 {
		pageSize = int(req.Page.PageSize)
	}

	links, err := s.knowledgeLinkService.ListKnowledgeLinksBySpace(ctx, spaceID, domain.RelationType(req.RelationType.String()), pageSize, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list space links: %v", err)
	}

	protoLinks := make([]*knowledgev1.KnowledgeLink, len(links))
	for i, link := range links {
		protoLinks[i] = s.domainKnowledgeLinkToProto(link)
	}

	return &knowledgev1.ListAllSpaceLinksResponse{Items: protoLinks}, nil
}

func (s *GRPCServer) UpdateKnowledgeLink(ctx context.Context, req *knowledgev1.UpdateKnowledgeLinkRequest) (*knowledgev1.KnowledgeLink, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid link id: %v", err)
	}

	var relationType domain.RelationType
	var weight float64

	// Check which fields to update based on field mask
	if req.UpdateMask != nil {
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "relation_type":
				relationType = domain.RelationType(req.Link.RelationType.String())
			case "weight":
				weight = req.Link.Weight
			}
		}
	} else {
		// If no field mask, update all fields
		relationType = domain.RelationType(req.Link.RelationType.String())
		weight = req.Link.Weight
	}

	link, err := s.knowledgeLinkService.UpdateKnowledgeLink(ctx, id, relationType, weight)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update knowledge link: %v", err)
	}

	return s.domainKnowledgeLinkToProto(link), nil
}

func (s *GRPCServer) DeleteKnowledgeLink(ctx context.Context, req *knowledgev1.DeleteKnowledgeLinkRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid link id: %v", err)
	}

	err = s.knowledgeLinkService.DeleteKnowledgeLink(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete knowledge link: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *GRPCServer) GetBacklinks(ctx context.Context, req *knowledgev1.GetBacklinksRequest) (*knowledgev1.GetBacklinksResponse, error) {
	contentID, err := uuid.Parse(req.ContentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid content id: %v", err)
	}

	links, err := s.knowledgeLinkService.GetBacklinks(ctx, contentID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get backlinks: %v", err)
	}

	protoLinks := make([]*knowledgev1.KnowledgeLink, len(links))
	for i, link := range links {
		protoLinks[i] = s.domainKnowledgeLinkToProto(link)
	}

	return &knowledgev1.GetBacklinksResponse{
		Items: protoLinks,
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
	// Check Neo4j health
	if err := s.checkNeo4j(ctx); err != nil {
		components["neo4j"] = "FAIL: " + err.Error()
		overallStatus = "FAIL"
	} else {
		components["neo4j"] = "OK"
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

// checkNeo4j performs a simple health check for Neo4j.
func (s *GRPCServer) checkNeo4j(ctx context.Context) error {
	// Try a simple operation, e.g., get backlinks for a random UUID
	// This is a dummy check; in production, use a proper ping or status API
	dummyID := uuid.New()
	_, err := s.knowledgeLinkService.GetBacklinks(ctx, dummyID)
	// If the error is not a connection error, ignore "not found" errors
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
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
		CreatedAt:   timestamppb.New(space.CreatedAt),
		UpdatedAt:   timestamppb.New(space.LastUpdatedAt),
	}
}

func (s *GRPCServer) domainSpaceWithStatsToProto(space *domain.SpaceWithStats) *knowledgev1.Space {
	protoSpace := s.domainSpaceToProto(&space.Space)
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
		CreatedAt:         timestamppb.New(content.CreatedAt),
		UpdatedAt:         timestamppb.New(content.UpdatedAt),
	}
}

func (s *GRPCServer) domainKnowledgeLinkToProto(link *domain.KnowledgeLink) *knowledgev1.KnowledgeLink {
	return &knowledgev1.KnowledgeLink{
		Id:            link.ID.String(),
		FromContentId: link.FromContentID.String(),
		ToContentId:   link.ToContentID.String(),
		RelationType:  knowledgev1.RelationType(knowledgev1.RelationType_value[string(link.RelationType)]),
		Weight:        derefOrZero(link.Weight),
		CreatedAt:     timestamppb.New(link.CreatedAt),
		UpdatedAt:     timestamppb.New(link.UpdatedAt),
	}
}

func stringOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefOrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
