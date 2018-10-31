package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/datacollector/proto"
	pbg "github.com/brotherlogic/goserver/proto"
)

type retriever interface {
	retrieve(ctx context.Context, server, job string) (*pbg.ServerState, error)
}

type prodRetriever struct{}

func (p prodRetriever) retrieve(ctx context.Context, server, job string) (*pbg.ServerState, error) {
	servers, err := utils.ResolveAll(job)
	if err == nil {
		for _, s := range servers {
			if s.Identifier == server {
				conn, err := grpc.Dial(s.Ip+":"+strconv.Itoa(int(s.Port)), grpc.WithInsecure())
				if err == nil {
					defer conn.Close()
					client := pbg.NewGoserverServiceClient(conn)
					return client.State(ctx, &pbg.Empty{})
				}
			}
		}
	}

	return &pbg.ServerState{}, fmt.Errorf("Unable to locate %v on %v", job, server)
}

//Server main server type
type Server struct {
	*goserver.GoServer
	config    *pb.Config
	retriever retriever
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&pb.Config{},
		prodRetriever{},
	}
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	// Do nothing
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{}
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.PrepServer()
	server.Register = server
	server.RegisterServer("datacollector", false)
	fmt.Printf("%v", server.Serve())
}
