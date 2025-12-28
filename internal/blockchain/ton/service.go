package ton

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"go.uber.org/zap"
)

// ServiceConfig конфигурация TON сервиса
type ServiceConfig struct {
	// API настройки
	APIEndpoint string // URL для toncenter v3 API
	APIKey      string // API ключ

	// Кастодиальный кошелек
	MasterWalletAddress string // Адрес мастер-кошелька
	MasterWalletSeed    string // Seed фраза мастер-кошелька (24 слова)

	// Настройки сети
	Testnet               bool    // Использовать testnet
	MinWithdrawTON        float64 // Минимальная сумма вывода TON
	MinWithdrawUSDT       float64 // Минимальная сумма вывода USDT
	WithdrawFeeTON        float64 // Комиссия за вывод TON
	RequiredConfirmations int     // Количество подтверждений

	// USDT Jetton
	USDTMasterAddress string // Адрес мастер-контракта USDT Jetton
}

// Service реализация TON сервиса
type Service struct {
	config       ServiceConfig
	client       ton.APIClientWrapped
	masterWallet *wallet.Wallet
	masterAddr   *address.Address
	usdtMaster   *address.Address
	mu           sync.RWMutex
	logger       *zap.Logger
}

// NewService создаёт новый TON сервис
func NewService(config ServiceConfig) (*Service, error) {
	log := logger.GetLogger(zap.String("service", "ton"))

	s := &Service{
		config: config,
		logger: log,
	}

	// Парсим адрес мастер-кошелька
	if config.MasterWalletAddress != "" {
		addr, err := address.ParseAddr(config.MasterWalletAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid master wallet address: %w", err)
		}
		s.masterAddr = addr
	}

	// Парсим адрес USDT Jetton
	if config.USDTMasterAddress != "" {
		addr, err := address.ParseAddr(config.USDTMasterAddress)
		if err != nil {
			log.Warn("Invalid USDT master address", zap.Error(err))
		} else {
			s.usdtMaster = addr
		}
	}

	return s, nil
}

// Connect подключается к TON сети
func (s *Service) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Connecting to TON network",
		zap.Bool("testnet", s.config.Testnet),
		zap.String("api_endpoint", s.config.APIEndpoint))

	// Создаём connection pool
	pool := liteclient.NewConnectionPool()

	// Используем конфиг для testnet или mainnet
	var configURL string
	if s.config.Testnet {
		configURL = "https://ton.org/testnet-global.config.json"
	} else {
		configURL = "https://ton.org/global.config.json"
	}

	err := pool.AddConnectionsFromConfigUrl(ctx, configURL)
	if err != nil {
		return fmt.Errorf("failed to add connections: %w", err)
	}

	// Создаём API клиент
	client := ton.NewAPIClient(pool, ton.ProofCheckPolicyFast).WithRetry()
	s.client = client

	// Инициализируем кошелёк если есть seed
	if s.config.MasterWalletSeed != "" {
		words := strings.Split(s.config.MasterWalletSeed, " ")
		if len(words) != 24 {
			return fmt.Errorf("invalid seed phrase: expected 24 words, got %d", len(words))
		}

		w, err := wallet.FromSeed(client, words, wallet.V4R2)
		if err != nil {
			return fmt.Errorf("failed to create wallet from seed: %w", err)
		}

		s.masterWallet = w
		s.masterAddr = w.WalletAddress()

		s.logger.Info("Master wallet initialized",
			zap.String("address", s.masterAddr.String()))
	}

	s.logger.Info("Connected to TON network successfully")
	return nil
}

// GetMasterWalletAddress возвращает адрес мастер-кошелька
func (s *Service) GetMasterWalletAddress() string {
	if s.masterAddr == nil {
		return s.config.MasterWalletAddress
	}
	return s.masterAddr.String()
}

// ValidateAddress проверяет валидность TON адреса
func (s *Service) ValidateAddress(addr string) bool {
	if addr == "" {
		return false
	}

	_, err := address.ParseAddr(addr)
	return err == nil
}

// GetBalance получает баланс адреса в TON
func (s *Service) GetBalance(ctx context.Context, addr string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return 0, fmt.Errorf("client not connected")
	}

	parsedAddr, err := address.ParseAddr(addr)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %w", err)
	}

	// Получаем текущий блок
	block, err := s.client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get masterchain info: %w", err)
	}

	// Получаем состояние аккаунта
	acc, err := s.client.GetAccount(ctx, block, parsedAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to get account: %w", err)
	}

	if !acc.IsActive {
		return 0, nil
	}

	// Конвертируем из нано-TON в TON
	balance := float64(acc.State.Balance.Nano().Int64()) / 1e9
	return balance, nil
}

// GetTransactions получает транзакции адреса
func (s *Service) GetTransactions(ctx context.Context, addr string, limit int, afterLt int64, afterHash string) ([]*models.BlockchainTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	parsedAddr, err := address.ParseAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Получаем текущий блок
	block, err := s.client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get masterchain info: %w", err)
	}

	// Получаем состояние аккаунта
	acc, err := s.client.GetAccount(ctx, block, parsedAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if !acc.IsActive {
		return []*models.BlockchainTransaction{}, nil
	}

	// Получаем транзакции
	var txs []*tlb.Transaction
	if afterLt > 0 && afterHash != "" {
		// Получаем транзакции после указанной
		hashBytes, err := base64.StdEncoding.DecodeString(afterHash)
		if err != nil {
			return nil, fmt.Errorf("invalid hash format: %w", err)
		}

		txs, err = s.client.ListTransactions(ctx, parsedAddr, uint32(limit), uint64(afterLt), hashBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to list transactions: %w", err)
		}
	} else {
		// Получаем последние транзакции
		txs, err = s.client.ListTransactions(ctx, parsedAddr, uint32(limit), acc.LastTxLT, acc.LastTxHash)
		if err != nil {
			return nil, fmt.Errorf("failed to list transactions: %w", err)
		}
	}

	// Конвертируем в наш формат
	result := make([]*models.BlockchainTransaction, 0, len(txs))
	for _, tx := range txs {
		bcTx := s.convertTransaction(tx, parsedAddr)
		if bcTx != nil {
			result = append(result, bcTx)
		}
	}

	return result, nil
}

// GetNewTransactions получает новые транзакции после указанного lt
func (s *Service) GetNewTransactions(ctx context.Context, afterLt int64) ([]*models.BlockchainTransaction, error) {
	return s.GetTransactions(ctx, s.GetMasterWalletAddress(), 100, 0, "")
}

// convertTransaction конвертирует транзакцию TON в наш формат
func (s *Service) convertTransaction(tx *tlb.Transaction, ourAddr *address.Address) *models.BlockchainTransaction {
	if tx == nil {
		return nil
	}

	result := &models.BlockchainTransaction{
		Lt:        int64(tx.LT),
		Timestamp: time.Unix(int64(tx.Now), 0),
		Currency:  models.CurrencyTON,
	}

	// Получаем хеш транзакции
	result.Hash = base64.StdEncoding.EncodeToString(tx.Hash)

	// Обрабатываем входящее сообщение
	if tx.IO.In != nil {
		inMsg := tx.IO.In.AsInternal()
		if inMsg != nil {
			result.FromAddress = inMsg.SrcAddr.String()
			result.ToAddress = ourAddr.String()
			result.Amount = float64(inMsg.Amount.Nano().Int64()) / 1e9
			result.IsIncoming = true

			// Извлекаем комментарий
			if inMsg.Body != nil {
				comment := s.extractComment(inMsg.Body)
				result.Comment = comment
			}
		}
	}

	// Обрабатываем исходящие сообщения
	if tx.IO.Out != nil {
		outList, err := tx.IO.Out.ToSlice()
		if err == nil {
			for _, outMsgItem := range outList {
				outMsg := outMsgItem.AsInternal()
				if outMsg != nil {
					result.FromAddress = ourAddr.String()
					result.ToAddress = outMsg.DstAddr.String()
					result.Amount = float64(outMsg.Amount.Nano().Int64()) / 1e9
					result.IsIncoming = false
					break // Берём первое исходящее сообщение
				}
			}
		}
	}

	// Комиссия
	if !tx.TotalFees.Coins.IsZero() {
		result.Fee = float64(tx.TotalFees.Coins.Nano().Int64()) / 1e9
	}

	return result
}

// extractComment извлекает комментарий из тела сообщения
func (s *Service) extractComment(body *cell.Cell) string {
	if body == nil {
		return ""
	}

	slice := body.BeginParse()
	op, err := slice.LoadUInt(32)
	if err != nil {
		return ""
	}

	// 0 - это opcode для текстового комментария
	if op == 0 {
		comment, err := slice.LoadStringSnake()
		if err != nil {
			return ""
		}
		return comment
	}

	return ""
}

// SendTON отправляет TON на указанный адрес
func (s *Service) SendTON(ctx context.Context, toAddress string, amount float64, comment string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.masterWallet == nil {
		return "", fmt.Errorf("master wallet not initialized")
	}

	toAddr, err := address.ParseAddr(toAddress)
	if err != nil {
		return "", fmt.Errorf("invalid destination address: %w", err)
	}

	// Конвертируем сумму в нано-TON
	amountNano := tlb.MustFromTON(fmt.Sprintf("%.9f", amount))

	// Создаём сообщение
	var body *cell.Cell
	if comment != "" {
		// Создаём комментарий
		body = cell.BeginCell().
			MustStoreUInt(0, 32). // opcode для текстового комментария
			MustStoreStringSnake(comment).
			EndCell()
	}

	// Отправляем транзакцию
	err = s.masterWallet.Send(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: true,
			Bounce:      false,
			DstAddr:     toAddr,
			Amount:      amountNano,
			Body:        body,
		},
	}, true)

	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	// Возвращаем хеш (в реальности нужно ждать подтверждения и получить хеш)
	txHash := fmt.Sprintf("pending_%d", time.Now().UnixNano())
	return txHash, nil
}

// SendUSDT отправляет USDT Jetton на указанный адрес
func (s *Service) SendUSDT(ctx context.Context, toAddress string, amount float64, comment string) (string, error) {
	// USDT Jetton transfer требует более сложной логики
	// Здесь нужно отправить jetton transfer на jetton wallet
	return "", fmt.Errorf("USDT transfers not yet implemented")
}

// GeneratePaymentDeepLink генерирует deep link для оплаты через TON кошелёк
func (s *Service) GeneratePaymentDeepLink(toAddress string, amount float64, comment string) string {
	// Формат: ton://transfer/<address>?amount=<nanotons>&text=<comment>
	amountNano := int64(amount * 1e9)

	params := url.Values{}
	params.Set("amount", fmt.Sprintf("%d", amountNano))
	if comment != "" {
		params.Set("text", comment)
	}

	return fmt.Sprintf("ton://transfer/%s?%s", toAddress, params.Encode())
}

// GenerateTonkeeperDeepLink генерирует deep link для Tonkeeper
func (s *Service) GenerateTonkeeperDeepLink(toAddress string, amount float64, comment string) string {
	amountNano := int64(amount * 1e9)

	params := url.Values{}
	params.Set("amount", fmt.Sprintf("%d", amountNano))
	if comment != "" {
		params.Set("text", comment)
	}

	return fmt.Sprintf("https://app.tonkeeper.com/transfer/%s?%s", toAddress, params.Encode())
}

// GetMinWithdrawAmount возвращает минимальную сумму вывода
func (s *Service) GetMinWithdrawAmount(currency string) float64 {
	switch currency {
	case models.CurrencyTON:
		return s.config.MinWithdrawTON
	case models.CurrencyUSDT:
		return s.config.MinWithdrawUSDT
	default:
		return 0.1
	}
}

// GetWithdrawFee возвращает комиссию за вывод
func (s *Service) GetWithdrawFee(currency string) float64 {
	switch currency {
	case models.CurrencyTON:
		return s.config.WithdrawFeeTON
	case models.CurrencyUSDT:
		return 0.5 // Примерная комиссия для Jetton transfer
	default:
		return 0.05
	}
}

// Close закрывает соединение
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client = nil
	s.masterWallet = nil
	return nil
}
