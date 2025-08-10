package executors

import (
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
)

// accountSupportsCurrency checks if an account supports a specific currency
func accountSupportsCurrency(account *models.Account, currency string) bool {
	// Check if the account's currency matches the requested currency
	return account.Currency == currency
}
