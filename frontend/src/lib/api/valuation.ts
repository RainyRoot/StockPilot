import { api } from './client';

export interface MethodResult {
  name: string;
  value_cents: number;
  weight: number;
  description?: string;
}

export type ValuationVerdict = 'UNDERVALUED' | 'FAIRLY_VALUED_LOW' | 'FAIR_VALUE' | 'OVERVALUED';

export interface ValuationResult {
  ticker: string;
  current_price_cents: number;
  intrinsic_value_cents: number;
  margin_of_safety: number;
  verdict: ValuationVerdict;
  quality_score: number;
  methods: MethodResult[];
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
