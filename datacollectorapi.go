package main

import (
	pb "github.com/brotherlogic/datacollector/proto"
	"golang.org/x/net/context"
)

// GetDataSets gets the data sets
func (s *Server) GetDataSets(ctx context.Context, req *pb.GetDataSetsRequest) (*pb.GetDataSetsResponse, error) {
	return &pb.GetDataSetsResponse{DataSets: s.config.Data}, nil
}
