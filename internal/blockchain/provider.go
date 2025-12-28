package blockchain

import (
	"context"
)

// Network представляет тип блокчейн сети
type Network string

const (
	NetworkTON      Network = "TON"
	NetworkEthereum Network = "ETH"
	NetworkBSC      Network = "BSC"
	NetworkTron     Network = "TRX"
	NetworkSolana   Network = "SOL"
)

// TransactionStatus статус транзакции в блокчейне
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusConfirmed TransactionStatus = "confirmed"
	TxStatusFailed    TransactionStatus = "failed"
	TxStatusNotFound  TransactionStatus = "not_found"
)

// TransactionInfo информация о транзакции в блокчейне
type TransactionInfo struct {
	Hash        string            `json:"hash"`
	From        string            `json:"from"`
	To          string            `json:"to"`
	Amount      float64           `json:"amount"`
	Currency    string            `json:"currency"`
	Status      TransactionStatus `json:"status"`
	BlockNumber uint64            `json:"block_number"`
	Timestamp   int64             `json:"timestamp"`
	Fee         float64           `json:"fee"`
	Network     Network           `json:"network"`
	Comment     string            `json:"comment,omitempty"` // Added for TON comments
}

// WithdrawRequest запрос на вывод средств
type WithdrawRequest struct {
	UserID        uint64  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	ToAddress     string  `json:"to_address"`
	TransactionID string  `json:"transaction_id"`
	Memo          string  `json:"memo,omitempty"`
}

// WithdrawResult результат операции вывода
type WithdrawResult struct {
	Success       bool    `json:"success"`
	TransactionID string  `json:"transaction_id"`
	TxHash        string  `json:"tx_hash"`
	Fee           float64 `json:"fee"`
	ErrorMessage  string  `json:"error_message,omitempty"`
}

// DepositAddress адрес для депозита
type DepositAddress struct {
	Address  string  `json:"address"`
	Memo     string  `json:"memo,omitempty"`
	Network  Network `json:"network"`
	Currency string  `json:"currency"`
	QRCode   string  `json:"qr_code,omitempty"`
}

// BlockchainProvider интерфейс для работы с блокчейном
// Реализуйте этот интерфейс для поддержки разных блокчейнов (TON, ETH, BSC, etc.)
type BlockchainProvider interface {
	// GetNetwork возвращает тип сети блокчейна
	GetNetwork() Network

	// GetSupportedCurrencies возвращает список поддерживаемых валют
	GetSupportedCurrencies() []string

	// VerifyTransaction проверяет транзакцию в блокчейне
	VerifyTransaction(ctx context.Context, txHash string) (*TransactionInfo, error)

	// IsTransactionProcessed проверяет, была ли транзакция уже обработана
	IsTransactionProcessed(ctx context.Context, txHash string) (bool, error)

	// GenerateDepositAddress генерирует адрес для депозита пользователя
	GenerateDepositAddress(ctx context.Context, userID uint64, currency string) (*DepositAddress, error)

	// GetDepositAddress возвращает существующий адрес депозита пользователя
	GetDepositAddress(ctx context.Context, userID uint64, currency string) (*DepositAddress, error)

	// ProcessWithdraw инициирует вывод средств
	ProcessWithdraw(ctx context.Context, request *WithdrawRequest) (*WithdrawResult, error)

	// GetTransactionStatus получает статус транзакции
	GetTransactionStatus(ctx context.Context, txHash string) (TransactionStatus, error)

	// ValidateAddress проверяет валидность адреса кошелька
	ValidateAddress(ctx context.Context, address string) (bool, error)

	// GetBalance получает баланс адреса (опционально)
	GetBalance(ctx context.Context, address string, currency string) (float64, error)

	// GetMinWithdrawAmount возвращает минимальную сумму вывода
	GetMinWithdrawAmount(currency string) float64

	// GetWithdrawFee возвращает комиссию за вывод
	GetWithdrawFee(currency string, amount float64) float64

	// GetRequiredConfirmations возвращает количество необходимых подтверждений
	GetRequiredConfirmations() int

	// GetRecentTransactions returns the latest incoming transactions for the custodial wallet
	GetRecentTransactions(ctx context.Context, limit int) ([]TransactionInfo, error)
}

// MultiChainProvider провайдер, поддерживающий несколько блокчейнов
type MultiChainProvider struct {
	providers map[Network]BlockchainProvider
	primary   Network
}

// NewMultiChainProvider создает новый мультичейн провайдер
func NewMultiChainProvider(primary Network) *MultiChainProvider {
	return &MultiChainProvider{
		providers: make(map[Network]BlockchainProvider),
		primary:   primary,
	}
}

// RegisterProvider регистрирует провайдер для сети
func (m *MultiChainProvider) RegisterProvider(network Network, provider BlockchainProvider) {
	m.providers[network] = provider
}

// GetProvider возвращает провайдер для указанной сети
func (m *MultiChainProvider) GetProvider(network Network) (BlockchainProvider, bool) {
	provider, ok := m.providers[network]
	return provider, ok
}

// GetPrimaryProvider возвращает основной провайдер
func (m *MultiChainProvider) GetPrimaryProvider() (BlockchainProvider, bool) {
	return m.GetProvider(m.primary)
}

// GetAllProviders возвращает все зарегистрированные провайдеры
func (m *MultiChainProvider) GetAllProviders() map[Network]BlockchainProvider {
	return m.providers
}

// GetSupportedNetworks возвращает список поддерживаемых сетей
func (m *MultiChainProvider) GetSupportedNetworks() []Network {
	networks := make([]Network, 0, len(m.providers))
	for network := range m.providers {
		networks = append(networks, network)
	}
	return networks
}
