-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role TEXT NOT NULL CHECK (role IN ('student', 'admin')),
    login_code TEXT UNIQUE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Policies table
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 10 AND 200),
    description TEXT NOT NULL CHECK (char_length(description) BETWEEN 50 AND 2000),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'uncertain')),
    admin_comment TEXT,
    submitted_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Votes table
CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type TEXT NOT NULL CHECK (vote_type IN ('upvote', 'downvote')),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(policy_id, user_id)
);

-- Indexes for performance
CREATE INDEX idx_policies_status ON policies(status);
CREATE INDEX idx_policies_submitted_by ON policies(submitted_by);
CREATE INDEX idx_votes_policy_id ON votes(policy_id);
CREATE INDEX idx_votes_user_id ON votes(user_id);
CREATE INDEX idx_users_login_code ON users(login_code) WHERE login_code IS NOT NULL;

-- Insert sample admin (for testing - remove in production)
-- Password will be managed via Supabase Auth
INSERT INTO users (id, role, login_code, is_active) 
VALUES ('00000000-0000-0000-0000-000000000001', 'admin', NULL, true);

-- Insert sample student codes
INSERT INTO users (role, login_code, is_active) VALUES
('student', 'STUDENT01', true),
('student', 'STUDENT02', true),
('student', 'STUDENT03', true),
('student', 'DEMO2024', true),
('student', 'POLICY99', true);