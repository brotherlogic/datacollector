package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"github.com/golang/protobuf/proto"
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
	servers, err := utils.ResolveAll("")
	if err == nil {
		for _, s := range servers {
			if (server == "" || s.Identifier == server) && (job == "" || s.Name == job) {
				conn, err := grpc.Dial(s.Ip+":"+strconv.Itoa(int(s.Port)), grpc.WithInsecure())
				if err == nil {
					defer conn.Close()
					client := pbg.NewGoserverServiceClient(conn)
					return client.State(ctx, &pbg.Empty{})
				}
			}
		}
	}

	return &pbg.ServerState{}, fmt.Errorf("Unable to locate %v on %v: %v", job, server, err)
}

//Server main server type
type Server struct {
	*goserver.GoServer
	config     *pb.Config
	retriever  retriever
	readConfig *pb.ReadConfig
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&pb.Config{},
		prodRetriever{},
		&pb.ReadConfig{},
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

func (s *Server) collect(ctx context.Context) {
	for _, c := range s.readConfig.Spec {
		s.retrieve(ctx, "", c.JobName, c.MeasureKey)
	}
}

func (s *Server) deliver(w http.ResponseWriter, r *http.Request) {
	data := s.getJSON("recordwants", "budget")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(data)
}

func (s *Server) serveUp() {
	http.HandleFunc("/", s.deliver)
	err := http.ListenAndServe(":8085", nil)
	if err != nil {
		panic(err)
	}
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "collected", Value: int64(len(s.config.Data))},
	}
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
	server.RegisterRepeatingTask(server.collect, "collect", time.Minute*5)
	go server.serveUp()

	data, err := Asset("config/config.pb")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	err = proto.UnmarshalText(string(data), server.readConfig)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	fmt.Printf("%v", server.Serve())
}
