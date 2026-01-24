package base

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/blockchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	defaultUSDCDecimals = 6
)

var (
	transferEventSig = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	erc20ABI         = mustParseABI()
)

// Config конфигурация Base провайдера
type Config struct {
	RPCURL                string  `envconfig:"BASE_RPC_URL" default:""`
	ChainID               int64   `envconfig:"BASE_CHAIN_ID" default:"8453"`
	MasterWallet          string  `envconfig:"BASE_MASTER_WALLET" default:""`
	PrivateKey            string  `envconfig:"BASE_PRIVATE_KEY" default:""`
	USDCContractAddress   string  `envconfig:"BASE_USDC_CONTRACT" default:""`
	MinWithdrawUSDC       float64 `envconfig:"BASE_MIN_WITHDRAW" default:"1"`
	WithdrawFeeUSDC       float64 `envconfig:"BASE_WITHDRAW_FEE" default:"0.5"`
	RequiredConfirmations int     `envconfig:"BASE_REQUIRED_CONFIRMATIONS" default:"3"`
	TokenDecimals         int     `envconfig:"BASE_TOKEN_DECIMALS" default:"6"`
}

// Provider реализация BlockchainProvider для Base (USDC)
type Provider struct {
	config           Config
	client           *ethclient.Client
	processedTxCache sync.Map
	mu               sync.RWMutex
}

// NewProvider создает новый Base провайдер
func NewProvider(config Config) *Provider {
	return &Provider{
		config: config,
	}
}

// GetNetwork возвращает тип сети
func (p *Provider) GetNetwork() blockchain.Network {
	return blockchain.NetworkBase
}

// GetSupportedCurrencies возвращает поддерживаемые валюты
func (p *Provider) GetSupportedCurrencies() []string {
	return []string{"USDC"}
}

// VerifyTransaction проверяет транзакцию в блокчейне Base
func (p *Provider) VerifyTransaction(ctx context.Context, txHash string) (*blockchain.TransactionInfo, error) {
	if txHash == "" {
		return nil, fmt.Errorf("transaction hash cannot be empty")
	}

	client, err := p.getClient(ctx)
	if err != nil {
		return nil, err
	}

	hash := common.HexToHash(txHash)
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		if err == ethereum.NotFound {
			return &blockchain.TransactionInfo{
				Hash:   txHash,
				Status: blockchain.TxStatusNotFound,
			}, nil
		}
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt == nil {
		return &blockchain.TransactionInfo{
			Hash:   txHash,
			Status: blockchain.TxStatusPending,
		}, nil
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return &blockchain.TransactionInfo{
			Hash:   txHash,
			Status: blockchain.TxStatusFailed,
		}, nil
	}

	masterAddr, err := p.parseAddress(p.config.MasterWallet)
	if err != nil {
		return nil, fmt.Errorf("invalid master wallet address: %w", err)
	}
	usdcAddr, err := p.parseAddress(p.config.USDCContractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid USDC contract address: %w", err)
	}

	var (
		foundTransfer bool
		fromAddress   string
		toAddress     string
		amount        float64
	)

	for _, logEntry := range receipt.Logs {
		if logEntry.Address != usdcAddr {
			continue
		}
		if len(logEntry.Topics) < 3 || logEntry.Topics[0] != transferEventSig {
			continue
		}

		from := common.HexToAddress(logEntry.Topics[1].Hex())
		to := common.HexToAddress(logEntry.Topics[2].Hex())
		if to != masterAddr && from != masterAddr {
			continue
		}

		rawAmount := new(big.Int).SetBytes(logEntry.Data)
		amount = p.fromBaseUnits(rawAmount)
		fromAddress = from.Hex()
		toAddress = to.Hex()
		foundTransfer = true
		break
	}

	if !foundTransfer {
		return &blockchain.TransactionInfo{
			Hash:   txHash,
			Status: blockchain.TxStatusFailed,
		}, nil
	}

	blockNumber := uint64(0)
	confirmations := uint64(0)
	if receipt.BlockNumber != nil {
		blockNumber = receipt.BlockNumber.Uint64()
		if latest, err := client.BlockNumber(ctx); err == nil && latest >= blockNumber {
			confirmations = latest - blockNumber + 1
		}
	}

	status := blockchain.TxStatusConfirmed
	if confirmations > 0 && int(confirmations) < p.config.RequiredConfirmations {
		status = blockchain.TxStatusPending
	}

	timestamp := time.Now().Unix()
	if receipt.BlockNumber != nil {
		if header, err := client.HeaderByNumber(ctx, receipt.BlockNumber); err == nil && header != nil {
			timestamp = int64(header.Time)
		}
	}

	return &blockchain.TransactionInfo{
		Hash:        txHash,
		From:        fromAddress,
		To:          toAddress,
		Amount:      amount,
		Currency:    "USDC",
		Status:      status,
		BlockNumber: blockNumber,
		Timestamp:   timestamp,
		Fee:         0,
		Network:     blockchain.NetworkBase,
	}, nil
}

// IsTransactionProcessed проверяет, была ли транзакция уже обработана
func (p *Provider) IsTransactionProcessed(ctx context.Context, txHash string) (bool, error) {
	if txHash == "" {
		return false, fmt.Errorf("transaction hash cannot be empty")
	}
	_, exists := p.processedTxCache.Load(txHash)
	return exists, nil
}

// MarkTransactionProcessed отмечает транзакцию как обработанную
func (p *Provider) MarkTransactionProcessed(txHash string) {
	p.processedTxCache.Store(txHash, true)
}

// GenerateDepositAddress генерирует адрес для депозита
func (p *Provider) GenerateDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	return &blockchain.DepositAddress{
		Address:  p.config.MasterWallet,
		Network:  blockchain.NetworkBase,
		Currency: currency,
	}, nil
}

// GetDepositAddress возвращает существующий адрес депозита
func (p *Provider) GetDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	return p.GenerateDepositAddress(ctx, userID, currency)
}

// ProcessWithdraw инициирует вывод средств
func (p *Provider) ProcessWithdraw(ctx context.Context, request *blockchain.WithdrawRequest) (*blockchain.WithdrawResult, error) {
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

	valid, err := p.ValidateAddress(ctx, request.ToAddress)
	if err != nil || !valid {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: "invalid destination address",
		}, nil
	}

	minAmount := p.GetMinWithdrawAmount(request.Currency)
	if request.Amount < minAmount {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("minimum withdraw amount is %.6f %s", minAmount, request.Currency),
		}, nil
	}

	client, err := p.getClient(ctx)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(p.config.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	masterAddr, err := p.parseAddress(p.config.MasterWallet)
	if err != nil {
		return nil, fmt.Errorf("invalid master wallet address: %w", err)
	}
	if masterAddr != fromAddress {
		return nil, fmt.Errorf("master wallet does not match private key address")
	}

	toAddress := common.HexToAddress(request.ToAddress)
	usdcAddr, err := p.parseAddress(p.config.USDCContractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid USDC contract address: %w", err)
	}

	amountUnits := p.toBaseUnits(request.Amount)
	data, err := erc20ABI.Pack("transfer", toAddress, amountUnits)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transfer data: %w", err)
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	callMsg := ethereum.CallMsg{
		From:  fromAddress,
		To:    &usdcAddr,
		Value: big.NewInt(0),
		Data:  data,
	}
	gasLimit, err := client.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	tx := types.NewTransaction(nonce, usdcAddr, big.NewInt(0), gasLimit, gasPrice, data)
	signer := types.NewEIP155Signer(big.NewInt(p.config.ChainID))
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return &blockchain.WithdrawResult{
		Success:       true,
		TransactionID: request.TransactionID,
		TxHash:        signedTx.Hash().Hex(),
		Fee:           p.GetWithdrawFee(request.Currency, request.Amount),
	}, nil
}

// GetTransactionStatus получает статус транзакции
func (p *Provider) GetTransactionStatus(ctx context.Context, txHash string) (blockchain.TransactionStatus, error) {
	txInfo, err := p.VerifyTransaction(ctx, txHash)
	if err != nil {
		return blockchain.TxStatusPending, err
	}
	return txInfo.Status, nil
}

// ValidateAddress проверяет валидность адреса
func (p *Provider) ValidateAddress(ctx context.Context, address string) (bool, error) {
	if address == "" {
		return false, nil
	}
	return common.IsHexAddress(address), nil
}

// GetBalance получает баланс адреса
func (p *Provider) GetBalance(ctx context.Context, address string, currency string) (float64, error) {
	if address == "" {
		return 0, fmt.Errorf("address cannot be empty")
	}

	client, err := p.getClient(ctx)
	if err != nil {
		return 0, err
	}

	usdcAddr, err := p.parseAddress(p.config.USDCContractAddress)
	if err != nil {
		return 0, fmt.Errorf("invalid USDC contract address: %w", err)
	}
	account := common.HexToAddress(address)
	data, err := erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return 0, fmt.Errorf("failed to pack balanceOf data: %w", err)
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &usdcAddr,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	balance := new(big.Int).SetBytes(result)
	return p.fromBaseUnits(balance), nil
}

// GetMinWithdrawAmount возвращает минимальную сумму вывода
func (p *Provider) GetMinWithdrawAmount(currency string) float64 {
	if currency == "USDC" {
		return p.config.MinWithdrawUSDC
	}
	return 0
}

// GetWithdrawFee возвращает комиссию за вывод
func (p *Provider) GetWithdrawFee(currency string, amount float64) float64 {
	if currency == "USDC" {
		return p.config.WithdrawFeeUSDC
	}
	return 0
}

// GetRequiredConfirmations возвращает количество необходимых подтверждений
func (p *Provider) GetRequiredConfirmations() int {
	return p.config.RequiredConfirmations
}

func (p *Provider) getClient(ctx context.Context) (*ethclient.Client, error) {
	p.mu.RLock()
	if p.client != nil {
		client := p.client
		p.mu.RUnlock()
		return client, nil
	}
	p.mu.RUnlock()

	if p.config.RPCURL == "" {
		return nil, fmt.Errorf("RPC URL is not configured")
	}

	client, err := ethclient.DialContext(ctx, p.config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	p.mu.Lock()
	p.client = client
	p.mu.Unlock()

	return client, nil
}

func (p *Provider) parseAddress(address string) (common.Address, error) {
	if !common.IsHexAddress(address) {
		return common.Address{}, fmt.Errorf("invalid address")
	}
	return common.HexToAddress(address), nil
}

func (p *Provider) decimals() int {
	if p.config.TokenDecimals > 0 {
		return p.config.TokenDecimals
	}
	return defaultUSDCDecimals
}

func (p *Provider) fromBaseUnits(amount *big.Int) float64 {
	if amount == nil {
		return 0
	}
	decimals := p.decimals()
	if decimals == 0 {
		value, _ := new(big.Float).SetInt(amount).Float64()
		return value
	}
	divisor := new(big.Float).SetFloat64(math.Pow10(decimals))
	value := new(big.Float).SetInt(amount)
	value.Quo(value, divisor)
	result, _ := value.Float64()
	return result
}

func (p *Provider) toBaseUnits(amount float64) *big.Int {
	decimals := p.decimals()
	multiplier := math.Pow10(decimals)
	value := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(multiplier))
	result := new(big.Int)
	value.Int(result)
	return result
}

func mustParseABI() abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(`[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"},{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`))
	if err != nil {
		panic(err)
	}
	return parsed
}
