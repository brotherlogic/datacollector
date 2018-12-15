package main

import (
	"encoding/json"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datacollector/proto"
)

func (s *Server) retrieve(ctx context.Context, server, job, variable string, name string) {
	stats, err := s.retriever.retrieve(ctx, server, job)
	if err == nil {
		for _, stat := range stats.States {
			if stat.Key == variable {
				seen := false
				for _, set := range s.config.Data {
					if set.JobName == job && set.Identifier == server {
						set.Readings = append(set.Readings, &pb.Reading{Timestamp: time.Now().Unix(), Measure: stat})
						seen = true
					}
				}

				if !seen {
					s.config.Data = append(s.config.Data, &pb.DataSet{JobName: job, Identifier: server, SpecName: name, Readings: []*pb.Reading{&pb.Reading{Timestamp: time.Now().Unix(), Measure: stat}}})
				}
			}
		}
	}
}

type jsonResponse struct {
	Timestamp int64 `json:"timestamp"`
	Value     int64 `json:"value"`
}

func (s *Server) getJSON(job, variable string) []byte {
	resp := []jsonResponse{}
	for _, dataset := range s.config.Data {
		if dataset.JobName == job {
			for _, r := range dataset.Readings {
				if r.Measure.Key == variable {
					resp = append(resp, jsonResponse{Timestamp: r.Timestamp, Value: r.Measure.Value})
				}
			}
		}
	}

	data, _ := json.Marshal(resp)
	return data
}
