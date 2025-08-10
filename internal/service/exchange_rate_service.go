package service

import (
	"context"
)

// ExchangeRateService defines the interface for exchange rate related operations
type ExchangeRateService interface {
	// GetExchangeRate gets the exchange rate between two currencies
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error)
	
	// ConvertAmount converts an amount from one currency to another using the current exchange rate
	ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, error)
}

// exchangeRateService implements ExchangeRateService
type exchangeRateService struct {
	// Add any dependencies here, like a cache or external API client
}

// NewExchangeRateService creates a new ExchangeRateService
func NewExchangeRateService() ExchangeRateService {
	return &exchangeRateService{}
}

// GetExchangeRate implements ExchangeRateService
func (s *exchangeRateService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	// TODO: Implement actual exchange rate lookup
	// This is a placeholder implementation that returns a fixed rate
	return 1.0, nil
}

// ConvertAmount implements ExchangeRateService
func (s *exchangeRateService) ConvertAmount(ctx context.Context, amount float64, fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	rate, err := s.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}
