package chat

import (
	"go-chat-service/pkg/pb"
	"log"
	"sync"
)

type Connection struct {
	UserID   string
	Username string
	Stream   pb.ChatService_ManageChatServer
}

type ConnectionManager struct {
	connections map[string]*Connection
	mutex       sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
	}
}

func (cm *ConnectionManager) AddConnection(userID, username string, stream pb.ChatService_ManageChatServer) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.connections[userID] = &Connection{
		UserID:   userID,
		Username: username,
		Stream:   stream,
	}

	log.Printf("User %s connected (%d total)", username, len(cm.connections))
}

func (cm *ConnectionManager) RemoveConnection(userID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	username := cm.connections[userID].Username
	delete(cm.connections, userID)
	log.Printf("User %s disconnected (%d total)", username, len(cm.connections))
}

func (cm *ConnectionManager) BroadcastMessage(message *pb.MessageBroadcast, senderUserID string) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	response := &pb.ChatResponse{
		ResponseType: &pb.ChatResponse_Message{Message: message},
	}

	for userID, conn := range cm.connections {
		if userID != senderUserID {
			if err := conn.Stream.Send(response); err != nil {
				log.Printf("Failed to send to %s: %v", conn.Username, err)
				go cm.RemoveConnection(userID)
			}
		}
	}
}

func (cm *ConnectionManager) BroadcastSystemNotification(notification *pb.SystemNotification, UserID string) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	response := &pb.ChatResponse{
		ResponseType: &pb.ChatResponse_System{System: notification},
	}

	for userID, conn := range cm.connections {
		if userID != UserID {
			if err := conn.Stream.Send(response); err != nil {
				log.Printf("Failed to send to %s: %v", conn.Username, err)
				go cm.RemoveConnection(userID)
			}
		}
	}
}
