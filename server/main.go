package main

import (
	"context"
	"flag"

	//"errors"
	"fmt"
	"log"
	"net"
	"os"
	//"peter"

	gRPC "github.com/DarkLordOfDeadstiny/Mini-project-3/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedAuctionServiceServer
	name string
	highestBid int64
}

var serverName = flag.String("name", "default", "Senders name")
var port = flag.String("port", "5400", "Server port")

func main() {
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// Create listener tcp on given port or port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", *port, err)
	}

	//Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	//connect to log file
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	gRPC.RegisterAuctionServiceServer(grpcServer, newServer())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}

func newServer() *Server {
	s := &Server{name: *serverName}
	fmt.Println(s) //prints the server struct to console
	return s
}

func (s *Server) Bid(ctx context.Context, amount *gRPC.Amount) (*gRPC.Ack, error) {
	var status string
	
	if amount.Amount > s.highestBid { //eller højere end Result(){
		s.highestBid = amount.Amount
		
		status = "success"
	} else {
		status = "fail" 
	}
	
	return &gRPC.Ack{Status: status}, nil //return status as fail, success or exception
}

func (s *Server) Result(ctx context.Context, void *gRPC.Void) (*gRPC.Outcome, error){
	return &gRPC.Outcome{Status: "Highest current bid:", HighestBid: s.highestBid }, nil
}

func (s *Server) MultiCast(ctx context.Context, amount *gRPC.Amount) (*gRPC.Amount, error){
	s.highestBid = amount.Amount
	return amount, nil
}
