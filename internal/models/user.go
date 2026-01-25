package models

import (
	"time"
)

// User представляет собой модель пользователя в системе
type User struct {
	TelegramID          uint64     `json:"telegram_id" db:"telegram_id"`
	Username            string     `json:"username" db:"username"`
	FirstName           string     `json:"first_name" db:"first_name"`
	LastName            string     `json:"last_name" db:"last_name"`
	AuthProvider        string     `json:"auth_provider" db:"auth_provider"` // Источник авторизации (telegram/base)
	BaseFID             uint64     `json:"base_fid,omitempty" db:"base_fid"` // Farcaster FID (Base Mini App)
	BaseUsername        string     `json:"base_username,omitempty" db:"base_username"`
	BaseDisplayName     string     `json:"base_display_name,omitempty" db:"base_display_name"`
	BasePfpURL          string     `json:"base_pfp_url,omitempty" db:"base_pfp_url"`
	BaseCustodyAddr     string     `json:"base_custody_address,omitempty" db:"base_custody_address"`
	BaseSmartWallet     string     `json:"base_smart_wallet_address,omitempty" db:"base_smart_wallet_address"`
	Wallet              string     `json:"wallet" db:"wallet"`                                         // Основной кошелек (TON/EVM)
	BalanceTon          float64    `json:"balance_ton" db:"balance_ton"`                               // Баланс в TON
	BalanceUsdt         float64    `json:"balance_usdt" db:"balance_usdt"`                             // Баланс в USDT
	BalanceUsdc         float64    `json:"balance_usdc" db:"balance_usdc"`                             // Баланс в USDC
	PendingWithdrawal   float64    `json:"pending_withdrawal" db:"pending_withdrawal"`                 // Сумма в процессе вывода
	WithdrawalLockUntil *time.Time `json:"withdrawal_lock_until,omitempty" db:"withdrawal_lock_until"` // Блокировка вывода до
	Wins                int        `json:"wins" db:"wins"`                                             // Количество побед
	Losses              int        `json:"losses" db:"losses"`                                         // Количество поражений
	TotalDeposited      float64    `json:"total_deposited" db:"total_deposited"`                       // Всего депозитов
	TotalWithdrawn      float64    `json:"total_withdrawn" db:"total_withdrawn"`                       // Всего выведено
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// CanWithdraw проверяет, может ли пользователь сделать вывод
func (u *User) CanWithdraw() bool {
	if u.WithdrawalLockUntil != nil && time.Now().Before(*u.WithdrawalLockUntil) {
		return false
	}
	return u.PendingWithdrawal == 0
}

// GetAvailableBalance возвращает доступный баланс (за вычетом pending)
func (u *User) GetAvailableBalance(currency string) float64 {
	switch currency {
	case CurrencyTON:
		return u.BalanceTon - u.PendingWithdrawal
	case CurrencyUSDT:
		return u.BalanceUsdt
	case CurrencyUSDC:
		return u.BalanceUsdc
	default:
		return 0
	}
}

// HasSufficientBalance проверяет, достаточно ли средств
func (u *User) HasSufficientBalance(amount float64, currency string) bool {
	return u.GetAvailableBalance(currency) >= amount
}
