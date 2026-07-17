package repository

import (
	"camera-dashboard-backend/internal/models"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ChatRepository struct {
	db *sqlx.DB
}

func NewChatRepository(db *sqlx.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// GetUserConversations trả về danh sách hội thoại của user, kèm last message + unread count
func (r *ChatRepository) GetUserConversations(userID int64) ([]models.Conversation, error) {
	convs := make([]models.Conversation, 0)
	err := r.db.Select(&convs, `
		SELECT c.*
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id AND cm.user_id = ?
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}

	for i := range convs {
		// Load members
		members, _ := r.GetMembers(convs[i].ID)
		convs[i].Members = members

		// Last message
		var msgs []models.Message
		r.db.Select(&msgs, `
			SELECT m.*, u.name as sender_name
			FROM messages m JOIN users u ON u.id = m.sender_id
			WHERE m.conversation_id = ?
			ORDER BY m.created_at DESC LIMIT 1
		`, convs[i].ID)
		if len(msgs) > 0 {
			convs[i].LastMessage = &msgs[0]
		}

		// Unread count
		var cm models.ConversationMember
		r.db.Get(&cm, `SELECT * FROM conversation_members WHERE conversation_id=? AND user_id=?`, convs[i].ID, userID)
		var unread int
		r.db.Get(&unread, `SELECT COUNT(*) FROM messages WHERE conversation_id=? AND created_at > ? AND sender_id != ?`,
			convs[i].ID, cm.LastReadAt, userID)
		convs[i].UnreadCount = unread

		// DM: lấy thông tin user còn lại
		if convs[i].Type == "dm" {
			for _, m := range members {
				if m.UserID != userID {
					convs[i].OtherUser = &models.UserSummary{
						ID:    m.UserID,
						Name:  m.Name,
						Email: m.Email,
					}
					break
				}
			}
		}
	}
	return convs, nil
}

// FindDMBetween kiểm tra DM đã tồn tại giữa 2 user chưa
func (r *ChatRepository) FindDMBetween(userA, userB int64) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Get(&conv, `
		SELECT c.* FROM conversations c
		JOIN conversation_members a ON a.conversation_id = c.id AND a.user_id = ?
		JOIN conversation_members b ON b.conversation_id = c.id AND b.user_id = ?
		WHERE c.type = 'dm'
		LIMIT 1
	`, userA, userB)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// CreateConversation tạo hội thoại mới + thêm members
func (r *ChatRepository) CreateConversation(conv *models.Conversation, memberIDs []int64) (*models.Conversation, error) {
	res, err := r.db.Exec(`
		INSERT INTO conversations (type, name, created_by) VALUES (?, ?, ?)
	`, conv.Type, conv.Name, conv.CreatedBy)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	conv.ID = id

	for _, uid := range memberIDs {
		r.db.Exec(`INSERT OR IGNORE INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, id, uid)
	}

	members, _ := r.GetMembers(id)
	conv.Members = members
	return conv, nil
}

// GetConversation lấy thông tin hội thoại
func (r *ChatRepository) GetConversation(convID int64) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Get(&conv, `SELECT * FROM conversations WHERE id = ?`, convID)
	if err != nil {
		return nil, err
	}
	members, _ := r.GetMembers(convID)
	conv.Members = members
	return &conv, nil
}

// IsMember kiểm tra user có trong conversation không
func (r *ChatRepository) IsMember(convID, userID int64) bool {
	var count int
	r.db.Get(&count, `SELECT COUNT(*) FROM conversation_members WHERE conversation_id=? AND user_id=?`, convID, userID)
	return count > 0
}

// GetMembers trả về danh sách thành viên của hội thoại
func (r *ChatRepository) GetMembers(convID int64) ([]models.ConversationMember, error) {
	var members []models.ConversationMember
	err := r.db.Select(&members, `
		SELECT cm.*, u.name, u.email
		FROM conversation_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
	`, convID)
	return members, err
}

// AddMember thêm user vào group
func (r *ChatRepository) AddMember(convID, userID int64) error {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, convID, userID)
	return err
}

// RemoveMember xóa user khỏi group
func (r *ChatRepository) RemoveMember(convID, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM conversation_members WHERE conversation_id=? AND user_id=?`, convID, userID)
	return err
}

// RenameGroup đổi tên nhóm
func (r *ChatRepository) RenameGroup(convID int64, name string) error {
	_, err := r.db.Exec(`UPDATE conversations SET name=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, name, convID)
	return err
}

// GetMessages lấy tin nhắn theo trang (20 tin/trang, mới nhất trước)
func (r *ChatRepository) GetMessages(convID int64, limit, offset int) ([]models.Message, error) {
	msgs := make([]models.Message, 0)
	err := r.db.Select(&msgs, `
		SELECT m.*, u.name as sender_name
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = ?
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, convID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Load attachments cho mỗi message
	for i := range msgs {
		var atts []models.Attachment
		r.db.Select(&atts, `SELECT * FROM message_attachments WHERE message_id = ?`, msgs[i].ID)
		msgs[i].Attachments = atts
	}

	// Reverse để trả về thứ tự cũ → mới
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// SaveMessage lưu tin nhắn, cập nhật updated_at của conversation
func (r *ChatRepository) SaveMessage(msg *models.Message) error {
	res, err := r.db.Exec(`
		INSERT INTO messages (conversation_id, sender_id, content, message_type) VALUES (?, ?, ?, ?)
	`, msg.ConversationID, msg.SenderID, msg.Content, msg.MessageType)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	msg.ID = id

	r.db.Exec(`UPDATE conversations SET updated_at=CURRENT_TIMESTAMP WHERE id=?`, msg.ConversationID)
	return nil
}

// SaveAttachment lưu attachment
func (r *ChatRepository) SaveAttachment(att *models.Attachment) error {
	res, err := r.db.Exec(`
		INSERT INTO message_attachments (message_id, filename, original_name, file_size, mime_type, url)
		VALUES (?, ?, ?, ?, ?, ?)
	`, att.MessageID, att.Filename, att.OriginalName, att.FileSize, att.MimeType, att.URL)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	att.ID = id
	return nil
}

// MarkAsRead cập nhật last_read_at
func (r *ChatRepository) MarkAsRead(convID, userID int64) error {
	_, err := r.db.Exec(`
		UPDATE conversation_members SET last_read_at=CURRENT_TIMESTAMP
		WHERE conversation_id=? AND user_id=?
	`, convID, userID)
	return err
}

// GetMemberIDs trả về danh sách user_id trong conversation
func (r *ChatRepository) GetMemberIDs(convID int64) []int64 {
	var ids []int64
	r.db.Select(&ids, `SELECT user_id FROM conversation_members WHERE conversation_id=?`, convID)
	return ids
}

// SearchUsers tìm user theo tên hoặc email, loại trừ currentUser
func (r *ChatRepository) SearchUsers(query string, currentUserID int64) ([]models.UserSummary, error) {
	users := make([]models.UserSummary, 0)
	q := fmt.Sprintf("%%%s%%", query)
	err := r.db.Select(&users, `
		SELECT id, name, email FROM users
		WHERE id != ? AND (name LIKE ? OR email LIKE ?)
		LIMIT 20
	`, currentUserID, q, q)
	return users, err
}
