import { api } from './client';

export interface HealthScore {
  portfolio_id: number;
  overall_score: number;
  grade: string;
  components: HealthComponent[];
  alerts: HealthAlert[];
  top_picks: TopPick[];
}

export interface HealthComponent {
  name: string;
  score: number;
  weight: number;
  description: string;
}

export interface HealthAlert {
  severity: 'info' | 'warning' | 'critical';
  title: string;
  message: string;
  ticker?: string;
}

export interface TopPick {
  ticker: string;
  name: string;
  action: 'BUY' | 'HOLD' | 'WATCH';
  reason: string;
  margin_of_safety: number;
  quality_score: number;
}

export async function getHealthScore(portfolioId: number): Promise<HealthScore> {
  return api<HealthScore>(`/portfolios/${portfolioId}/health`);
}

export async function getAlerts(portfolioId: number): Promise<HealthAlert[]> {
  return api<HealthAlert[]>(`/portfolios/${portfolioId}/health/alerts`);
}

export async function getTopPicks(portfolioId: number): Promise<TopPick[]> {
  return api<TopPick[]>(`/portfolios/${portfolioId}/health/top-picks`);
}
