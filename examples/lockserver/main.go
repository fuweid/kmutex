package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/fuweid/kmutex"
	pb "github.com/fuweid/kmutex/examples/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const port = ":51234"

// server implements LockService.
type server struct {
	km *kmutex.KMutex
}

// Lock grants lock for key.
func (s *server) Lock(ctx context.Context, in *pb.LockRequest) (_ *empty.Empty, err0 error) {
	if in.GetKey() == "" {
		return nil, fmt.Errorf("empty key is not valid")
	}

	if err := s.km.Lock(ctx, in.GetKey()); err != nil {
		return nil, err
	}
	return new(empty.Empty), nil
}

// Unlock revokes lock for key.
func (s *server) Unlock(ctx context.Context, in *pb.UnlockRequest) (_ *empty.Empty, err0 error) {
	defer func() {
		if err := recover(); err != nil {
			err0 = fmt.Errorf("recover: %v", err)
		}
	}()

	if in.GetKey() == "" {
		return nil, fmt.Errorf("empty key is not valid")
	}

	s.km.Unlock(in.GetKey())
	return new(empty.Empty), nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLockerServer(s, &server{km: kmutex.NewKMutex()})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
