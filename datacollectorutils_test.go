package main

import (
	"log"
	"strconv"
	"testing"
	"time"

	pb "github.com/brotherlogic/datacollector/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	"golang.org/x/net/context"
)

type testRetriever struct{}

func (p testRetriever) retrieve(ctx context.Context, server, job string) (*pbg.ServerState, error) {
	log.Printf("RETREUVE")
	return &pbg.ServerState{States: []*pbg.State{&pbg.State{Key: "blah", Value: 12}}}, nil
}

func InitTestServer() *Server {
	s := Init()
	s.retriever = &testRetriever{}
	return s
}

func TestRetrieve(t *testing.T) {
	s := InitTestServer()
	s.retrieve(context.Background(), "madeup", "madeup", "blah", "thing")

	if len(s.config.Data) == 0 {
		t.Errorf("Did not read data")
	}
}

func TestRetrieveWithAppend(t *testing.T) {
	s := InitTestServer()
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup"})
	s.retrieve(context.Background(), "madeup", "madeup", "blah", "thing")

	if len(s.config.Data) == 0 {
		t.Errorf("Did not read data")
	}

	if len(s.config.Data[0].Readings) == 2 {
		t.Errorf("Did not read data")
	}
}

func TestGetData(t *testing.T) {
	s := InitTestServer()
	tstamp := time.Now().Unix()
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup", Readings: []*pb.Reading{&pb.Reading{Timestamp: tstamp, Measure: &pbg.State{Key: "blah", Value: int64(12)}}}})

	json := s.getJSON("madeup", "blah")
	want := "[{\"timestamp\":" + strconv.Itoa(int(tstamp)) + ",\"value\":12}]"
	if string(json) != want {
		t.Errorf("Json has come back bad: %v", string(json))
		t.Errorf("It should have been   : %v", want)
	}
}

func TestGetDataFollowingFlush(t *testing.T) {
	s := InitTestServer()
	tstamp := time.Now().Add(time.Hour * -1).Unix()
	tstamp2 := time.Now().Unix()
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup", Readings: []*pb.Reading{&pb.Reading{Timestamp: tstamp, Measure: &pbg.State{Key: "blah", Value: int64(12)}}}})
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup", Readings: []*pb.Reading{&pb.Reading{Timestamp: tstamp2, Measure: &pbg.State{Key: "blah", Value: int64(12)}}}})

	s.flushToStaging(context.Background())

	if len(s.config.Data[0].Staging) != 1 {
		t.Fatalf("Data has not been flushed to staging: %v", s.config.Data[0])
	}

	json := s.getJSON("madeup", "blah")
	want := "[{\"timestamp\":" + strconv.Itoa(int(tstamp)) + ",\"value\":12},{\"timestamp\":" + strconv.Itoa(int(tstamp2)) + ",\"value\":12}]"
	if string(json) != want {
		t.Errorf("Json has come back bad: %v", string(json))
		t.Errorf("It should have been   : %v", want)
	}
}

func TestRunCollapse(t *testing.T) {
	s := InitTestServer()

	tstamp := time.Now().Add(time.Hour * -1).Unix()
	tstamp2 := time.Now().Unix()
	tstamp3 := time.Now().Add(time.Hour * -2).Unix()
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup", Staging: []*pb.Reading{&pb.Reading{Timestamp: tstamp3, Measure: &pbg.State{Key: "blah", Value: int64(20)}}, &pb.Reading{Timestamp: tstamp, Measure: &pbg.State{Key: "blah", Value: int64(12)}}, &pb.Reading{Timestamp: tstamp2, Measure: &pbg.State{Key: "blah", Value: int64(12)}}}})

	s.collapseStaging(context.Background())

	if len(s.config.Data[0].Staging) != 2 {
		t.Errorf("Staging has not been collapsed correctly")
	}
}

func TestLoadDataBadRead(t *testing.T) {
	s := InitTestServer()
	err := s.loadData("madeupdirectory")
	if err == nil {
		t.Errorf("Bad directory did not cause error")
	}
}
