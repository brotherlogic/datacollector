package main

import (
	"io/ioutil"
	"testing"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datacollector/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	"github.com/golang/protobuf/proto"
)

func TestSaveAndLoadData(t *testing.T) {
	s := InitTestServer()
	tstamp := time.Now().Add(time.Hour * -1).Unix()
	s.config.Data = append(s.config.Data, &pb.DataSet{JobName: "madeup", Identifier: "madeup", Staging: []*pb.Reading{&pb.Reading{Timestamp: tstamp, Measure: &pbg.State{Key: "blah", Value: int64(20)}, Collapsed: true}}, Readings: []*pb.Reading{&pb.Reading{Timestamp: tstamp, Measure: &pbg.State{Key: "blah", Value: int64(12)}}}})
	data, file := s.saveData(context.Background())

	if len(file) == 0 {
		t.Errorf("No filename specified")
	}

	if len(data.Data) == 0 || len(data.Data[0].Readings) > 0 {
		t.Errorf("Readings have not been stripped")
	}

	by, _ := proto.Marshal(data)
	err := ioutil.WriteFile(file, by, 0644)

	if err != nil {
		t.Fatalf("Unable to write data: %v", err)
	}

	s2 := InitTestServer()
	err = s2.loadData("")
	if err != nil {
		t.Fatalf("Unable to read data: %v", err)
	}

	if len(s2.config.Data) != 1 {
		t.Errorf("Reading data has failed")
	}

}
