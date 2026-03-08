CREATE TABLE IF NOT EXISTS stocks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    ticker      TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    exchange    TEXT NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    sector      TEXT,
    industry    TEXT,
    market_cap_cents INTEGER,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_stocks_ticker ON stocks(ticker);

CREATE TABLE IF NOT EXISTS price_cache (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    stock_id    INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    price_cents INTEGER NOT NULL,
    open_cents  INTEGER,
    high_cents  INTEGER,
    low_cents   INTEGER,
    volume      INTEGER,
    change_percent REAL,
    fetched_at  INTEGER NOT NULL,
    UNIQUE(stock_id, fetched_at)
);
CREATE INDEX IF NOT EXISTS idx_price_cache_stock_fetched ON price_cache(stock_id, fetched_at DESC);

CREATE TABLE IF NOT EXISTS price_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    stock_id    INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    date        TEXT NOT NULL,
    open_cents  INTEGER NOT NULL,
    high_cents  INTEGER NOT NULL,
    low_cents   INTEGER NOT NULL,
    close_cents INTEGER NOT NULL,
    adj_close_cents INTEGER NOT NULL,
    volume      INTEGER NOT NULL,
    UNIQUE(stock_id, date)
);
CREATE INDEX IF NOT EXISTS idx_price_history_stock_date ON price_history(stock_id, date DESC);

CREATE TABLE IF NOT EXISTS fundamentals (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    stock_id            INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    fiscal_year         INTEGER NOT NULL,
    revenue_cents       INTEGER,
    net_income_cents    INTEGER,
    eps_cents           INTEGER,
    book_value_cents    INTEGER,
    free_cash_flow_cents INTEGER,
    dividends_cents     INTEGER,
    shares_outstanding  INTEGER,
    debt_cents          INTEGER,
    cash_cents          INTEGER,
    roe                 REAL,
    profit_margin       REAL,
    fetched_at          INTEGER NOT NULL,
    UNIQUE(stock_id, fiscal_year)
);

CREATE TABLE IF NOT EXISTS watchlists (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL DEFAULT 'Default',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS watchlist_items (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    watchlist_id INTEGER NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
    stock_id     INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    added_at     INTEGER NOT NULL,
    notes        TEXT,
    UNIQUE(watchlist_id, stock_id)
);

CREATE TABLE IF NOT EXISTS portfolios (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT,
    currency    TEXT NOT NULL DEFAULT 'EUR',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS trades (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id    INTEGER NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    stock_id        INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    trade_type      TEXT NOT NULL CHECK(trade_type IN ('BUY', 'SELL')),
    quantity        REAL NOT NULL,
    price_cents     INTEGER NOT NULL,
    fee_cents       INTEGER NOT NULL DEFAULT 100,
    currency        TEXT NOT NULL,
    executed_at     INTEGER NOT NULL,
    notes           TEXT,
    created_at      INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_trades_portfolio ON trades(portfolio_id, executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_trades_stock ON trades(stock_id);

CREATE TABLE IF NOT EXISTS allocation_targets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id    INTEGER NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    category        TEXT NOT NULL,
    target_percent  REAL NOT NULL,
    UNIQUE(portfolio_id, category)
);

CREATE TABLE IF NOT EXISTS stock_categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    stock_id    INTEGER NOT NULL REFERENCES stocks(id) ON DELETE CASCADE,
    category    TEXT NOT NULL,
    UNIQUE(stock_id)
);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    from_currency TEXT NOT NULL,
    to_currency   TEXT NOT NULL,
    rate          REAL NOT NULL,
    fetched_at    INTEGER NOT NULL,
    UNIQUE(from_currency, to_currency)
);
