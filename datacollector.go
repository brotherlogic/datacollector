package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	flushTime  time.Duration
	saveTime   time.Duration
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&pb.Config{},
		prodRetriever{},
		&pb.ReadConfig{},
		0,
		0,
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
	//Confirm that the config is correct
	for _, rs := range s.readConfig.Spec {
		if rs.GetName() == "" {
			return fmt.Errorf("Unable to mote: %v is missing a name", rs)
		}
	}
	return nil
}

func (s *Server) collect(ctx context.Context) {
	for _, c := range s.readConfig.Spec {
		s.retrieve(ctx, "", c.JobName, c.MeasureKey, c.Name)
	}
}

func (s *Server) deliver(w http.ResponseWriter, r *http.Request) {
	elems := strings.Split(r.URL.Path[1:], "/")
	if len(elems) != 2 {
		w.Write([]byte(fmt.Sprintf("Unable to handle request: %v", r.URL.Path)))
	} else {
		data := s.getJSON(elems[0], elems[1])
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(data)
	}
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
	staging := int64(0)
	for _, data := range s.config.Data {
		staging += int64(len(data.Staging))
	}
	return []*pbg.State{
		&pbg.State{Key: "collected", Value: int64(len(s.config.Data))},
		&pbg.State{Key: "flush_time", TimeDuration: s.flushTime.Nanoseconds()},
		&pbg.State{Key: "save_time", TimeDuration: s.saveTime.Nanoseconds()},
	}
}

func (s *Server) runSave(ctx context.Context) {
	data, filename := s.saveData(ctx)
	by, _ := proto.Marshal(data)

	err := ioutil.WriteFile("/media/scratch/datacollector/"+filename, by, 0644)

	if err != nil {
		s.Log(fmt.Sprintf("Error writing data: %v", err))
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
	server.RegisterRepeatingTask(server.flushToStaging, "flush_to_staging", time.Minute*30)
	server.RegisterRepeatingTask(server.collapseStaging, "collapse_staging", time.Minute*30)
	server.RegisterRepeatingTask(server.runSave, "save_data", time.Minute*5)
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
