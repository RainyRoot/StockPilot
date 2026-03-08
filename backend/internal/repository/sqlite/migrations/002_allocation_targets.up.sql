CREATE TABLE IF NOT EXISTS allocation_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    target_percent REAL NOT NULL DEFAULT 0,
    UNIQUE(portfolio_id, category)
);
