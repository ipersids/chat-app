package chat

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"go-chat-service/internal/auth"
	pb "go-chat-service/pkg/pb"

	"github.com/google/uuid"
)

type Handler struct {
	pb.UnimplementedChatServiceServer
	connManager *ConnectionManager
	userClient  *auth.AuthClient
	messages    []*pb.MessageBroadcast
	msgMutex    sync.RWMutex
}

func NewHandler(authClient *auth.AuthClient) *Handler {
	return &Handler{
		connManager: NewConnectionManager(),
		userClient:  authClient,
		messages:    make([]*pb.MessageBroadcast, 0),
	}
}

func (h *Handler) ManageChat(stream pb.ChatService_ManageChatServer) error {
	var userID string
	var username string
	var authenticated bool

	defer func() {
		if authenticated {
			h.connManager.RemoveConnection(userID)
			// Notify others that user left
			h.connManager.BroadcastSystemNotification(&pb.SystemNotification{
				Type:     pb.SystemNotification_USER_LEFT,
				Username: username,
				Message:  username + " has left the chat",
			}, userID)
		}
	}()

	log.Println("New client connected, waiting for authentication...")

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Client %s closed connection", username)
			return nil
		}
		if err != nil {
			log.Printf("Stream error: %v", err)
			return err
		}

		switch req := request.GetRequestType().(type) {
		case *pb.ChatRequest_Auth:
			if authenticated {
				h.sendErrorResponse(stream, "ALREADY_AUTHENTICATED", "Already authenticated")
				continue
			}

			userID, username, err = h.handleAuthentication(stream, req.Auth)
			if err != nil {
				return err
			}
			authenticated = true

		case *pb.ChatRequest_Message:
			if !authenticated {
				h.sendErrorResponse(stream, "NOT_AUTHENTICATED", "Please authenticate first")
				continue
			}

			h.handleMessage(req.Message, userID, username)

		default:
			h.sendErrorResponse(stream, "UNKNOWN_REQUEST", "Unknown request type")
		}
	}
}

func (h *Handler) handleAuthentication(stream pb.ChatService_ManageChatServer, auth *pb.AuthRequest) (string, string, error) {
	log.Printf("Authenticating user: %s", auth.GetLogin())

	// Validate with Rust user service
	user, err := h.userClient.LoginUser(context.Background(), auth.GetLogin(), auth.GetPassword())
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		h.sendAuthResponse(stream, false, "Invalid credentials", nil)
		return "", "", err
	}

	userID := user.GetUuid()
	username := user.GetLogin()

	// Send success response
	h.sendAuthResponse(stream, true, "", user)

	// Add to connection manager
	h.connManager.AddConnection(userID, username, stream)

	// Notify others that user joined
	h.connManager.BroadcastSystemNotification(&pb.SystemNotification{
		Type:     pb.SystemNotification_USER_JOINED,
		Username: username,
		Message:  username + " has joined the chat",
	}, userID)

	log.Printf("User %s authenticated and connected", username)
	return userID, username, nil
}

func (h *Handler) handleMessage(msgReq *pb.MessageRequest, userID, username string) {
	if msgReq.GetText() == "" {
		return // Ignore empty messages
	}

	// Create message
	message := &pb.MessageBroadcast{
		MessageId: uuid.New().String(),
		UserId:    userID,
		Username:  username,
		Text:      msgReq.GetText(),
		Timestamp: time.Now().Unix(),
	}

	// Store message
	h.storeMessage(message)

	// Broadcast to all other users
	h.connManager.BroadcastMessage(message, userID)

	log.Printf("[CHAT] -> [%s]: %s", username, msgReq.GetText())
}

func (h *Handler) sendAuthResponse(stream pb.ChatService_ManageChatServer, success bool, errorMsg string, user *pb.User) {
	chatResponse := &pb.ChatResponse{
		ResponseType: &pb.ChatResponse_Auth{
			Auth: &pb.AuthResponse{
				Success: success,
				Error:   errorMsg,
				User:    user,
			},
		},
	}

	if err := stream.Send(chatResponse); err != nil {
		log.Printf("Failed to send auth response: %v", err)
	}
}

func (h *Handler) sendErrorResponse(stream pb.ChatService_ManageChatServer, code, message string) {
	chatResponse := &pb.ChatResponse{
		ResponseType: &pb.ChatResponse_Error{
			Error: &pb.ErrorResponse{
				Code:    code,
				Message: message,
			},
		},
	}

	if err := stream.Send(chatResponse); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
}

func (h *Handler) storeMessage(message *pb.MessageBroadcast) {
	h.msgMutex.Lock()
	defer h.msgMutex.Unlock()

	h.messages = append(h.messages, message)

	if len(h.messages) > 100 {
		h.messages = h.messages[1:]
	}
}
