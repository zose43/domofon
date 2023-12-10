package auth

import (
	"context"
	domofon_v1 "github.com/zose43/domofon-proto/out/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	EmptyValue = 0
)

type Auth interface {
	Login(ctx context.Context, pass string, email string, appID int) (string, error)
	Register(ctx context.Context, pass string, email string) (int64, error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
}

type handler struct {
	domofon_v1.UnimplementedAuthServer
	auth Auth
}

func Register(grpcSrv *grpc.Server, auth Auth) {
	domofon_v1.RegisterAuthServer(grpcSrv, &handler{auth: auth})
}

func (h handler) Login(
	ctx context.Context,
	request *domofon_v1.LoginRequest,
) (*domofon_v1.LoginResponse, error) {
	if err := validateLogin(request); err != nil {
		return nil, err
	}

	token, err := h.auth.Login(
		ctx,
		request.GetPassword(),
		request.GetEmail(),
		int(request.GetAppId()),
	)
	if err != nil {
		return nil, err
	}

	return &domofon_v1.LoginResponse{Token: token}, nil
}

func validateLogin(request *domofon_v1.LoginRequest) error {
	if request.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}
	if request.GetAppId() == EmptyValue {
		return status.Error(codes.InvalidArgument, "empty app_id")
	}
	if request.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	return nil
}

func (h handler) IsAdmin(
	ctx context.Context,
	request *domofon_v1.IsAdminRequest,
) (*domofon_v1.IsAdminResponse, error) {
	if err := validateIsAdmin(request); err != nil {
		return nil, err
	}

	res, err := h.auth.IsAdmin(ctx, int(request.GetUserId()))
	if err != nil {
		return nil, err
	}

	return &domofon_v1.IsAdminResponse{IsAdmin: res}, nil
}

func validateIsAdmin(request *domofon_v1.IsAdminRequest) error {
	if request.GetUserId() == EmptyValue {
		return status.Error(codes.InvalidArgument, "empty user_id")
	}

	return nil
}

func (h handler) Register(
	ctx context.Context,
	request *domofon_v1.RegisterRequest,
) (*domofon_v1.RegisterResponse, error) {
	if err := validateRegister(request); err != nil {
		return nil, err
	}

	userID, err := h.auth.Register(ctx, request.GetPassword(), request.GetEmail())
	if err != nil {
		return nil, err
	}

	return &domofon_v1.RegisterResponse{Id: userID}, nil
}

func validateRegister(request *domofon_v1.RegisterRequest) error {
	if request.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}
	if request.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	return nil
}
