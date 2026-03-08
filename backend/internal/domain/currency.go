package domain

import "time"

type ExchangeRate struct {
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	FetchedAt    time.Time `json:"fetched_at"`
}
