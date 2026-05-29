ALTER TABLE subscription_requests
    ADD COLUMN IF NOT EXISTS payment_method_id UUID REFERENCES payment_methods(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_subscription_requests_payment_method
    ON subscription_requests(payment_method_id);
