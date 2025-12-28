package ton

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/blockchain"
	"github.com/takurobreath/wordle/internal/logger"
	"go.uber.org/zap"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

// Config конфигурация TON провайдера
type Config struct {
	// APIEndpoint URL для конфига (global.config.json)
	APIEndpoint string `envconfig:"TON_CONFIG_URL" default:"https://ton.org/global.config.json"`
	// APIKey ключ API (не используется для liteclient, но оставим для совместимости)
	APIKey string `envconfig:"TON_API_KEY" default:""`
	// MasterWallet адрес мастер-кошелька (отображаемый)
	MasterWallet string `envconfig:"TON_MASTER_WALLET" default:""`
	// MasterWalletSecret сид-фраза мастер-кошелька (24 слова)
	MasterWalletSecret string `envconfig:"TON_MASTER_WALLET_SECRET" default:""`
	// MinWithdrawTON минимальная сумма вывода в TON
	MinWithdrawTON float64 `envconfig:"TON_MIN_WITHDRAW" default:"0.1"`
	// WithdrawFeeTON комиссия за вывод в TON
	WithdrawFeeTON float64 `envconfig:"TON_WITHDRAW_FEE" default:"0.01"`
	// RequiredConfirmations количество подтверждений (для TON обычно 1 достаточно, если блок финализирован)
	RequiredConfirmations int `envconfig:"TON_REQUIRED_CONFIRMATIONS" default:"1"`
	// Testnet использовать тестовую сеть
	Testnet bool `envconfig:"TON_TESTNET" default:"false"`
}

// TonProvider реализация BlockchainProvider для TON
type TonProvider struct {
	config           Config
	client           *ton.APIClient
	wallet           *wallet.Wallet
	processedTxCache sync.Map
	mx               sync.RWMutex
}

// NewTonProvider создает новый TON провайдер
func NewTonProvider(ctx context.Context, config Config) (*TonProvider, error) {
	if config.Testnet {
		config.APIEndpoint = "https://ton-blockchain.github.io/testnet-global.config.json"
	}

	client := liteclient.NewConnectionPool()

	// Скачиваем конфиг
	err := client.AddConnectionsFromConfigUrl(ctx, config.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to load ton config: %w", err)
	}

	api := ton.NewAPIClient(client)

	provider := &TonProvider{
		config: config,
		client: api,
	}

	// Инициализируем кошелек, если есть секрет
	if config.MasterWalletSecret != "" {
		words := strings.Split(strings.TrimSpace(config.MasterWalletSecret), " ")
		w, err := wallet.FromSeed(api, words, wallet.V4R2)
		if err != nil {
			logger.Log.Warn("Failed to initialize wallet from seed", zap.Error(err))
		} else {
			provider.wallet = w
			logger.Log.Info("Initialized custodial wallet", zap.String("address", w.Address().String()))
			// Обновляем адрес в конфиге на реальный из сида, если он отличается (или если не задан)
			provider.config.MasterWallet = w.Address().String()
		}
	}

	return provider, nil
}

// GetNetwork возвращает тип сети
func (p *TonProvider) GetNetwork() blockchain.Network {
	return blockchain.NetworkTON
}

// GetSupportedCurrencies возвращает поддерживаемые валюты
func (p *TonProvider) GetSupportedCurrencies() []string {
	return []string{"TON", "USDT"}
}

// VerifyTransaction проверяет транзакцию по хешу
// В TON utils мы можем получить транзакцию, если знаем блок, или через liteserver,
// но поиск произвольной транзакции по хешу в liteserver сложен без архивной ноды.
// Обычно мы сканируем свои транзакции.
// Если нужно проверить произвольный хеш, лучше использовать indexer API (toncenter/tonapi).
// Здесь мы попробуем использовать API если liteserver не найдет, или просто заглушку если это "чужая" транзакция.
// НО: Поскольку мы custodial, мы обычно проверяем ВХОДЯЩИЕ к НАМ.
func (p *TonProvider) VerifyTransaction(ctx context.Context, txHash string) (*blockchain.TransactionInfo, error) {
	// Для LiteClient поиск транзакции по хешу сложен (нужно знать account).
	// Если мы ищем транзакцию, отправленную НАМ, мы можем поискать в истории нашего кошелька.
	
	// Вариант 1: Ищем в последних транзакциях кошелька
	if p.wallet != nil {
		txs, err := p.wallet.ListTransactions(ctx, 50) // последние 50
		if err == nil {
			for _, tx := range txs {
				if fmt.Sprintf("%x", tx.Hash) == txHash || base64ToHex(txHash) == fmt.Sprintf("%x", tx.Hash) {
					return p.mapTransaction(tx), nil
				}
			}
		}
	}

	// Если не нашли, и есть API Key/Endpoint для HTTP API (fallback), можно было бы использовать его.
	// Но пока вернем NotFound
	return nil, fmt.Errorf("transaction not found in recent history or not supported via liteclient lookup")
}

func base64ToHex(b64 string) string {
	// TODO: implement specific TON hash conversion if needed, 
	// usually hashes are hex or base64.
	return b64 
}

// mapTransaction конвертирует транзакцию TON в нашу структуру
func (p *TonProvider) mapTransaction(tx *tlb.Transaction) *blockchain.TransactionInfo {
	if tx.IO.In == nil {
		return nil
	}

	inMsg, ok := tx.IO.In.Msg.Payload.(*tlb.InternalMessage)
	if !ok {
		// External message or unsupported
		return nil
	}
	
	amt := inMsg.Amount.Nano()
	amount := float64(amt.Uint64()) / 1e9

	// Get Sender
	sender := inMsg.SrcAddr.String()
	
	// Get Comment
	var comment string
	if inMsg.Body != nil {
		slice := inMsg.Body.BeginParse()
		// Try to parse comment (text)
		// Standard text comment opcode is 0, but often it's just raw text in body if op is 0
		// Or 0x00000000 prefix for comment.
		op, err := slice.LoadUInt(32)
		if err == nil && op == 0 {
			txt, err := slice.LoadStringSnake()
			if err == nil {
				comment = txt
			}
		}
	}

	return &blockchain.TransactionInfo{
		Hash:      fmt.Sprintf("%x", tx.Hash),
		From:      sender,
		To:        p.config.MasterWallet,
		Amount:    amount,
		Currency:  "TON", // TODO: Detect Jetton transfers for USDT
		Status:    blockchain.TxStatusConfirmed,
		Timestamp: int64(tx.Now),
		Network:   blockchain.NetworkTON,
		// We can put comment in a field if we extend TransactionInfo, 
		// but interface doesn't have it. We rely on the caller or casting.
	}
}


// IsTransactionProcessed checks cache
func (p *TonProvider) IsTransactionProcessed(ctx context.Context, txHash string) (bool, error) {
	if _, exists := p.processedTxCache.Load(txHash); exists {
		return true, nil
	}
	return false, nil
}

func (p *TonProvider) MarkTransactionProcessed(txHash string) {
	p.processedTxCache.Store(txHash, true)
}

// GenerateDepositAddress returns master wallet + memo
func (p *TonProvider) GenerateDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	// Memo format: "user:<userID>" or just "<userID>"?
	// Prompt says: "Comment in transaction will be game identifier" for creation.
	// For joining/deposit: "Need to think...". I decided on "user_id" or "ref:user_id".
	
	memo := fmt.Sprintf("user_%d", userID)
	return &blockchain.DepositAddress{
		Address:  p.config.MasterWallet,
		Memo:     memo,
		Network:  blockchain.NetworkTON,
		Currency: currency,
	}, nil
}

func (p *TonProvider) GetDepositAddress(ctx context.Context, userID uint64, currency string) (*blockchain.DepositAddress, error) {
	return p.GenerateDepositAddress(ctx, userID, currency)
}

// ProcessWithdraw sends TON
func (p *TonProvider) ProcessWithdraw(ctx context.Context, request *blockchain.WithdrawRequest) (*blockchain.WithdrawResult, error) {
	if p.wallet == nil {
		return nil, fmt.Errorf("wallet not initialized (no secret provided)")
	}

	addr, err := address.ParseAddr(request.ToAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	amountNano := tlb.MustFromTON(fmt.Sprintf("%f", request.Amount))
	
	var body *cell.Cell
	if request.Memo != "" {
		body, err = wallet.CreateCommentCell(request.Memo)
		if err != nil {
			return nil, fmt.Errorf("failed to create comment: %w", err)
		}
	}

	// Send transfer
	// wait for confirmation? usually Transfer returns msg info, checking blocks is separate.
	// xssnick/tonutils-go Transfer sends message to outbox.
	
	logger.Log.Info("Sending withdrawal", zap.Float64("amount", request.Amount), zap.String("to", request.ToAddress))
	
	// We use Default wait behavior or specific?
	// Transfer returns error if failed to send to LS.
	err = p.wallet.Transfer(ctx, addr, amountNano, body, true) // true = bounce
	if err != nil {
		return &blockchain.WithdrawResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Generating a fake hash or trying to calculate msg hash?
	// Calculating msg hash is possible but complex without response parsing.
	// We'll return success and pending.
	
	return &blockchain.WithdrawResult{
		Success:       true,
		TransactionID: request.TransactionID,
		TxHash:        "pending", // We don't get the hash immediately from Transfer method in this lib version easily without subscribing
		Fee:           0.005, // Estimate
	}, nil
}

func (p *TonProvider) GetTransactionStatus(ctx context.Context, txHash string) (blockchain.TransactionStatus, error) {
	// Implement if needed
	return blockchain.TxStatusConfirmed, nil
}

func (p *TonProvider) ValidateAddress(ctx context.Context, addr string) (bool, error) {
	_, err := address.ParseAddr(addr)
	return err == nil, nil
}

func (p *TonProvider) GetBalance(ctx context.Context, addrStr string, currency string) (float64, error) {
	block, err := p.client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, err
	}

	addr, err := address.ParseAddr(addrStr)
	if err != nil {
		return 0, err
	}

	acc, err := p.client.GetAccount(ctx, block, addr)
	if err != nil {
		return 0, err
	}

	if !acc.IsActive {
		return 0, nil
	}

	return float64(acc.State.Balance.Nano().Uint64()) / 1e9, nil
}

func (p *TonProvider) GetMinWithdrawAmount(currency string) float64 {
	return p.config.MinWithdrawTON
}

func (p *TonProvider) GetWithdrawFee(currency string, amount float64) float64 {
	return p.config.WithdrawFeeTON
}

func (p *TonProvider) GetRequiredConfirmations() int {
	return 1
}

// Extra methods for Worker

// GetRecentTransactions fetches transactions for the wallet
func (p *TonProvider) GetRecentTransactions(ctx context.Context, limit int) ([]blockchain.TransactionInfo, error) {
	if p.wallet == nil {
		// Try to inspect address without wallet struct?
		// We can use p.client.ListTransactions if we have the address.
		// But wallet.ListTransactions is convenient.
		// If wallet is nil (no secret), we can construct a wallet instance just for reading if we have address?
		// Yes, wallet.New(api, addr)
		
		if p.config.MasterWallet == "" {
			return nil, fmt.Errorf("master wallet address not configured")
		}

		addr, err := address.ParseAddr(p.config.MasterWallet)
		if err != nil {
			return nil, fmt.Errorf("master wallet address invalid: %w", err)
		}
		// Create a read-only wallet wrapper
		w, err := wallet.FromAddr(p.client, addr, wallet.V4R2)
		if err != nil {
			return nil, err
		}
		p.wallet = w
	}

	txs, err := p.wallet.ListTransactions(ctx, uint32(limit))
	if err != nil {
		return nil, err
	}

	var results []blockchain.TransactionInfo
	for _, tx := range txs {
		if tx.IO.In == nil {
			continue // We care about incoming
		}
		
		inMsg, ok := tx.IO.In.Msg.Payload.(*tlb.InternalMessage)
		if !ok {
			continue
		}

		// Check if it's a bounce (we shouldn't process bounces as deposits usually)
		if inMsg.Bounced {
			continue
		}

		amt := inMsg.Amount.Nano()
		amount := float64(amt.Uint64()) / 1e9
		sender := inMsg.SrcAddr.String()

		var comment string
		if inMsg.Body != nil {
			// Try parsing comment
			slice := inMsg.Body.BeginParse()
			op, err := slice.LoadUInt(32)
			if err == nil && op == 0 {
				txt, err := slice.LoadStringSnake()
				if err == nil {
					comment = txt
				}
			}
		}

		results = append(results, blockchain.TransactionInfo{
			Hash:      fmt.Sprintf("%x", tx.Hash),
			From:      sender,
			To:        p.config.MasterWallet,
			Amount:    amount,
			Currency:  "TON", // Default to TON
			Status:    blockchain.TxStatusConfirmed,
			Timestamp: int64(tx.Now),
			Comment:   comment,
			Network:   blockchain.NetworkTON,
		})
	}
	return results, nil
}
