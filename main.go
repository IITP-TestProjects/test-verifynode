package main

import (
	"log"
	"net"
	pv "test-verifier/proto_verify"

	"google.golang.org/grpc"
)

func main() {
	srv := newVerifyService()
	s := grpc.NewServer()
	pv.RegisterCommitteeServiceServer(s, srv)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		panic(err)
	}
	log.Println("Verify Service is running on port 50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
