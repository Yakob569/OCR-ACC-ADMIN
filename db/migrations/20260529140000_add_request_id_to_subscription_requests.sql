-- Public business id for subscription requests (REQ- + 7 chars), set on merchant submit.

ALTER TABLE subscription_requests
    ADD COLUMN IF NOT EXISTS request_id VARCHAR(12);

UPDATE subscription_requests
SET request_id = 'REQ-' || upper(substr(md5(id::text), 1, 7))
WHERE request_id IS NULL;

ALTER TABLE subscription_requests
    ALTER COLUMN request_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_requests_request_id
    ON subscription_requests(request_id);
