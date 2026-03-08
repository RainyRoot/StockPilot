import { api } from './client';

export interface Stock {
  id: number;
  ticker: string;
  name: string;
  exchange: string;
  currency: string;
  sector?: string;
  industry?: string;
}

export interface Quote {
  ticker: string;
  name: string;
  price_cents: number;
  open_cents: number;
  high_cents: number;
  low_cents: number;
  volume: number;
  change_percent: number;
  currency: string;
  exchange: string;
}

export interface StockSearchResult {
  ticker: string;
  name: string;
  exchange: string;
  type: string;
}

export interface StockQuoteResponse {
  stock: Stock;
  quote: Quote;
}

export async function searchStocks(query: string): Promise<StockSearchResult[]> {
  return api<StockSearchResult[]>(`/stocks/search?q=${encodeURIComponent(query)}`);
}

export async function getStockQuote(ticker: string): Promise<StockQuoteResponse> {
  return api<StockQuoteResponse>(`/stocks/${encodeURIComponent(ticker)}`);
}
