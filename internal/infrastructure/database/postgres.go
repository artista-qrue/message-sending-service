package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"message-sending-service/internal/infrastructure/config"
)

func NewPostgreSQLConnection(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

func CreateTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		content TEXT NOT NULL,
		phone_number VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		sent_at TIMESTAMP WITH TIME ZONE,
		external_message_id VARCHAR(255),
		error_message TEXT,
		
		CONSTRAINT valid_status CHECK (status IN ('pending', 'sent', 'failed')),
		CONSTRAINT valid_content_length CHECK (char_length(content) <= 160)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_messages_phone_number ON messages(phone_number);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}
