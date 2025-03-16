package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userspb "2k4sm/grpc-crud/proto/users"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}


	grpcServer := grpc.NewServer()

	log.Println("Serving gRPC on localhost:8080")
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	conn, err := grpc.NewClient(
		"localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	mux := runtime.NewServeMux()
	err = userspb.RegisterUsersHandler(context.Background(), mux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	httpServer := &http.Server{
		Addr:    ":6969",
		Handler: mux,
	}

	log.Println("Serving gRPC-Gateway on http://localhost:6969")
	log.Fatalln(httpServer.ListenAndServe())
}
