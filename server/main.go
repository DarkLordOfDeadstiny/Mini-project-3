package main

import (
	"context"
	"flag"
	"strconv"
	"time"

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
	// bidders []gRPC.Amount
	biddermap map[int64]string
	auctionRunning bool
}
var server *Server

var serverName = flag.String("name", "default", "Senders name")
var port = flag.String("port", "5400", "Server port")
var minutes = flag.String("minutes", "1", "amount of minutes the auction is running")

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
	go run()
	for {
		time.Sleep(time.Second*5)
	}
}

func run()  {
	min, _ := strconv.ParseInt(*minutes, 10, 64)
	time.Sleep(time.Minute*time.Duration(min))
	server.auctionRunning = false
	log.Printf("The auction is closed and the winner of the auction with a bid of %d by %s." , server.highestBid, server.biddermap[server.highestBid])
}

func newServer(serverPort string) *Server {
	s := &Server{
		name: *serverName,
		port: serverPort,
		// bidders: []gRPC.Amount{},
		highestBid: 0,
		biddermap: make(map[int64]string),
		auctionRunning: true,
	}
	fmt.Println(s) //prints the server struct to console
	return s
}

func launchServer(ports []string) {
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
	server = newServer(ports[0])
	gRPC.RegisterAuctionServiceServer(grpcServer, server)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
	//code here is unreachable
}

func (s *Server) Bid(ctx context.Context, amount *gRPC.Amount) (*gRPC.Ack, error) { // lacks a way to end the auction
	if (s.auctionRunning) {
		
		if amount.Amount > s.highestBid { //eller højere end Result(){
			s.highestBid = amount.Amount

			// s.bidders = append(s.bidders, *amount) //maybe check if they exist first and then update the old amount if it does exist in the slice ------------------------------
			s.biddermap[amount.Amount] = amount.BiddersName
		} else {
			return &gRPC.Ack{Status: "fail"}, nil
		}
		
		log.Printf("Server %s: %s Bid for %d, with %s", *serverName, amount.BiddersName, amount.Amount, "success")
		return &gRPC.Ack{Status: "success"}, nil
	} else {
		return &gRPC.Ack{Status: "The auction has closed, please leave"}, nil
	}
}

func (s *Server) Result(ctx context.Context, void *gRPC.Void) (*gRPC.Outcome, error){
	log.Printf("Client askes for highest bid of %d", s.highestBid)
	return &gRPC.Outcome{Status: "Highest current bid:", HighestBid: s.highestBid}, nil
}
