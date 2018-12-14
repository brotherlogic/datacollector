package main

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datacollector/proto"
)

func TestGetDataSets(t *testing.T) {
	s := InitTestServer()

	resp, err := s.GetDataSets(context.Background(), &pb.GetDataSetsRequest{})

	if err != nil {
		t.Errorf("Error in getting data sets: %v, %v", resp, err)
	}
}
