-- llm-todo Schema v1
-- Simple sessions + todos model

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL,
    applied_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    type TEXT DEFAULT 'quick',        -- quick, code, research
    goal TEXT,
    success_criteria TEXT,
    boundaries TEXT,                  -- JSON array
    deliverables TEXT,                -- JSON array (research mode)
    status TEXT DEFAULT 'active',     -- active, completed
    metadata TEXT DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    type TEXT DEFAULT 'task',         -- task, research, coordination, analysis, deliverable
    priority TEXT DEFAULT 'p0',       -- p0, p1, p2, p3, p4
    priority_order INTEGER DEFAULT 100, -- Sub-ordering within priority (100, 200, 300...)
    status TEXT DEFAULT 'pending',    -- pending, in_progress, completed, blocked
    task TEXT NOT NULL,
    active_form TEXT,

    -- Code mode fields
    files TEXT,                       -- JSON array

    -- Research mode fields
    refs TEXT,                        -- JSON array (URLs, docs)
    waiting_on TEXT,
    output TEXT,                      -- Deliverable output
    audience TEXT,                    -- Deliverable audience

    -- Universal fields
    instructions TEXT,                -- JSON: {must_do: [...], must_not_do: [...]}
    notes TEXT,
    blocking_reason TEXT,
    dependant_ids TEXT,               -- JSON array of task IDs
    effort TEXT,                      -- xs, s, m
    metadata TEXT DEFAULT '{}',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX IF NOT EXISTS idx_session_priority ON todos(session_id, priority, priority_order);
CREATE INDEX IF NOT EXISTS idx_session_status ON todos(session_id, status);
