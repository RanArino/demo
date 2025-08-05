package server

import (
	"context"
	"fmt"
	userv1 "demo/ms_user/api/proto/v1"
	"demo/ms_user/internal/domain"
	"demo/ms_user/internal/middleware"
	"demo/ms_user/internal/service"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// grpcServer implements the userv1.UserServiceServer interface.
type grpcServer struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService
}

// NewGRPCServer creates a new gRPC server.
func NewGRPCServer(userService *service.UserService) userv1.UserServiceServer {
	return &grpcServer{userService: userService}
}

// CreateUser handles the gRPC request to create a user.
func (s *grpcServer) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user, err := s.userService.CreateUser(ctx, req.ClerkUserId, req.Email)
	if err != nil {
		return nil, err
	}
	return &userv1.CreateUserResponse{User: toUserPb(user)}, nil
}

// GetUser handles the gRPC request to get a user.
func (s *grpcServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	return &userv1.GetUserResponse{User: toUserPb(user)}, nil
}

// UpdateUser handles the gRPC request to update a user.
func (s *grpcServer) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	user, err := s.userService.UpdateUser(
		ctx,
		req.Email,
		req.FullName,
		req.Username,
	)
	if err != nil {
		return nil, err
	}
	return &userv1.UpdateUserResponse{User: toUserPb(user)}, nil
}

// DeleteUser handles the gRPC request to delete a user.
func (s *grpcServer) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx)
	if err != nil {
		return nil, err
	}
	return &userv1.DeleteUserResponse{}, nil
}

// UpdateUserPreferences handles the gRPC request to update user preferences.
func (s *grpcServer) UpdateUserPreferences(ctx context.Context, req *userv1.UpdateUserPreferencesRequest) (*userv1.UpdateUserPreferencesResponse, error) {
	var canvasSettings, notificationSettings, accessibilitySettings map[string]interface{}
	if req.CanvasSettings != nil {
		canvasSettings = req.CanvasSettings.AsMap()
	}
	if req.NotificationSettings != nil {
		notificationSettings = req.NotificationSettings.AsMap()
	}
	if req.AccessibilitySettings != nil {
		accessibilitySettings = req.AccessibilitySettings.AsMap()
	}

	prefs, err := s.userService.UpdateUserPreferences(
		ctx,
		req.Theme,
		req.Language,
		req.Timezone,
		canvasSettings,
		notificationSettings,
		accessibilitySettings,
	)
	if err != nil {
		return nil, err
	}
	return &userv1.UpdateUserPreferencesResponse{Preferences: toUserPreferencesPb(prefs)}, nil
}

func (s *grpcServer) ActivateUser(ctx context.Context, req *userv1.ActivateUserRequest) (*userv1.ActivateUserResponse, error) {
	user, err := s.userService.ActivateUser(ctx, req.FullName, req.Username)
	if err != nil {
		return nil, err
	}
	return &userv1.ActivateUserResponse{User: toUserPb(user)}, nil
}

// CheckUserStatus handles the gRPC request to check user status and profile completion.
func (s *grpcServer) CheckUserStatus(ctx context.Context, req *userv1.CheckUserStatusRequest) (*userv1.CheckUserStatusResponse, error) {
	status, err := s.userService.CheckUserStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &userv1.CheckUserStatusResponse{
		ProfileCompleted: status.ProfileCompleted,
		NeedsRedirect:    status.NeedsRedirect,
		RedirectUrl:      status.RedirectURL,
		User:             toUserPb(status.User),
	}, nil
}

// --- Conversion Helpers ---

func toUserPb(user *domain.User) *userv1.User {
	if user == nil {
		return nil
	}
	var deletedAt *timestamppb.Timestamp
	if user.DeletedAt != nil {
		deletedAt = timestamppb.New(*user.DeletedAt)
	}

	return &userv1.User{
		Id:                user.ID.String(),
		ClerkUserId:       user.ClerkUserID,
		Email:             user.Email,
		FullName:          user.FullName,
		Username:          user.Username,
		StorageUsedBytes:  user.StorageUsedBytes,
		StorageQuotaBytes: user.StorageQuotaBytes,
		Status:            user.Status,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
		DeletedAt:         deletedAt,
	}
}

func toUserPreferencesPb(prefs *domain.UserPreferences) *userv1.UserPreferences {
	if prefs == nil {
		return nil
	}

	canvas, _ := structpb.NewStruct(prefs.CanvasSettings)
	notification, _ := structpb.NewStruct(prefs.NotificationSettings)
	accessibility, _ := structpb.NewStruct(prefs.AccessibilitySettings)

	return &userv1.UserPreferences{
		UserId:                prefs.UserID.String(),
		Theme:                 prefs.Theme,
		Language:              prefs.Language,
		Timezone:              prefs.Timezone,
		CanvasSettings:        canvas,
		NotificationSettings:  notification,
		AccessibilitySettings: accessibility,
		CreatedAt:             timestamppb.New(prefs.CreatedAt),
		UpdatedAt:             timestamppb.New(prefs.UpdatedAt),
	}
}
