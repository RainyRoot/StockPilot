import { api } from './client';

export interface UserSettings {
  currency: string;
  discount_rate: number;
  growth_rate: number;
  terminal_growth: number;
  max_position_pct: number;
  theme: string;
}

export async function getSettings(): Promise<UserSettings> {
  return api<UserSettings>('/settings');
}

export async function updateSettings(settings: UserSettings): Promise<void> {
  await api('/settings', {
    method: 'PUT',
    body: JSON.stringify(settings),
  });
}
