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
	//"time"

	// "time"

	gRPC "github.com/DarkLordOfDeadstiny/Mini-project-3/proto"

	"google.golang.org/grpc"
)

var biddersName = flag.String("name", "default", "Senders name")
var tcpServer = flag.String("server", "5400", "Tcp server")

var _ports [5]string = [5]string{*tcpServer, "5401", "5402", "5403", "5404"}

var ctx context.Context
var server gRPC.AuctionServiceClient
var conn *grpc.ClientConn

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
	joinServer(_ports[:])

	//start the biding
	bid()
	defer conn.Close()
}

func joinServer(ports []string) {
	//connect to server
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())

	log.Printf("client %s: Attempts to dial on port %s\n", *biddersName, ports[0])
	conn, err := grpc.Dial(fmt.Sprintf(":%s", ports[0]), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		if len(ports) > 1 {
			joinServer(ports[1:])
		} else {
			log.Fatalf("Client %s: Failed to find open port", *biddersName)
		}
	}

	ctx = context.Background()
	server = gRPC.NewAuctionServiceClient(conn)
}

// func joinServer(ports []string) {
// 	//connect to server
// 	var opts []grpc.DialOption
// 	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
// 	timeContext, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
// 	log.Printf("client %s: Attempts to dial on port %s\n", *biddersName, ports[0])
// 	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", ports[0]), opts...)
// 	if err != nil {
// 		log.Printf("Fail to Dial : %v", err)
// 		if len(ports) > 1 {
// 			joinServer(ports[1:])
// 		} else {
// 			log.Fatalf("Client %s: Failed to find open port", *biddersName)
// 		}
// 	}

// 	// ctx = context.Background()
// 	server = gRPC.NewAuctionServiceClient(conn)
// }

func bid() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type your bidding amount here")
	fmt.Println("--------------------")

	for {
		fmt.Printf("The current highest bid is %d\n", getResult())
		fmt.Print("-> ")
		in, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		in = strings.TrimSpace(in)
		bidval, err := strconv.ParseInt(in, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		amount := &gRPC.Amount{
			BiddersName: *biddersName,
			Amount:      bidval,
		}
		ack, err := server.Bid(ctx, amount)
		if err != nil {
			log.Printf("Client %s: no response from server, tries to reconnect", *biddersName)
			// conn.Close()
			joinServer(_ports[:])

			log.Fatal(err)
		}
		if ack.Status == "fail" {
			fmt.Println("The bid was unsuccessful, must be above the current highest bid")
		} else {
			fmt.Println("The bid was successsful")
		}

	}
}

func getResult() int64 {
	outcome, error := server.Result(ctx, &gRPC.Void{})
	if error != nil {
		return getResult()
	}
	return outcome.HighestBid
}
