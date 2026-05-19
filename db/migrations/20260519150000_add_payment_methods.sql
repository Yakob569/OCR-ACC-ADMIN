-- Migration to add payment_methods table
BEGIN;

CREATE TABLE IF NOT EXISTS payment_methods (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_url TEXT,
    account_number VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert some default/example payment methods for Ethiopian Birr
INSERT INTO payment_methods (id, name, account_number, status)
VALUES 
('c3b3184d-2d41-4775-9276-8051a82f3ef8', 'CBE Birr', '1000123456789', 'active'),
('a6828551-7ab6-476c-843e-c689d04bc7a2', 'Telebirr', '0912345678', 'active')
ON CONFLICT (id) DO NOTHING;

COMMIT;
