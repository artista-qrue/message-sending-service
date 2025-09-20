CREATE
EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS messages
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    content TEXT NOT NULL,
    phone_number VARCHAR
(
    20
) NOT NULL,
    status VARCHAR
(
    20
) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP
                         WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP
                         WITH TIME ZONE,
                             external_message_id VARCHAR (255),
    error_message TEXT,
    CONSTRAINT valid_status CHECK
(
    status
    IN
(
    'pending',
    'sent',
    'failed'
)),
    CONSTRAINT valid_content_length CHECK
(
    char_length
(
    content
) <= 160)
    );

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_phone_number ON messages(phone_number);
CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at);

CREATE
OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at
= NOW();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;
CREATE TRIGGER update_messages_updated_at
    BEFORE UPDATE
    ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO messages (content, phone_number, status)
VALUES ('Test message 1', '+1234567890', 'pending'),
       ('Test message 2', '+1234567891', 'pending'),
       ('Test message 3', '+1234567892', 'pending'),
       ('Test message 4', '+1234567893', 'pending'),
       ('Test message 5', '+1234567894', 'pending'),
       ('Sample sent message', '+1234567895', 'sent'),
       ('Sample failed message', '+1234567896', 'failed') ON CONFLICT DO NOTHING;
