package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	cats, err := s.services.Categories.List(ctx)
	if err != nil {
		return nil, handleError(err)
	}

	out := make([]*pb.Category, len(cats))
	for i := range cats {
		out[i] = toProtoCategory(&cats[i])
	}
	return &pb.ListCategoriesResponse{Categories: out}, nil
}

func (s *Server) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.Category, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "category id is required")
	}
	cat, err := s.services.Categories.Get(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return toProtoCategory(cat), nil
}

func (s *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	if req.GetSlug() == "" || req.GetLabel() == "" {
		return nil, status.Error(codes.InvalidArgument, "slug and label are required")
	}
	id, err := s.services.Categories.Create(ctx, req.GetSlug(), req.GetLabel(), req.GetColor())
	if err != nil {
		return nil, handleError(err)
	}
	return &pb.CreateCategoryResponse{Id: id}, nil
}

func (s *Server) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "category id is required")
	}
	if req.GetUpdateMask() == nil {
		return nil, status.Error(codes.InvalidArgument, "update_mask is required")
	}

	fields := fieldsFromUpdateMask(req.GetUpdateMask(), req.GetUpdates())
	if len(fields) == 0 {
		return nil, status.Error(codes.InvalidArgument, "update_mask contains no valid fields")
	}

	if err := s.services.Categories.Update(ctx, req.GetId(), fields); err != nil {
		return nil, handleError(err)
	}

	updatedCat, err := s.services.Categories.Get(ctx, req.GetId())
	if err != nil {
		return nil, handleError(err)
	}
	return toProtoCategory(updatedCat), nil
}

func (s *Server) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "category id is required")
	}
	if err := s.services.Categories.Delete(ctx, req.GetId()); err != nil {
		return nil, handleError(err)
	}
	return &pb.DeleteCategoryResponse{}, nil
}
