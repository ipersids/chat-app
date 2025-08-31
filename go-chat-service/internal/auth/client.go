package auth

import (
	"context"
	"fmt"
	pb "go-chat-service/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthClient(authServiceAddr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (auth *AuthClient) Close() error {
	return auth.conn.Close()
}

func (auth AuthClient) LoginUser(info context.Context, login, password string) (*pb.User, error) {
	request := &pb.LoginRequest{
		Login:    login,
		Password: password,
	}

	response, err := auth.client.Login(info, request)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("authentication failed: %s", response.Error)
	}

	return response.GetUser(), nil
}

func (auth AuthClient) CreateUser(info context.Context, login, password string) (*pb.User, error) {
	request := &pb.CreateRequest{
		Login:    login,
		Password: password,
	}

	response, err := auth.client.Create(info, request)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("create request failed: %s", response.Error)
	}

	return response.User, nil
}
