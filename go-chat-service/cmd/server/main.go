package main

import (
	auth "go-chat-service/internal/auth"
	chat "go-chat-service/internal/chat"
	pb "go-chat-service/pkg/pb"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

const (
	chatPort        = ":50052"
	userServiceAddr = "localhost:50051"
)

func main() {
	log.Println("Starting Go Chat Service...")
	log.Printf("Connecting to Rust user service at %s...", userServiceAddr)
	authClient, err := auth.NewAuthClient(userServiceAddr)
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer func() {
		if err := authClient.Close(); err != nil {
			log.Printf("Error closing auth client: %v", err)
		}
	}()
	log.Println("Connected to Rust user service")

	chatHandler := chat.NewHandler(authClient)
	server := grpc.NewServer()
	pb.RegisterChatServiceServer(server, chatHandler)
	listener, err := net.Listen("tcp", chatPort)
	log.Printf("server listening at %v", listener.Addr())
	if err != nil {
		log.Fatalf("❌ Failed to listen on %s: %v", chatPort, err)
	}
	log.Printf("Chat server listening on %s", chatPort)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		log.Printf("Received signal %v, shutting down gracefully...", sig)
		server.GracefulStop()
		log.Println("Server stopped")
	}()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}
