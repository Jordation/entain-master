package service

import (
	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports_pb"
	"golang.org/x/net/context"
)

type Sports interface {
	// ListSports will return a collection of sports_pb.
	ListSports(ctx context.Context, in *sports_pb.ListSportsRequest) (*sports_pb.ListSportsResponse, error)
	GetSport(ctx context.Context, in *sports_pb.GetSportRequest) (*sports_pb.GetSportResponse, error)
}

// sportsService implements the Sports interface.
type sportsService struct {
	sportsRepo db.SportsRepo
}

// NewSportsService instantiates and returns a new sportsService.
func NewSportsService(sportsRepo db.SportsRepo) Sports {
	return &sportsService{sportsRepo}
}

// ListSports returns a list of sports matching the provided filter params
func (s *sportsService) ListSports(ctx context.Context, in *sports_pb.ListSportsRequest) (*sports_pb.ListSportsResponse, error) {
	sports, err := s.sportsRepo.List(in.Filter)
	if err != nil {
		return nil, err
	}
	return &sports_pb.ListSportsResponse{Sports: sports}, nil
}

// GetSport returns a single sport provided an ID
func (s *sportsService) GetSport(ctx context.Context, in *sports_pb.GetSportRequest) (*sports_pb.GetSportResponse, error) {
	sport, err := s.sportsRepo.Get(in.GetSportId())
	if err != nil {
		return nil, err
	}
	return &sports_pb.GetSportResponse{Sport: sport}, nil
}
