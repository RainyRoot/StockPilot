import { api } from './client';
import type { Stock } from './stocks';

export type PortfolioStrategy = 'VALUE' | 'TRADING' | 'INDEX';

export interface Portfolio {
  id: number;
  name: string;
  description?: string;
  currency: string;
  strategy: PortfolioStrategy;
  created_at: string;
  updated_at: string;
}

export interface Trade {
  id: number;
  portfolio_id: number;
  stock_id: number;
  ticker: string;
  trade_type: 'BUY' | 'SELL';
  quantity: number;
  price_cents: number;
  fee_cents: number;
  currency: string;
  executed_at: string;
  notes?: string;
  created_at: string;
}

export interface Holding {
  stock: Stock;
  quantity: number;
  avg_cost_cents: number;
  current_price_cents: number;
  market_value_cents: number;
  unrealized_pl_cents: number;
  unrealized_pl_percent: number;
  category?: string;
}

export interface TradeListResponse {
  trades: Trade[];
  total: number;
  limit: number;
  offset: number;
}

export async function getPortfolios(): Promise<Portfolio[]> {
  return api<Portfolio[]>('/portfolios');
}

export async function getPortfolio(id: number): Promise<Portfolio> {
  return api<Portfolio>(`/portfolios/${id}`);
}

export async function createPortfolio(name: string, description: string = '', currency: string = 'EUR', strategy: PortfolioStrategy = 'VALUE'): Promise<Portfolio> {
  return api<Portfolio>('/portfolios', {
    method: 'POST',
    body: JSON.stringify({ name, description, currency, strategy }),
  });
}

export async function updatePortfolio(id: number, name: string, description: string = ''): Promise<void> {
  await api(`/portfolios/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, description }),
  });
}

export async function deletePortfolio(id: number): Promise<void> {
  await api(`/portfolios/${id}`, { method: 'DELETE' });
}

export async function addTrade(
  portfolioId: number,
  ticker: string,
  tradeType: 'BUY' | 'SELL',
  quantity: number,
  priceCents: number,
  feeCents: number,
  currency: string,
  executedAt: string,
  notes: string = '',
): Promise<Trade> {
  return api<Trade>(`/portfolios/${portfolioId}/trades`, {
    method: 'POST',
    body: JSON.stringify({
      ticker,
      trade_type: tradeType,
      quantity,
      price_cents: priceCents,
      fee_cents: feeCents,
      currency,
      executed_at: executedAt,
      notes,
    }),
  });
}

export async function deleteTrade(portfolioId: number, tradeId: number): Promise<void> {
  await api(`/portfolios/${portfolioId}/trades/${tradeId}`, { method: 'DELETE' });
}

export async function listTrades(portfolioId: number, limit: number = 50, offset: number = 0): Promise<TradeListResponse> {
  return api<TradeListResponse>(`/portfolios/${portfolioId}/trades?limit=${limit}&offset=${offset}`);
}

export async function getHoldings(portfolioId: number): Promise<Holding[]> {
  return api<Holding[]>(`/portfolios/${portfolioId}/holdings`);
}
