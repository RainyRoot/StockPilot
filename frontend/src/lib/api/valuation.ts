import { api } from './client';

export interface MethodResult {
  name: string;
  value_cents: number;
  weight: number;
  description?: string;
}

export type ValuationVerdict = 'UNDERVALUED' | 'FAIRLY_VALUED_LOW' | 'FAIR_VALUE' | 'OVERVALUED';

export type TrapFlag =
  | 'DECLINING_REVENUE'
  | 'SHRINKING_MARGINS'
  | 'GROWING_DEBT'
  | 'DECLINING_FCF'
  | 'DECLINING_EARNINGS';

export interface FundamentalTrend {
  is_value_trap: boolean;
  trap_score: number;
  flags: TrapFlag[];
  reasons: string[];
  revenue_cagr: number;
  earnings_cagr: number;
  fcf_trend: string;
  debt_trend: string;
}

export interface ValuationResult {
  ticker: string;
  current_price_cents: number;
  intrinsic_value_cents: number;
  margin_of_safety: number;
  verdict: ValuationVerdict;
  quality_score: number;
  methods: MethodResult[];
  trend?: FundamentalTrend;
}

export interface Fundamentals {
  stock_id: number;
  fiscal_year: number;
  revenue_cents: number;
  net_income_cents: number;
  eps_cents: number;
  book_value_cents: number;
  free_cash_flow_cents: number;
  dividends_cents: number;
  shares_outstanding: number;
  debt_cents: number;
  cash_cents: number;
  roe: number;
  profit_margin: number;
}

export async function getValuation(ticker: string): Promise<ValuationResult> {
  return api<ValuationResult>(`/valuation/${encodeURIComponent(ticker)}`);
}

export async function getFundamentals(ticker: string): Promise<Fundamentals> {
  return api<Fundamentals>(`/valuation/${encodeURIComponent(ticker)}/fundamentals`);
}

export async function screenWatchlist(watchlistId: number): Promise<ValuationResult[]> {
  return api<ValuationResult[]>(`/watchlists/${watchlistId}/screen`);
}

export interface BacktestResult {
  ticker: string;
  entry_date: string;
  entry_price_cents: number;
  current_price_cents: number;
  return_pct: number;
  annualized_return_pct: number;
  holding_days: number;
  entry_valuation: ValuationResult;
  model_correct: boolean;
  model_verdict: string;
  actual_outcome: string;
}

export async function backtestValuation(ticker: string, date: string): Promise<BacktestResult> {
  return api<BacktestResult>(
    `/valuation/${encodeURIComponent(ticker)}/backtest?date=${encodeURIComponent(date)}`
  );
}
