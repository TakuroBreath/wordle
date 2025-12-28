package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/blockchain"
)

// MockProvider mock-реализация BlockchainProvider для тестирования
type MockProvider struct {
	network          blockchain.Network
	currencies       []string
	processedTxCache sync.Map
	deposits         sync.Map // userID -> DepositAddress
	withdrawals      sync.Map // txID -> WithdrawResult
}

// NewMockProvider создает новый mock провайдер
func NewMockProvider(network blockchain.Network, currencies []string) *MockProvider {
	return &MockProvider{
		network:    network,
		currencies: currencies,
	}
}

// GetNetwork возвращает тип сети
func (p *MockProvider) GetNetwork() blockchain.Network {
	return p.network
}

// GetSupportedCurrencies возвращает поддерживаемые валюты
func (p *MockProvider) GetSupportedCurrencies() []string {
	return p.currencies
}

// VerifyTransaction проверяет транзакцию (mock всегда возвращает успех)
func (p *MockProvider) VerifyTransaction(ctx context.Context, txHash string) (*blockchain.TransactionInfo, error) {
	if txHash == "" {
		return nil, fmt.Errorf("transaction hash cannot be empty")
	}

	// Mock: возвращаем успешную транзакцию
	return &blockchain.TransactionInfo{
		Hash:        txHash,
		From:        "mock_from_address",
		To:          "mock_to_address",
		Amount:      1.0,
		Currency:    "TON",
		Status:      blockchain.TxStatusConfirmed,
		BlockNumber: 12345678,
		Timestamp:   time.Now().Unix(),
		Fee:         0.01,
		Network:     p.network,
	}, nil
}

// IsTransactionProcessed проверяет, была ли транзакция уже обработана
func (p *MockProvider) IsTransactionProcessed(ctx context.Context, txHash string) (bool, error) {
	_, exists := p.processedTxCache.Load(txHash)
	return exists, nil
}

// MarkTransactionProcessed отмечает транзакцию как обработанную
func (p *MockProvider) MarkTransactionProcessed(txHash string) {
	p.processedTxCache.Store(txHash, true)
}

// GenerateDepositAddress генерирует адрес для депозита
func (p *MockProvider) GenerateDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	address := &blockchain.DepositAddress{
		Address:  fmt.Sprintf("mock_%s_address_%d", p.network, userID),
		Memo:     fmt.Sprintf("deposit_%d_%d", userID, time.Now().Unix()),
		Network:  p.network,
		Currency: currency,
	}
	p.deposits.Store(userID, address)
	return address, nil
}

// GetDepositAddress возвращает существующий адрес депозита
func (p *MockProvider) GetDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	if addr, ok := p.deposits.Load(userID); ok {
		return addr.(*blockchain.DepositAddress), nil
	}
	return p.GenerateDepositAddress(ctx, userID, currency)
}

// ProcessWithdraw инициирует вывод средств (mock)
func (p *MockProvider) ProcessWithdraw(ctx context.Context, request *blockchain.WithdrawRequest) (*blockchain.WithdrawResult, error) {
	if request == nil {
		return nil, fmt.Errorf("withdraw request cannot be nil")
	}

	if request.Amount <= 0 {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: "amount must be positive",
		}, nil
	}

	if request.ToAddress == "" {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: "destination address is required",
		}, nil
	}

	result := &blockchain.WithdrawResult{
		Success:       true,
		TransactionID: request.TransactionID,
		TxHash:        fmt.Sprintf("mock_tx_%s_%d", request.TransactionID, time.Now().UnixNano()),
		Fee:           p.GetWithdrawFee(request.Currency, request.Amount),
	}

	p.withdrawals.Store(request.TransactionID, result)
	return result, nil
}

// GetTransactionStatus получает статус транзакции (mock всегда confirmed)
func (p *MockProvider) GetTransactionStatus(ctx context.Context, txHash string) (blockchain.TransactionStatus, error) {
	return blockchain.TxStatusConfirmed, nil
}

// ValidateAddress проверяет валидность адреса (mock всегда true)
func (p *MockProvider) ValidateAddress(ctx context.Context, address string) (bool, error) {
	return address != "", nil
}

// GetBalance получает баланс адреса (mock возвращает фиксированный баланс)
func (p *MockProvider) GetBalance(ctx context.Context, address string, currency string) (float64, error) {
	return 100.0, nil
}

// GetMinWithdrawAmount возвращает минимальную сумму вывода
func (p *MockProvider) GetMinWithdrawAmount(currency string) float64 {
	return 0.01
}

// GetWithdrawFee возвращает комиссию за вывод
func (p *MockProvider) GetWithdrawFee(currency string, amount float64) float64 {
	return 0.001
}

// GetRequiredConfirmations возвращает количество необходимых подтверждений
func (p *MockProvider) GetRequiredConfirmations() int {
	return 1
}
