package ton

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/blockchain"
)

// Config конфигурация TON провайдера
type Config struct {
	// APIEndpoint URL API ноды TON (например, https://toncenter.com/api/v2)
	APIEndpoint string `envconfig:"TON_API_ENDPOINT" default:"https://toncenter.com/api/v2"`
	// APIKey ключ API для доступа к ноде
	APIKey string `envconfig:"TON_API_KEY" default:""`
	// MasterWallet адрес мастер-кошелька для выводов
	MasterWallet string `envconfig:"TON_MASTER_WALLET" default:""`
	// MasterWalletSecret секрет мастер-кошелька (для подписи транзакций)
	MasterWalletSecret string `envconfig:"TON_MASTER_WALLET_SECRET" default:""`
	// MinWithdrawTON минимальная сумма вывода в TON
	MinWithdrawTON float64 `envconfig:"TON_MIN_WITHDRAW" default:"0.1"`
	// WithdrawFeeTON комиссия за вывод в TON
	WithdrawFeeTON float64 `envconfig:"TON_WITHDRAW_FEE" default:"0.01"`
	// RequiredConfirmations количество подтверждений для транзакции
	RequiredConfirmations int `envconfig:"TON_REQUIRED_CONFIRMATIONS" default:"1"`
	// Testnet использовать тестовую сеть
	Testnet bool `envconfig:"TON_TESTNET" default:"false"`
}

// TonProvider реализация BlockchainProvider для TON
type TonProvider struct {
	config           Config
	httpClient       *http.Client
	processedTxCache sync.Map // Кеш обработанных транзакций
}

// NewTonProvider создает новый TON провайдер
func NewTonProvider(config Config) *TonProvider {
	return &TonProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetNetwork возвращает тип сети
func (p *TonProvider) GetNetwork() blockchain.Network {
	return blockchain.NetworkTON
}

// GetSupportedCurrencies возвращает поддерживаемые валюты
func (p *TonProvider) GetSupportedCurrencies() []string {
	return []string{"TON", "USDT"} // TON нативный токен + Jetton USDT
}

// VerifyTransaction проверяет транзакцию в блокчейне TON
func (p *TonProvider) VerifyTransaction(ctx context.Context, txHash string) (*blockchain.TransactionInfo, error) {
	if txHash == "" {
		return nil, fmt.Errorf("transaction hash cannot be empty")
	}

	// Нормализуем хеш
	txHash = strings.TrimSpace(txHash)

	// Делаем запрос к API TON
	url := fmt.Sprintf("%s/getTransactions?hash=%s&limit=1", p.config.APIEndpoint, txHash)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("X-API-Key", p.config.APIKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Парсим ответ
	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			TransactionID struct {
				Hash string `json:"hash"`
			} `json:"transaction_id"`
			InMsg struct {
				Source      string `json:"source"`
				Destination string `json:"destination"`
				Value       string `json:"value"`
			} `json:"in_msg"`
			Utime int64  `json:"utime"`
			Fee   string `json:"fee"`
		} `json:"result"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		if result.Error != "" {
			return nil, fmt.Errorf("API error: %s", result.Error)
		}
		return &blockchain.TransactionInfo{
			Hash:   txHash,
			Status: blockchain.TxStatusNotFound,
		}, nil
	}

	if len(result.Result) == 0 {
		return &blockchain.TransactionInfo{
			Hash:   txHash,
			Status: blockchain.TxStatusNotFound,
		}, nil
	}

	tx := result.Result[0]
	
	// Конвертируем значение из нанотонов в TON
	var amount float64
	fmt.Sscanf(tx.InMsg.Value, "%f", &amount)
	amount = amount / 1e9 // nanoTON -> TON

	var fee float64
	fmt.Sscanf(tx.Fee, "%f", &fee)
	fee = fee / 1e9

	return &blockchain.TransactionInfo{
		Hash:      txHash,
		From:      tx.InMsg.Source,
		To:        tx.InMsg.Destination,
		Amount:    amount,
		Currency:  "TON",
		Status:    blockchain.TxStatusConfirmed,
		Timestamp: tx.Utime,
		Fee:       fee,
		Network:   blockchain.NetworkTON,
	}, nil
}

// IsTransactionProcessed проверяет, была ли транзакция уже обработана
func (p *TonProvider) IsTransactionProcessed(ctx context.Context, txHash string) (bool, error) {
	if txHash == "" {
		return false, fmt.Errorf("transaction hash cannot be empty")
	}

	// Проверяем в кеше
	if _, exists := p.processedTxCache.Load(txHash); exists {
		return true, nil
	}

	// В реальной реализации здесь должен быть запрос к базе данных
	// для проверки, была ли транзакция уже обработана
	return false, nil
}

// MarkTransactionProcessed отмечает транзакцию как обработанную
func (p *TonProvider) MarkTransactionProcessed(txHash string) {
	p.processedTxCache.Store(txHash, true)
}

// GenerateDepositAddress генерирует адрес для депозита
func (p *TonProvider) GenerateDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	// В TON можно использовать memo/comment для идентификации пользователя
	// или генерировать отдельные субкошельки

	// Для простоты используем мастер-кошелек с уникальным memo
	memo := fmt.Sprintf("deposit_%d_%d", userID, time.Now().Unix())

	return &blockchain.DepositAddress{
		Address:  p.config.MasterWallet,
		Memo:     memo,
		Network:  blockchain.NetworkTON,
		Currency: currency,
	}, nil
}

// GetDepositAddress возвращает существующий адрес депозита
func (p *TonProvider) GetDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	// В реальной реализации здесь должен быть запрос к базе данных
	// для получения ранее сгенерированного адреса

	// Пока возвращаем мастер-кошелек
	return &blockchain.DepositAddress{
		Address:  p.config.MasterWallet,
		Memo:     fmt.Sprintf("user_%d", userID),
		Network:  blockchain.NetworkTON,
		Currency: currency,
	}, nil
}

// ProcessWithdraw инициирует вывод средств
func (p *TonProvider) ProcessWithdraw(ctx context.Context, request *blockchain.WithdrawRequest) (*blockchain.WithdrawResult, error) {
	if request == nil {
		return nil, fmt.Errorf("withdraw request cannot be nil")
	}

	// Валидация
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

	// Проверяем минимальную сумму
	minAmount := p.GetMinWithdrawAmount(request.Currency)
	if request.Amount < minAmount {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("minimum withdraw amount is %.4f %s", minAmount, request.Currency),
		}, nil
	}

	// Валидируем адрес
	valid, err := p.ValidateAddress(ctx, request.ToAddress)
	if err != nil {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to validate address: %v", err),
		}, nil
	}
	if !valid {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: "invalid destination address",
		}, nil
	}

	// Рассчитываем комиссию
	fee := p.GetWithdrawFee(request.Currency, request.Amount)

	// TODO: Реальная отправка транзакции через TON SDK
	// В production здесь должен быть вызов TON SDK для подписи и отправки транзакции
	
	// Пока возвращаем заглушку для разработки
	mockTxHash := fmt.Sprintf("mock_tx_%s_%d", request.TransactionID, time.Now().UnixNano())

	return &blockchain.WithdrawResult{
		Success:       true,
		TransactionID: request.TransactionID,
		TxHash:        mockTxHash,
		Fee:           fee,
	}, nil
}

// GetTransactionStatus получает статус транзакции
func (p *TonProvider) GetTransactionStatus(ctx context.Context, txHash string) (blockchain.TransactionStatus, error) {
	txInfo, err := p.VerifyTransaction(ctx, txHash)
	if err != nil {
		return blockchain.TxStatusPending, err
	}
	return txInfo.Status, nil
}

// ValidateAddress проверяет валидность адреса TON
func (p *TonProvider) ValidateAddress(ctx context.Context, address string) (bool, error) {
	if address == "" {
		return false, nil
	}

	// Базовая проверка формата адреса TON
	// TON адреса обычно начинаются с EQ, UQ или 0:
	address = strings.TrimSpace(address)
	
	// Raw format: 0:...
	if strings.HasPrefix(address, "0:") || strings.HasPrefix(address, "-1:") {
		// Проверяем длину (workchain:address)
		parts := strings.Split(address, ":")
		if len(parts) == 2 && len(parts[1]) == 64 {
			return true, nil
		}
	}

	// User-friendly format: EQ... или UQ...
	if strings.HasPrefix(address, "EQ") || strings.HasPrefix(address, "UQ") {
		// Проверяем длину (48 символов для user-friendly формата)
		if len(address) == 48 {
			return true, nil
		}
	}

	// Также может быть формат kQ... для testnet
	if p.config.Testnet && (strings.HasPrefix(address, "kQ") || strings.HasPrefix(address, "0Q")) {
		if len(address) == 48 {
			return true, nil
		}
	}

	return false, nil
}

// GetBalance получает баланс адреса
func (p *TonProvider) GetBalance(ctx context.Context, address string, currency string) (float64, error) {
	if address == "" {
		return 0, fmt.Errorf("address cannot be empty")
	}

	url := fmt.Sprintf("%s/getAddressBalance?address=%s", p.config.APIEndpoint, address)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	if p.config.APIKey != "" {
		req.Header.Set("X-API-Key", p.config.APIKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch balance: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool   `json:"ok"`
		Result string `json:"result"`
		Error  string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return 0, fmt.Errorf("API error: %s", result.Error)
	}

	var balance float64
	fmt.Sscanf(result.Result, "%f", &balance)
	balance = balance / 1e9 // nanoTON -> TON

	return balance, nil
}

// GetMinWithdrawAmount возвращает минимальную сумму вывода
func (p *TonProvider) GetMinWithdrawAmount(currency string) float64 {
	switch currency {
	case "TON":
		return p.config.MinWithdrawTON
	case "USDT":
		return 1.0 // Минимум 1 USDT
	default:
		return 0.1
	}
}

// GetWithdrawFee возвращает комиссию за вывод
func (p *TonProvider) GetWithdrawFee(currency string, amount float64) float64 {
	switch currency {
	case "TON":
		return p.config.WithdrawFeeTON
	case "USDT":
		return 0.5 // 0.5 USDT комиссия
	default:
		return 0.01
	}
}

// GetRequiredConfirmations возвращает количество необходимых подтверждений
func (p *TonProvider) GetRequiredConfirmations() int {
	return p.config.RequiredConfirmations
}
