package main

import "golang.org/x/net/context"
import pb "github.com/brotherlogic/datacollector/proto"

// GetDataSets gets the data sets
func (s *Server) GetDataSets(ctx context.Context, req *pb.GetDataSetsRequest) (*pb.GetDataSetsResponse, error) {
	return &pb.GetDataSetsResponse{}, nil
}
