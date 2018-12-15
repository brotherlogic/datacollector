package main

import (
	"testing"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datacollector/proto"
)

func TestGetDataSets(t *testing.T) {
	s := InitTestServer()
	s.readConfig.Spec = append(s.readConfig.Spec, &pb.ReadSpec{MeasureKey: "blah", Name: "Made Up Name"})
	s.collect(context.Background())

	resp, err := s.GetDataSets(context.Background(), &pb.GetDataSetsRequest{})
	if err != nil {
		t.Errorf("Error in getting data sets: %v, %v", resp, err)
	}

	if len(resp.DataSets) != 1 {
		t.Fatalf("Wrong number of data sets returned: %v (1)", len(resp.DataSets))
	}

	if resp.DataSets[0].SpecName != "Made Up Name" {
		t.Errorf("Name has not been transferred: %v", resp.DataSets[0])
	}
}
