package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userspb "2k4sm/grpc-crud/proto/users"
	"2k4sm/grpc-crud/src/db"
	"2k4sm/grpc-crud/src/repositories"
	"2k4sm/grpc-crud/src/services"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file, using default variables.")
	}

	session := db.InitDb()
	defer session.Close()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	userRepo := repositories.NewUserRepository(session)
	userService := services.NewUserService(userRepo)
	userspb.RegisterUsersServer(grpcServer, userService)

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
	mux.HandlePath("GET", "/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := struct {
			Time    string `json:"time"`
			Message string `json:"message"`
		}{
			Time:    time.Now().Format(time.RFC3339),
			Message: "healthy",
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})

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
