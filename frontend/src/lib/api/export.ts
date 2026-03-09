const BASE_URL = 'http://localhost:8080/api/v1';

export async function downloadHoldingsCSV(portfolioId: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/portfolios/${portfolioId}/export/holdings.csv`);
  const blob = await res.blob();
  downloadBlob(blob, `holdings-portfolio-${portfolioId}.csv`);
}

export async function downloadTradesCSV(portfolioId: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/portfolios/${portfolioId}/export/trades.csv`);
  const blob = await res.blob();
  downloadBlob(blob, `trades-portfolio-${portfolioId}.csv`);
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
