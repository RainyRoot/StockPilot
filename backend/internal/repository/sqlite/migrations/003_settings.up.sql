CREATE TABLE IF NOT EXISTS user_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT OR IGNORE INTO user_settings (key, value) VALUES
    ('currency', 'USD'),
    ('discount_rate', '0.10'),
    ('growth_rate', '0.05'),
    ('terminal_growth', '0.025'),
    ('theme', 'dark');
