package main

import (
	"context"
	"flag"
	"time"

	//"errors"
	"fmt"
	"log"
	"net"
	"os"

	gRPC "github.com/DarkLordOfDeadstiny/Mini-project-3/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedAuctionServiceServer
	name string
	port string
	highestBid int64
	bidders []gRPC.Amount
}

var serverName = flag.String("name", "default", "Senders name")
var port = flag.String("port", "5400", "Server port")

var _ports [5]string = [5]string{*port, "5401", "5402", "5403", "5404"}


func main() {
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

	flag.Parse()
	fmt.Println(".:server is starting:.")
	go launchServer(_ports[:])
	//code here is unreachable
	for {
		time.Sleep(time.Second*5)
	}
}

func newServer(serverPort string) *Server {
	s := &Server{
		name: *serverName,
		port: serverPort,
		bidders: []gRPC.Amount{},
	}
	fmt.Println(s) //prints the server struct to console
	return s
}

func launchServer(ports []string)  {
	// Create listener tcp on given port or port 5400
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, ports[0])
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", ports[0]))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err)
		if len(ports) > 1 {
			launchServer(ports[1:])
		} else {
			log.Fatalf("Server %s: Failed to find open port", *serverName)
		}
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	gRPC.RegisterAuctionServiceServer(grpcServer, newServer(ports[0]))

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
	//code here is unreachable
}

func (s *Server) Bid(ctx context.Context, amount *gRPC.Amount) (*gRPC.Ack, error) { // lacks a way to end the auction
	var status string
	
	if amount.Amount > s.highestBid { //eller højere end Result(){
		s.highestBid = amount.Amount

		status = "success"
		s.bidders = append(s.bidders, *amount) //maybe check if they exist first and then update the old amount if it does exist in the slice ------------------------------
	} else {
		status = "fail" 
	}
	
	log.Printf("%s Bid for %d, with %s", amount.BiddersName, amount.Amount, status)
	return &gRPC.Ack{Status: status}, nil //return status as fail, success or exception
}

func (s *Server) Result(ctx context.Context, void *gRPC.Void) (*gRPC.Outcome, error){
	log.Printf("Client askes for highest bid of %d", s.highestBid)
	return &gRPC.Outcome{Status: "Highest current bid:", HighestBid: s.highestBid }, nil
}

func (s *Server) MultiCast(ctx context.Context, amount *gRPC.Amount) (*gRPC.Amount, error){
	s.highestBid = amount.Amount
	return amount, nil
}
