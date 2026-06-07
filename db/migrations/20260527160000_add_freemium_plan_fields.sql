-- Add freemium plan fields and seed default Free plan
BEGIN;

ALTER TABLE pricing_plans
  ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS ocr_lifetime_limit INT CHECK (ocr_lifetime_limit IS NULL OR ocr_lifetime_limit >= 0);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pricing_plans_single_default
  ON pricing_plans (is_default) WHERE is_default = TRUE;

-- Seed Free plan if no default exists
INSERT INTO pricing_plans (
  pricing_plan_id, name, description, amount, duration_days, status,
  trial_days, token_per_month, ocr_per_day, is_active, is_default, ocr_lifetime_limit
)
SELECT
  'PPL-FREE001',
  'Free',
  'Free tier with lifetime OCR photo limit. Upgrade for more capacity.',
  0,
  36500,
  'active',
  0,
  0,
  0,
  TRUE,
  TRUE,
  10
WHERE NOT EXISTS (
  SELECT 1 FROM pricing_plans WHERE is_default = TRUE
);

-- Backfill: assign Free plan to users without an active subscription
INSERT INTO user_subscriptions (id, user_id, plan_id, status, start_date, end_date)
SELECT
  gen_random_uuid(),
  u.id,
  p.id,
  'active',
  NOW(),
  NOW() + (p.duration_days || ' days')::INTERVAL
FROM users u
CROSS JOIN pricing_plans p
WHERE p.is_default = TRUE
  AND NOT EXISTS (
    SELECT 1 FROM user_subscriptions us
    WHERE us.user_id = u.id AND us.status = 'active' AND us.end_date > NOW()
  );

COMMIT;
