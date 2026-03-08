import { api } from './client';

export interface AllocationTarget {
  id?: number;
  portfolio_id?: number;
  category: string;
  target_percent: number;
}

export interface AllocationActual {
  category: string;
  actual_percent: number;
  value_cents: number;
}

export interface AllocationComparison {
  category: string;
  target_percent: number;
  actual_percent: number;
  diff_percent: number;
  diff_cents: number;
}

export interface RebalanceAction {
  ticker: string;
  action: 'BUY' | 'SELL';
  quantity: number;
  value_cents: number;
  reason: string;
}

export async function getTargets(portfolioId: number): Promise<AllocationTarget[]> {
  return api<AllocationTarget[]>(`/portfolios/${portfolioId}/allocation/targets`);
}

export async function setTargets(portfolioId: number, targets: AllocationTarget[]): Promise<void> {
  await api(`/portfolios/${portfolioId}/allocation/targets`, {
    method: 'PUT',
    body: JSON.stringify({ targets }),
  });
}

export async function getActual(portfolioId: number): Promise<AllocationActual[]> {
  return api<AllocationActual[]>(`/portfolios/${portfolioId}/allocation/actual`);
}

export async function getComparison(portfolioId: number): Promise<AllocationComparison[]> {
  return api<AllocationComparison[]>(`/portfolios/${portfolioId}/allocation/comparison`);
}

export async function getRebalance(portfolioId: number): Promise<RebalanceAction[]> {
  return api<RebalanceAction[]>(`/portfolios/${portfolioId}/allocation/rebalance`);
}
