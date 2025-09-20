package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/domain/repositories"
)

type messageRepositoryImpl struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) repositories.MessageRepository {
	return &messageRepositoryImpl{
		db: db,
	}
}

func (r *messageRepositoryImpl) Create(ctx context.Context, message *entities.Message) error {
	query := `
		INSERT INTO messages (id, content, phone_number, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.Content,
		message.PhoneNumber,
		message.Status,
		message.CreatedAt,
		message.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

func (r *messageRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	query := `
		SELECT id, content, phone_number, status, created_at, updated_at, 
		       sent_at, external_message_id, error_message
		FROM messages 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	message := &entities.Message{}
	err := row.Scan(
		&message.ID,
		&message.Content,
		&message.PhoneNumber,
		&message.Status,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.SentAt,
		&message.ExternalMessageID,
		&message.ErrorMessage,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

func (r *messageRepositoryImpl) GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	query := `
		SELECT id, content, phone_number, status, created_at, updated_at, 
		       sent_at, external_message_id, error_message
		FROM messages 
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", err)
	}
	defer rows.Close()

	var messages []*entities.Message
	for rows.Next() {
		message := &entities.Message{}
		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.PhoneNumber,
			&message.Status,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.SentAt,
			&message.ExternalMessageID,
			&message.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *messageRepositoryImpl) GetSentMessages(ctx context.Context, offset, limit int) ([]*entities.Message, error) {
	query := `
		SELECT id, content, phone_number, status, created_at, updated_at, 
		       sent_at, external_message_id, error_message
		FROM messages 
		WHERE status = 'sent'
		ORDER BY sent_at DESC
		OFFSET $1 LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent messages: %w", err)
	}
	defer rows.Close()

	var messages []*entities.Message
	for rows.Next() {
		message := &entities.Message{}
		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.PhoneNumber,
			&message.Status,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.SentAt,
			&message.ExternalMessageID,
			&message.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *messageRepositoryImpl) Update(ctx context.Context, message *entities.Message) error {
	query := `
		UPDATE messages 
		SET content = $2, phone_number = $3, status = $4, updated_at = $5,
		    sent_at = $6, external_message_id = $7, error_message = $8
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.Content,
		message.PhoneNumber,
		message.Status,
		message.UpdatedAt,
		message.SentAt,
		message.ExternalMessageID,
		message.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrMessageNotFound
	}

	return nil
}

func (r *messageRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrMessageNotFound
	}

	return nil
}

func (r *messageRepositoryImpl) CountByStatus(ctx context.Context, status entities.MessageStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM messages WHERE status = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by status: %w", err)
	}

	return count, nil
}

func (r *messageRepositoryImpl) GetAll(ctx context.Context, offset, limit int) ([]*entities.Message, error) {
	query := `
		SELECT id, content, phone_number, status, created_at, updated_at, 
		       sent_at, external_message_id, error_message
		FROM messages 
		ORDER BY created_at DESC
		OFFSET $1 LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get all messages: %w", err)
	}
	defer rows.Close()

	var messages []*entities.Message
	for rows.Next() {
		message := &entities.Message{}
		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.PhoneNumber,
			&message.Status,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.SentAt,
			&message.ExternalMessageID,
			&message.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}
