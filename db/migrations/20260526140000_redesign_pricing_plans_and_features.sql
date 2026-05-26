-- Migration: Redesign pricing_plans and features tables
BEGIN;

-- Drop dependent tables first
DROP TABLE IF EXISTS subscription_requests CASCADE;
DROP TABLE IF EXISTS user_subscriptions CASCADE;
DROP TABLE IF EXISTS plan_features CASCADE;
DROP TABLE IF EXISTS features CASCADE;
DROP TABLE IF EXISTS pricing_plans CASCADE;

-- Recreate pricing_plans with INT primary key and public pricing_plan_id
CREATE TABLE pricing_plans (
    id SERIAL PRIMARY KEY,
    pricing_plan_id VARCHAR(12) UNIQUE NOT NULL,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    amount NUMERIC(18, 2) NOT NULL CHECK (amount >= 0),
    duration_days INT NOT NULL CHECK (duration_days > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deactivated')),
    trial_days INT NOT NULL DEFAULT 0 CHECK (trial_days >= 0),
    token_per_month NUMERIC(18, 2) NOT NULL DEFAULT 0 CHECK (token_per_month >= 0),
    ocr_per_day INT NOT NULL DEFAULT 0 CHECK (ocr_per_day >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create features table
CREATE TABLE features (
    id SERIAL PRIMARY KEY,
    feature_id VARCHAR(12) UNIQUE NOT NULL,
    description TEXT UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'in-active')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create plan_features junction table
CREATE TABLE plan_features (
    plan_id INT REFERENCES pricing_plans(id) ON DELETE CASCADE,
    feature_id INT REFERENCES features(id) ON DELETE CASCADE,
    PRIMARY KEY (plan_id, feature_id)
);

-- Recreate user_subscriptions referencing INT plan_id
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id INT NOT NULL REFERENCES pricing_plans(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'canceled')),
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Recreate subscription_requests referencing INT plan_id
CREATE TABLE subscription_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id INT NOT NULL REFERENCES pricing_plans(id) ON DELETE RESTRICT,
    payment_link TEXT,
    payment_screenshot TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for maximum performance
CREATE INDEX idx_pricing_plans_active ON pricing_plans(is_active);
CREATE INDEX idx_pricing_plans_business_id ON pricing_plans(pricing_plan_id);
CREATE INDEX idx_features_business_id ON features(feature_id);
CREATE INDEX idx_user_subscriptions_user ON user_subscriptions(user_id, status);
CREATE INDEX idx_subscription_requests_status ON subscription_requests(status);
CREATE INDEX idx_subscription_requests_user ON subscription_requests(user_id);

-- Apply updated_at triggers
CREATE TRIGGER trg_pricing_plans_updated_at
    BEFORE UPDATE ON pricing_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_features_updated_at
    BEFORE UPDATE ON features
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_user_subscriptions_updated_at
    BEFORE UPDATE ON user_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_subscription_requests_updated_at
    BEFORE UPDATE ON subscription_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

COMMIT;
