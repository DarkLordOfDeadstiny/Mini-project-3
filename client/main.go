package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gRPC "github.com/DarkLordOfDeadstiny/Mini-project-3/proto"

	"google.golang.org/grpc"
)

var biddersName = flag.String("name", "default", "Senders name")
var tcpServer = flag.String("server", "5400", "Tcp server")

var _ports [5]string = [5]string{*tcpServer, "5401", "5402", "5403", "5404"}

var ctx context.Context
var servers []gRPC.AuctionServiceClient
var ServerConn map[gRPC.AuctionServiceClient]*grpc.ClientConn

func main() {
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//connect to log file
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	fmt.Println("--- join Server ---")
	ServerConn = make(map[gRPC.AuctionServiceClient]*grpc.ClientConn)
	joinServer()
	defer closeAll()

	//start the biding
	parseInput()
}

func closeAll()  {
	for _, c := range ServerConn {
		c.Close()
	}
}

func joinServer() {
	//connect to server
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, port := range _ports {
		log.Printf("client %s: Attempts to dial on port %s\n", *biddersName, port)
		conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", port), opts...)
		if err != nil {
			log.Printf("Fail to Dial : %v", err)
			continue
		}
		var s = gRPC.NewAuctionServiceClient(conn)
		servers = append(servers, s)
		ServerConn[s] = conn
		fmt.Println(conn.GetState().String())
	}
	ctx = context.Background()
}

func bid(in string) {
	for {
		bidval, err := strconv.ParseInt(in, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		amount := &gRPC.Amount{
			BiddersName: *biddersName,
			Amount:      bidval,
		}
		for _, s := range servers{ 
			if conReady(s) {
				fmt.Println(s)
				ack, err := s.Bid(ctx, amount)
				if err != nil {
					log.Printf("Client %s: no response from the server, attempting to reconnect", *biddersName)
					log.Println(err)
				}
				switch ack.Status {
				case "fail":
					fmt.Println("The bid was unsuccessful, must be above the current highest bid")
				case "success":	
					fmt.Println("The bid was successful")
				default:
					fmt.Println(ack.Status)
				}
			}
		}
		parseInput()
	}
}

func getResult() int64 {
	//fmt.Println(context.Background())
	void := &gRPC.Void{}
	for _, s := range servers{
		if conReady(s){
			outcome, _ := s.Result(ctx, void)
			return outcome.HighestBid
		}


	}
	return -1
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type your bidding amount here or type \"status\" to get the current highest bid")
	fmt.Println("--------------------")

	for {
		fmt.Print("-> ")
		in, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		in = strings.TrimSpace(in) 
		if in == "status" {
			fmt.Printf("The current highest bid is %d\n", getResult())
		} else {
			bid(in)
		}
	}
}

func conReady(s gRPC.AuctionServiceClient) bool {
	return ServerConn[s].GetState().String() == "READY"
}