import { api } from './client';
import type { Stock, Quote } from './stocks';

export interface Watchlist {
  id: number;
  name: string;
  items?: WatchlistItem[];
  created_at: string;
  updated_at: string;
}

export interface WatchlistItem {
  id: number;
  watchlist_id: number;
  stock_id: number;
  stock?: Stock;
  quote?: Quote;
  added_at: string;
  notes?: string;
}

export async function getWatchlists(): Promise<Watchlist[]> {
  return api<Watchlist[]>('/watchlists');
}

export async function getWatchlist(id: number): Promise<Watchlist> {
  return api<Watchlist>(`/watchlists/${id}`);
}

export async function createWatchlist(name: string): Promise<Watchlist> {
  return api<Watchlist>('/watchlists', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export async function updateWatchlist(id: number, name: string): Promise<void> {
  await api(`/watchlists/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name }),
  });
}

export async function deleteWatchlist(id: number): Promise<void> {
  await api(`/watchlists/${id}`, {
    method: 'DELETE',
  });
}

export async function addStockToWatchlist(watchlistId: number, ticker: string, notes: string = ''): Promise<void> {
  await api(`/watchlists/${watchlistId}/stocks`, {
    method: 'POST',
    body: JSON.stringify({ ticker, notes }),
  });
}

export async function removeStockFromWatchlist(watchlistId: number, ticker: string): Promise<void> {
  await api(`/watchlists/${watchlistId}/stocks`, {
    method: 'DELETE',
    body: JSON.stringify({ ticker }),
  });
}
