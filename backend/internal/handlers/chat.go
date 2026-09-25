package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"camera-dashboard-backend/internal/middleware"
	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/internal/services"
	jwtpkg "camera-dashboard-backend/pkg/jwt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatRepo       *repository.ChatRepository
	recordingsPath string
	hub            *services.ChatHub
	jwtSecret      string
}

func NewChatHandler(chatRepo *repository.ChatRepository, recordingsPath string, hub *services.ChatHub, jwtSecret string) *ChatHandler {
	return &ChatHandler{
		chatRepo:       chatRepo,
		recordingsPath: recordingsPath,
		hub:            hub,
		jwtSecret:      jwtSecret,
	}
}

func (h *ChatHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ChatHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

func currentUserID(r *http.Request) int64 {
	claims := middleware.GetClaims(r)
	if claims == nil {
		return 0
	}
	return claims.UserID
}

// ServeWS — WebSocket endpoint (token qua query param vì browser WS không hỗ trợ header)
func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	// Fallback: đọc token từ query param
	if userID == 0 {
		if token := r.URL.Query().Get("token"); token != "" {
			claims, err := jwtpkg.ValidateToken(token, h.jwtSecret)
			if err == nil {
				userID = claims.UserID
			}
		}
	}
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	h.hub.ServeWS(w, r, userID)
}

// GetOnlineUsers — danh sách user đang online
func (h *ChatHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"onlineUserIds": h.hub.OnlineUserIDs(),
	})
}

// SearchUsers — tìm user để nhắn tin
func (h *ChatHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		h.writeJSON(w, http.StatusOK, []models.UserSummary{})
		return
	}
	users, err := h.chatRepo.SearchUsers(q, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, users)
}

// GetConversations — danh sách hội thoại của user
func (h *ChatHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convs, err := h.chatRepo.GetUserConversations(userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, convs)
}

// CreateConversation — tạo DM hoặc Group
func (h *ChatHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	var req models.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Type == "dm" {
		if req.UserID == 0 {
			h.writeError(w, http.StatusBadRequest, "userId required")
			return
		}
		// Kiểm tra DM đã tồn tại chưa
		existing, err := h.chatRepo.FindDMBetween(userID, req.UserID)
		if err == nil && existing != nil {
			members, _ := h.chatRepo.GetMembers(existing.ID)
			existing.Members = members
			for _, m := range members {
				if m.UserID != userID {
					existing.OtherUser = &models.UserSummary{ID: m.UserID, Name: m.Name, Email: m.Email}
					break
				}
			}
			h.writeJSON(w, http.StatusOK, existing)
			return
		}
		// Tạo mới
		conv := &models.Conversation{Type: "dm", CreatedBy: userID}
		created, err := h.chatRepo.CreateConversation(conv, []int64{userID, req.UserID})
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, m := range created.Members {
			if m.UserID != userID {
				created.OtherUser = &models.UserSummary{ID: m.UserID, Name: m.Name, Email: m.Email}
				break
			}
		}
		h.hub.SendToUsers([]int64{req.UserID, userID}, models.WSEvent{
			Type: models.WSEventConvCreated, Payload: created,
		})
		h.writeJSON(w, http.StatusCreated, created)
		return
	}

	if req.Type == "group" {
		if strings.TrimSpace(req.Name) == "" {
			h.writeError(w, http.StatusBadRequest, "group name required")
			return
		}
		name := req.Name
		memberIDs := append(req.UserIDs, userID)
		// deduplicate
		seen := map[int64]bool{}
		unique := []int64{}
		for _, id := range memberIDs {
			if !seen[id] {
				seen[id] = true
				unique = append(unique, id)
			}
		}
		conv := &models.Conversation{Type: "group", Name: &name, CreatedBy: userID}
		created, err := h.chatRepo.CreateConversation(conv, unique)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.hub.SendToUsers(unique, models.WSEvent{
			Type: models.WSEventConvCreated, Payload: created,
		})
		h.writeJSON(w, http.StatusCreated, created)
		return
	}

	h.writeError(w, http.StatusBadRequest, "type must be dm or group")
}

// RenameGroup — đổi tên nhóm
func (h *ChatHandler) RenameGroup(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	conv, err := h.chatRepo.GetConversation(convID)
	if err != nil || conv.Type != "group" {
		h.writeError(w, http.StatusNotFound, "group not found")
		return
	}
	if !h.chatRepo.IsMember(convID, userID) {
		h.writeError(w, http.StatusForbidden, "not a member")
		return
	}

	var req models.RenameGroupRequest
	json.NewDecoder(r.Body).Decode(&req)
	if strings.TrimSpace(req.Name) == "" {
		h.writeError(w, http.StatusBadRequest, "name required")
		return
	}

	h.chatRepo.RenameGroup(convID, req.Name)
	memberIDs := h.chatRepo.GetMemberIDs(convID)
	h.hub.SendToUsers(memberIDs, models.WSEvent{
		Type:    models.WSEventGroupRenamed,
		Payload: map[string]interface{}{"conversationId": convID, "name": req.Name},
	})
	h.writeJSON(w, http.StatusOK, map[string]string{"name": req.Name})
}

// AddMember — thêm thành viên vào group
func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	targetID, _ := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)

	conv, err := h.chatRepo.GetConversation(convID)
	if err != nil || conv.Type != "group" {
		h.writeError(w, http.StatusNotFound, "group not found")
		return
	}
	if !h.chatRepo.IsMember(convID, userID) {
		h.writeError(w, http.StatusForbidden, "not a member")
		return
	}

	h.chatRepo.AddMember(convID, targetID)
	memberIDs := h.chatRepo.GetMemberIDs(convID)
	h.hub.SendToUsers(memberIDs, models.WSEvent{
		Type:    models.WSEventMemberAdded,
		Payload: map[string]interface{}{"conversationId": convID, "userId": targetID},
	})
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RemoveMember — xóa thành viên khỏi group
func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	targetID, _ := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)

	conv, err := h.chatRepo.GetConversation(convID)
	if err != nil || conv.Type != "group" {
		h.writeError(w, http.StatusNotFound, "group not found")
		return
	}
	// Chỉ admin group (creator) hoặc tự rời
	if conv.CreatedBy != userID && targetID != userID {
		h.writeError(w, http.StatusForbidden, "only group creator can remove members")
		return
	}

	memberIDs := h.chatRepo.GetMemberIDs(convID)
	h.chatRepo.RemoveMember(convID, targetID)
	h.hub.SendToUsers(memberIDs, models.WSEvent{
		Type:    models.WSEventMemberRemoved,
		Payload: map[string]interface{}{"conversationId": convID, "userId": targetID},
	})
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetMessages — lịch sử tin nhắn
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if !h.chatRepo.IsMember(convID, userID) {
		log.Printf("[GetMessages] FORBIDDEN convID=%d userID=%d — not a member", convID, userID)
		h.writeError(w, http.StatusForbidden, fmt.Sprintf("not a member (convID=%d userID=%d)", convID, userID))
		return
	}

	limit := 30
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	msgs, err := h.chatRepo.GetMessages(convID, limit, offset)
	if err != nil {
		log.Printf("[GetMessages] convID=%d userID=%d error: %v", convID, userID, err)
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, msgs)
}

// SendMessage — gửi tin nhắn text
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if !h.chatRepo.IsMember(convID, userID) {
		h.writeError(w, http.StatusForbidden, "not a member")
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		h.writeError(w, http.StatusBadRequest, "content required")
		return
	}

	claims := middleware.GetClaims(r)
	msg := &models.Message{
		ConversationID: convID,
		SenderID:       userID,
		Content:        &req.Content,
		MessageType:    "text",
		SenderName:     claims.Name,
	}
	if err := h.chatRepo.SaveMessage(msg); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	memberIDs := h.chatRepo.GetMemberIDs(convID)
	h.hub.SendToUsers(memberIDs, models.WSEvent{
		Type: models.WSEventNewMessage, Payload: msg,
	})
	h.writeJSON(w, http.StatusCreated, msg)
}

// UploadAttachment — upload file/ảnh trong chat
func (h *ChatHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if !h.chatRepo.IsMember(convID, userID) {
		h.writeError(w, http.StatusForbidden, "not a member")
		return
	}

	r.ParseMultipartForm(50 << 20) // 50MB max
	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	uploadDir := filepath.Join(h.recordingsPath, "chat_attachments")
	os.MkdirAll(uploadDir, 0755)

	destPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "cannot save file")
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Xác định type
	mimeType := header.Header.Get("Content-Type")
	msgType := "file"
	if strings.HasPrefix(mimeType, "image/") {
		msgType = "image"
	}

	claims := middleware.GetClaims(r)
	contentStr := header.Filename
	msg := &models.Message{
		ConversationID: convID,
		SenderID:       userID,
		Content:        &contentStr,
		MessageType:    msgType,
		SenderName:     claims.Name,
		CreatedAt:      time.Now(),
	}
	if err := h.chatRepo.SaveMessage(msg); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	att := &models.Attachment{
		MessageID:    msg.ID,
		Filename:     filename,
		OriginalName: header.Filename,
		FileSize:     header.Size,
		MimeType:     mimeType,
		URL:          fmt.Sprintf("/api/chat/attachments/%s", filename),
	}
	h.chatRepo.SaveAttachment(att)
	msg.Attachments = []models.Attachment{*att}

	memberIDs := h.chatRepo.GetMemberIDs(convID)
	h.hub.SendToUsers(memberIDs, models.WSEvent{
		Type: models.WSEventNewMessage, Payload: msg,
	})
	h.writeJSON(w, http.StatusCreated, msg)
}

// ServeAttachment — serve file đính kèm
func (h *ChatHandler) ServeAttachment(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if strings.Contains(filename, "..") {
		http.Error(w, "invalid", http.StatusBadRequest)
		return
	}
	path := filepath.Join(h.recordingsPath, "chat_attachments", filename)
	http.ServeFile(w, r, path)
}

// MarkRead — đánh dấu đã đọc
func (h *ChatHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	convID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.chatRepo.MarkAsRead(convID, userID)
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
