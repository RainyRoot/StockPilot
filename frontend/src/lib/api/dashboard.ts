import { api } from './client';

export interface PortfolioSummary {
  portfolio_id: number;
  portfolio_name: string;
  currency: string;
  total_value_cents: number;
  total_cost_cents: number;
  unrealized_pl_cents: number;
  unrealized_pl_percent: number;
  realized_pl_cents: number;
  day_change_cents: number;
  day_change_percent: number;
  holdings_count: number;
}

export interface DashboardData {
  summaries: PortfolioSummary[];
  total_value_cents: number;
  total_pl_cents: number;
  total_pl_percent: number;
}

export interface PortfolioSnapshot {
  date: string;
  value_cents: number;
  cost_cents: number;
}

export interface HoldingPL {
  ticker: string;
  name: string;
  unrealized_pl_cents: number;
  unrealized_pl_percent: number;
  weight: number;
}

export async function getDashboard(): Promise<DashboardData> {
  return api<DashboardData>('/dashboard');
}

export async function getPortfolioSummary(id: number): Promise<PortfolioSummary> {
  return api<PortfolioSummary>(`/dashboard/portfolios/${id}/summary`);
}

export async function getHoldingsPL(id: number): Promise<HoldingPL[]> {
  return api<HoldingPL[]>(`/dashboard/portfolios/${id}/holdings-pl`);
}

export async function getPerformance(id: number, range: string = '3m'): Promise<PortfolioSnapshot[]> {
  return api<PortfolioSnapshot[]>(`/dashboard/portfolios/${id}/performance?range=${range}`);
}
