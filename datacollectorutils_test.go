package main

import (
	"strconv"
	"testing"
	"time"

	pb "github.com/brotherlogic/datacollector/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	"golang.org/x/net/context"
)

type testRetriever struct{}

func (p testRetriever) retrieve(ctx context.Context, server, job string) (*pbg.ServerState, error) {
	return &pbg.ServerState{States: []*pbg.State{&pbg.State{Key: "blah", Value: 12}}}, nil
}

func InitTestServer() *Server {
	s := Init()
	s.retriever = &testRetriever{}
	return s
}

func TestRetrieve(t *testing.T) {
	s := InitTestServer()
	s.retrieve(context.Background(), "madeup", "madeup", "blah")

	if len(s.config.Data) == 0 {
		t.Errorf("Did not read data")
	}
}

func TestRetrieveWithAppend(t *testing.T) {
	s := InitTestServer()
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup"})
	s.retrieve(context.Background(), "madeup", "madeup", "blah")

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
