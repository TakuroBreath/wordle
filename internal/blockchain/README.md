# Blockchain Provider Interface

Этот пакет предоставляет абстракцию для работы с различными блокчейнами в проекте Wordle.

## Архитектура

```
blockchain/
├── provider.go         # Основной интерфейс BlockchainProvider
├── mock/
│   └── provider.go     # Mock-реализация для тестирования
└── ton/
    └── provider.go     # Реализация для TON блокчейна
```

## Интерфейс BlockchainProvider

Интерфейс определяет следующие методы для работы с блокчейном:

```go
type BlockchainProvider interface {
    // Получить тип сети (TON, ETH, BSC и т.д.)
    GetNetwork() Network
    
    // Получить список поддерживаемых валют
    GetSupportedCurrencies() []string
    
    // Проверить транзакцию в блокчейне
    VerifyTransaction(ctx context.Context, txHash string) (*TransactionInfo, error)
    
    // Проверить, была ли транзакция уже обработана
    IsTransactionProcessed(ctx context.Context, txHash string) (bool, error)
    
    // Сгенерировать адрес для депозита
    GenerateDepositAddress(ctx context.Context, userID uint64, currency string) (*DepositAddress, error)
    
    // Получить существующий адрес депозита
    GetDepositAddress(ctx context.Context, userID uint64, currency string) (*DepositAddress, error)
    
    // Инициировать вывод средств
    ProcessWithdraw(ctx context.Context, request *WithdrawRequest) (*WithdrawResult, error)
    
    // Получить статус транзакции
    GetTransactionStatus(ctx context.Context, txHash string) (TransactionStatus, error)
    
    // Проверить валидность адреса
    ValidateAddress(ctx context.Context, address string) (bool, error)
    
    // Получить баланс адреса
    GetBalance(ctx context.Context, address string, currency string) (float64, error)
    
    // Получить минимальную сумму вывода
    GetMinWithdrawAmount(currency string) float64
    
    // Получить комиссию за вывод
    GetWithdrawFee(currency string, amount float64) float64
    
    // Получить количество необходимых подтверждений
    GetRequiredConfirmations() int
}
```

## Добавление нового блокчейна

Для добавления поддержки нового блокчейна (например, Ethereum):

### 1. Создайте новый пакет

```bash
mkdir -p internal/blockchain/ethereum
```

### 2. Реализуйте интерфейс BlockchainProvider

```go
// internal/blockchain/ethereum/provider.go
package ethereum

import (
    "context"
    "github.com/TakuroBreath/wordle/internal/blockchain"
)

type Config struct {
    RPCURL       string
    ChainID      int64
    MasterWallet string
    PrivateKey   string
}

type EthereumProvider struct {
    config Config
    // ... другие поля
}

func NewEthereumProvider(config Config) *EthereumProvider {
    return &EthereumProvider{
        config: config,
    }
}

func (p *EthereumProvider) GetNetwork() blockchain.Network {
    return blockchain.NetworkEthereum
}

func (p *EthereumProvider) GetSupportedCurrencies() []string {
    return []string{"ETH", "USDT", "USDC"}
}

// ... реализуйте остальные методы интерфейса
```

### 3. Добавьте конфигурацию в config.go

```go
// internal/config/config.go
type BlockchainConfig struct {
    Provider string `envconfig:"BLOCKCHAIN_PROVIDER" default:"mock"`
    TON      TONConfig
    Ethereum EthereumConfig  // Добавьте новую конфигурацию
}

type EthereumConfig struct {
    RPCURL       string `envconfig:"ETH_RPC_URL" default:""`
    ChainID      int64  `envconfig:"ETH_CHAIN_ID" default:"1"`
    MasterWallet string `envconfig:"ETH_MASTER_WALLET" default:""`
}
```

### 4. Зарегистрируйте провайдер в service.go

```go
// internal/service/service.go
func createBlockchainProvider(cfg config.BlockchainConfig) blockchain.BlockchainProvider {
    switch cfg.Provider {
    case "ton":
        return ton.NewTonProvider(ton.Config{...})
    case "ethereum":
        return ethereum.NewEthereumProvider(ethereum.Config{
            RPCURL:       cfg.Ethereum.RPCURL,
            ChainID:      cfg.Ethereum.ChainID,
            MasterWallet: cfg.Ethereum.MasterWallet,
        })
    case "mock":
        fallthrough
    default:
        return mock.NewMockProvider(blockchain.NetworkTON, []string{"TON", "USDT"})
    }
}
```

## Конфигурация

### Переменные окружения

```bash
# Тип провайдера: "ton", "ethereum", "mock"
BLOCKCHAIN_PROVIDER=ton

# Конфигурация TON
TON_API_ENDPOINT=https://toncenter.com/api/v2
TON_API_KEY=your_api_key
TON_MASTER_WALLET=EQ...
TON_MIN_WITHDRAW=0.1
TON_WITHDRAW_FEE=0.01
TON_REQUIRED_CONFIRMATIONS=1
TON_TESTNET=false

# Конфигурация Ethereum (пример)
ETH_RPC_URL=https://mainnet.infura.io/v3/your_key
ETH_CHAIN_ID=1
ETH_MASTER_WALLET=0x...
```

## MultiChainProvider

Для поддержки нескольких блокчейнов одновременно используйте `MultiChainProvider`:

```go
multiProvider := blockchain.NewMultiChainProvider(blockchain.NetworkTON)
multiProvider.RegisterProvider(blockchain.NetworkTON, tonProvider)
multiProvider.RegisterProvider(blockchain.NetworkEthereum, ethProvider)

// Получить провайдер для конкретной сети
provider, ok := multiProvider.GetProvider(blockchain.NetworkTON)
if ok {
    // Использовать провайдер
}

// Получить все поддерживаемые сети
networks := multiProvider.GetSupportedNetworks()
```

## Тестирование

Используйте MockProvider для тестирования:

```go
import "github.com/TakuroBreath/wordle/internal/blockchain/mock"

provider := mock.NewMockProvider(blockchain.NetworkTON, []string{"TON", "USDT"})

// MockProvider всегда возвращает успешные результаты
txInfo, _ := provider.VerifyTransaction(ctx, "any_hash")
// txInfo.Status == blockchain.TxStatusConfirmed
```
