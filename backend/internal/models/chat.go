package models

import "time"

type Conversation struct {
	ID        int64      `db:"id" json:"id"`
	Type      string     `db:"type" json:"type"` // "dm" | "group"
	Name      *string    `db:"name" json:"name"`
	CreatedBy int64      `db:"created_by" json:"createdBy"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time  `db:"updated_at" json:"updatedAt"`

	// Joined fields
	Members      []ConversationMember `db:"-" json:"members,omitempty"`
	LastMessage  *Message             `db:"-" json:"lastMessage,omitempty"`
	UnreadCount  int                  `db:"-" json:"unreadCount"`
	OtherUser    *UserSummary         `db:"-" json:"otherUser,omitempty"` // DM only
}

type ConversationMember struct {
	ConversationID int64     `db:"conversation_id" json:"conversationId"`
	UserID         int64     `db:"user_id" json:"userId"`
	JoinedAt       time.Time `db:"joined_at" json:"joinedAt"`
	LastReadAt     time.Time `db:"last_read_at" json:"lastReadAt"`

	// Joined fields
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

type Message struct {
	ID             int64      `db:"id" json:"id"`
	ConversationID int64      `db:"conversation_id" json:"conversationId"`
	SenderID       int64      `db:"sender_id" json:"senderId"`
	Content        *string    `db:"content" json:"content"`
	MessageType    string     `db:"message_type" json:"type"` // "text" | "file" | "image"
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`

	// Joined fields
	SenderName  string        `db:"sender_name" json:"senderName"`
	Attachments []Attachment  `db:"-" json:"attachments,omitempty"`
}

type Attachment struct {
	ID           int64     `db:"id" json:"id"`
	MessageID    int64     `db:"message_id" json:"messageId"`
	Filename     string    `db:"filename" json:"filename"`
	OriginalName string    `db:"original_name" json:"originalName"`
	FileSize     int64     `db:"file_size" json:"fileSize"`
	MimeType     string    `db:"mime_type" json:"mimeType"`
	URL          string    `db:"url" json:"url"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

type UserSummary struct {
	ID    int64  `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

// WebSocket event types
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

const (
	WSEventNewMessage    = "new_message"
	WSEventUserOnline    = "user_online"
	WSEventUserOffline   = "user_offline"
	WSEventUserTyping    = "user_typing"
	WSEventConvCreated   = "conversation_created"
	WSEventMemberAdded   = "member_added"
	WSEventMemberRemoved = "member_removed"
	WSEventGroupRenamed  = "group_renamed"
)

// Request types
type CreateConversationRequest struct {
	Type    string  `json:"type"` // "dm" | "group"
	UserID  int64   `json:"userId"`  // dm: target user
	Name    string  `json:"name"`    // group: tên nhóm
	UserIDs []int64 `json:"userIds"` // group: danh sách thành viên
}

type SendMessageRequest struct {
	Content string `json:"content"`
	Type    string `json:"type"` // "text" | "file" | "image"
}

type RenameGroupRequest struct {
	Name string `json:"name"`
}
