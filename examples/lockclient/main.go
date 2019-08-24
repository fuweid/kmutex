package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	pb "github.com/fuweid/kmutex/examples/proto"
	"google.golang.org/grpc"
)

const address = "localhost:51234"

var (
	action = flag.String("action", "lock", "take action to lock or unlock")
	key    = flag.String("key", "", "key value")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	cli := pb.NewLockerClient(conn)

	switch strings.ToLower(*action) {
	case "lock":
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		if _, err := cli.Lock(ctx, &pb.LockRequest{Key: *key}); err != nil {
			log.Fatalf("could not lock %s: %v", *key, err)
		}
		log.Printf("grant lock: %s", *key)

	case "unlock":
		if _, err := cli.Unlock(context.TODO(), &pb.UnlockRequest{Key: *key}); err != nil {
			log.Fatalf("could not unlock %s: %v", *key, err)
		}
		log.Printf("revoke lock: %s", *key)

	default:
		log.Fatalf("undefine action: %s", *action)
	}
}
