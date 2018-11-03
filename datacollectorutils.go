package main

import (
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datacollector/proto"
)

func (s *Server) retrieve(ctx context.Context, server, job, variable string) {
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
					s.config.Data = append(s.config.Data, &pb.DataSet{JobName: job, Identifier: server, Readings: []*pb.Reading{&pb.Reading{Timestamp: time.Now().Unix(), Measure: stat}}})
				}
			}
		}
	}
}
