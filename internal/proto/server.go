// Package proto describes operation of GRPC server.
package proto

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go-shortener-url/internal/pkg/sign"
	"go-shortener-url/internal/usecase"
)

// GRPCServer contains necessary tools for running GRPC server.
type GRPCServer struct {
	UnimplementedShortenerServer
	server  *grpc.Server
	service *usecase.Manager
}

// Run starts grpc server for execution.
func (s *GRPCServer) Run() error {
	RegisterShortenerServer(s.server, s)

	listener, err := net.Listen("tcp", "localhost:3200")
	if err != nil {
		return err
	}

	return s.server.Serve(listener)
}

// Shutdown grpc server stops working.
func (s *GRPCServer) Shutdown(_ context.Context) error {
	s.server.GracefulStop()
	return nil
}

// NewGRPCServer server constructor.
func NewGRPCServer(m *usecase.Manager) *GRPCServer {
	srv := grpc.NewServer()

	return &GRPCServer{
		server:  srv,
		service: m,
	}
}

// ShortenURL shortens original URL.
func (s *GRPCServer) ShortenURL(ctx context.Context, r *ShortenURLRequest) (*ShortenURLResponse, error) {
	if r.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL not specified")
	}

	userID := r.UserId
	if r.UserId == "" || !sign.ValidateID(r.UserId) {
		userID = sign.UserID()
	}

	shortURL, err := s.service.CreateShortURL(ctx, r.Url, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUniqueValue) {
			return &ShortenURLResponse{Url: shortURL, UserId: userID}, err
		}

		return nil, err
	}

	return &ShortenURLResponse{Url: shortURL, UserId: userID}, nil
}

// GetFullURL allows you to get original URL using a shortened one.
func (s *GRPCServer) GetFullURL(ctx context.Context, r *GetFullURLRequest) (*GetFullURLResponse, error) {
	if r.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL not specified")
	}

	fullURL, err := s.service.GetFullURL(ctx, r.Url)
	if err != nil {
		if errors.Is(err, usecase.ErrDeletedURL) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &GetFullURLResponse{Url: fullURL}, nil
}

// ShortenBatchURLs allows you to shorten a bunch of original URLs.
func (s *GRPCServer) ShortenBatchURLs(ctx context.Context, r *ShortenBatchURLsRequest) (*ShortenBatchURLsResponse, error) {
	if len(r.Urls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty array URLs")
	}

	resp := make([]*ShortenBatchItemResponse, len(r.Urls))

	userID := r.UserId
	if r.UserId == "" || !sign.ValidateID(r.UserId) {
		userID = sign.UserID()
	}

	for _, v := range r.Urls {
		if v.Url == "" {
			return nil, status.Error(codes.InvalidArgument, "URL not specified")
		}

		shortURL, err := s.service.CreateShortURL(ctx, v.Url, r.UserId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		resp = append(resp, &ShortenBatchItemResponse{Id: v.Id, Url: shortURL})
	}

	return &ShortenBatchURLsResponse{UserId: userID, Urls: resp}, nil
}

// DeleteURLs marks an array of URLs for deletion.
func (s *GRPCServer) DeleteURLs(_ context.Context, r *DeleteURLsRequest) (*Empty, error) {
	if len(r.Urls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty array URLs")
	}

	if r.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID not specified")
	}

	go s.service.ExecDeleting(r.Urls, r.UserId)

	return &Empty{}, nil
}
