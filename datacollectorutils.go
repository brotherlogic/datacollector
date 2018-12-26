package main

import (
	"encoding/json"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datacollector/proto"
	pbgs "github.com/brotherlogic/goserver/proto"
)

func (s *Server) flushToStaging(ctx context.Context) {
	stTime := time.Now()
	for _, dataset := range s.config.Data {
		i := 0
		for i < len(dataset.Readings) {
			if dataset.Readings[i].Timestamp < time.Now().Add(time.Minute*-15).Unix() {
				dataset.Staging = append(dataset.Staging, &pb.Reading{Timestamp: dataset.Readings[i].Timestamp, Measure: dataset.Readings[i].Measure})
				dataset.Readings = append(dataset.Readings[:i], dataset.Readings[i+1:]...)
			} else {
				i++
			}
		}
	}
	s.flushTime = time.Now().Sub(stTime)
}

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
			for _, r := range dataset.Staging {
				if r.Measure.Key == variable {
					resp = append(resp, jsonResponse{Timestamp: r.Timestamp, Value: r.Measure.Value})
				}
			}

		}
	}

	data, _ := json.Marshal(resp)
	return data
}

func matchMeasure(a, b *pbgs.State) bool {
	return (a.Key == b.Key &&
		a.TimeValue == b.TimeValue &&
		a.Value == b.Value &&
		a.Text == b.Text &&
		a.Fraction == b.Fraction &&
		a.TimeDuration == b.TimeDuration)
}

func (s *Server) collapseStaging(ctx context.Context) {
	for _, dataset := range s.config.Data {
		i := 0
		for i < len(dataset.Staging)-1 {
			if matchMeasure(dataset.Staging[i].Measure, dataset.Staging[i+1].Measure) {
				dataset.Staging = append(dataset.Staging[:i], dataset.Staging[i+1:]...)
			} else {
				dataset.Staging[i].Collapsed = true
				i++
			}
		}
	}
}
