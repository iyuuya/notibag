package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Domain models
type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// Request/Response types
type CreateNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// WebSocket message types
type WSMessage struct {
	Type           string        `json:"type"`
	Notification   *Notification `json:"notification,omitempty"`
	Notifications  []Notification `json:"notifications,omitempty"`
	NotificationID string        `json:"notification_id,omitempty"`
}

// Repository interface
type NotificationRepository interface {
	GetUnread() []Notification
	GetAll() []Notification
	Create(notification Notification) error
	MarkAsRead(id string) error
	Clear() error
}

// Service interface
type NotificationService interface {
	GetUnreadNotifications() []Notification
	CreateNotification(title, message string) (*Notification, error)
	MarkNotificationAsRead(id string) error
	ClearAllNotifications() error
}

// WebSocket manager interface
type WSManager interface {
	AddClient(conn *websocket.Conn)
	RemoveClient(conn *websocket.Conn)
	BroadcastNotification(notification Notification)
	HandleMessage(conn *websocket.Conn, msg WSMessage) error
}

// In-memory repository implementation
type InMemoryNotificationRepository struct {
	notifications []Notification
	mu           sync.RWMutex
}

func NewInMemoryNotificationRepository() *InMemoryNotificationRepository {
	return &InMemoryNotificationRepository{
		notifications: []Notification{
			{
				ID:        "1",
				Title:     "システム起動",
				Message:   "Notibagが正常に起動しました",
				Timestamp: time.Now().Add(-5 * time.Minute),
				Read:      false,
			},
			{
				ID:        "2",
				Title:     "重要な更新",
				Message:   "新しいバージョンが利用可能です。アップデートを確認してください。",
				Timestamp: time.Now().Add(-2 * time.Minute),
				Read:      false,
			},
		},
	}
}

func (r *InMemoryNotificationRepository) GetUnread() []Notification {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	unread := make([]Notification, 0)
	for _, notification := range r.notifications {
		if !notification.Read {
			unread = append(unread, notification)
		}
	}
	return unread
}

func (r *InMemoryNotificationRepository) GetAll() []Notification {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]Notification, len(r.notifications))
	copy(result, r.notifications)
	return result
}

func (r *InMemoryNotificationRepository) Create(notification Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.notifications = append([]Notification{notification}, r.notifications...)
	return nil
}

func (r *InMemoryNotificationRepository) MarkAsRead(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i := range r.notifications {
		if r.notifications[i].ID == id {
			r.notifications[i].Read = true
			return nil
		}
	}
	return errors.New("notification not found")
}

func (r *InMemoryNotificationRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.notifications = []Notification{}
	return nil
}

// Service implementation
type NotificationServiceImpl struct {
	repo NotificationRepository
}

func NewNotificationService(repo NotificationRepository) *NotificationServiceImpl {
	return &NotificationServiceImpl{repo: repo}
}

func (s *NotificationServiceImpl) GetUnreadNotifications() []Notification {
	return s.repo.GetUnread()
}

func (s *NotificationServiceImpl) CreateNotification(title, message string) (*Notification, error) {
	if title == "" || message == "" {
		return nil, errors.New("title and message are required")
	}
	
	notification := Notification{
		ID:        generateID(),
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
		Read:      false,
	}
	
	if err := s.repo.Create(notification); err != nil {
		return nil, err
	}
	
	return &notification, nil
}

func (s *NotificationServiceImpl) MarkNotificationAsRead(id string) error {
	if id == "" {
		return errors.New("notification ID is required")
	}
	return s.repo.MarkAsRead(id)
}

func (s *NotificationServiceImpl) ClearAllNotifications() error {
	return s.repo.Clear()
}

// WebSocket manager implementation
type WSManagerImpl struct {
	clients   map[*websocket.Conn]bool
	mu        sync.RWMutex
	service   NotificationService
	upgrader  websocket.Upgrader
}

func NewWSManager(service NotificationService) *WSManagerImpl {
	return &WSManagerImpl{
		clients: make(map[*websocket.Conn]bool),
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 開発環境用、本番では適切に設定
			},
		},
	}
}

func (w *WSManagerImpl) AddClient(conn *websocket.Conn) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.clients[conn] = true
}

func (w *WSManagerImpl) RemoveClient(conn *websocket.Conn) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.clients, conn)
}

func (w *WSManagerImpl) BroadcastNotification(notification Notification) {
	message := WSMessage{
		Type:         "notification",
		Notification: &notification,
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	for client := range w.clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			client.Close()
			delete(w.clients, client)
		}
	}
}

func (w *WSManagerImpl) HandleMessage(conn *websocket.Conn, msg WSMessage) error {
	switch msg.Type {
	case "get_notifications":
		notifications := w.service.GetUnreadNotifications()
		response := WSMessage{
			Type:          "notifications_list",
			Notifications: notifications,
		}
		return conn.WriteJSON(response)
		
	case "mark_read":
		if msg.NotificationID != "" {
			return w.service.MarkNotificationAsRead(msg.NotificationID)
		}
		return errors.New("notification ID is required")
		
	case "clear_all":
		return w.service.ClearAllNotifications()
		
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// HTTP handlers
type NotificationHandler struct {
	service   NotificationService
	wsManager WSManager
}

func NewNotificationHandler(service NotificationService, wsManager WSManager) *NotificationHandler {
	return &NotificationHandler{
		service:   service,
		wsManager: wsManager,
	}
}

func (h *NotificationHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Notibag server is running",
	})
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	notification, err := h.service.CreateNotification(req.Title, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// WebSocketクライアントに通知を送信
	h.wsManager.BroadcastNotification(*notification)

	c.JSON(http.StatusCreated, notification)
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	notifications := h.service.GetUnreadNotifications()
	c.JSON(http.StatusOK, NotificationsResponse{Notifications: notifications})
}

func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	// デバッグ用：全ての通知を返す
	repo := h.service.(*NotificationServiceImpl).repo
	notifications := repo.GetAll()
	c.JSON(http.StatusOK, NotificationsResponse{Notifications: notifications})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.MarkNotificationAsRead(id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

func (h *NotificationHandler) ClearAll(c *gin.Context) {
	if err := h.service.ClearAllNotifications(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Success: true})
}

func (h *NotificationHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.wsManager.(*WSManagerImpl).upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// クライアントを登録
	h.wsManager.AddClient(conn)
	log.Println("WebSocket connection established")

	// 接続解除時にクライアントを削除
	defer h.wsManager.RemoveClient(conn)

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		if err := h.wsManager.HandleMessage(conn, msg); err != nil {
			log.Printf("WebSocket message handling error: %v", err)
		}
	}
}

func setupCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func main() {
	// 依存関係の注入
	repo := NewInMemoryNotificationRepository()
	service := NewNotificationService(repo)
	wsManager := NewWSManager(service)
	handler := NewNotificationHandler(service, wsManager)

	r := gin.Default()
	r.Use(setupCORS())

	// API routes
	api := r.Group("/api")
	{
		api.GET("/health", handler.HealthCheck)
		api.POST("/notifications", handler.CreateNotification)
		api.GET("/notifications", handler.GetNotifications)
		api.GET("/notifications/all", handler.GetAllNotifications) // デバッグ用
		api.PUT("/notifications/:id/read", handler.MarkAsRead)
		api.DELETE("/notifications", handler.ClearAll)
	}

	// WebSocket endpoint
	r.GET("/ws", handler.HandleWebSocket)

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

// Utility functions
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}