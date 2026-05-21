CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner TEXT NOT NULL,
    repo TEXT NOT NULL,
    email TEXT NOT NULL, 
    token UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    is_confirmed BOOLEAN DEFAULT FALSE,
    last_tag TEXT,
    last_checked TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_token ON subscriptions(token);
CREATE INDEX IF NOT EXISTS idx_subscriptions_email ON subscriptions(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_email_owner_repo ON subscriptions(email, owner, repo);