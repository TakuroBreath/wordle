import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { userAPI } from '../api';
import { useAuth } from '../context/AuthContext';
import { useTonConnect } from '../context/TonConnectContext';

const Container = styled.div`
  max-width: 600px;
  margin: 0 auto;
  padding: 16px;
  padding-bottom: 100px;
`;

const Header = styled.div`
  margin-bottom: 24px;
`;

const BackButton = styled.button`
  background: none;
  border: none;
  color: #0077cc;
  font-size: 16px;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  
  &:hover {
    text-decoration: underline;
  }
`;

const Title = styled.h1`
  font-size: 24px;
  margin: 0 0 8px;
  color: #333;
`;

const Subtitle = styled.p`
  font-size: 16px;
  color: #666;
  margin: 8px 0;
`;

const Card = styled.div`
  background-color: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
`;

const BalanceSection = styled.div`
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 24px;
  color: white;
  margin-bottom: 24px;
  text-align: center;
`;

const BalanceLabel = styled.div`
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
`;

const BalanceRow = styled.div`
  display: flex;
  justify-content: center;
  gap: 32px;
`;

const BalanceItem = styled.div`
  text-align: center;
`;

const BalanceAmount = styled.div`
  font-size: 28px;
  font-weight: 700;
`;

const FormGroup = styled.div`
  margin-bottom: 20px;
`;

const Label = styled.label`
  display: block;
  font-weight: bold;
  margin-bottom: 8px;
  font-size: 14px;
  color: #333;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  box-sizing: border-box;
  
  &:focus {
    outline: none;
    border-color: #0077cc;
  }

  &:disabled {
    background: #f5f5f5;
    color: #666;
  }
`;

const CurrencySelector = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
`;

const CurrencyButton = styled.button`
  padding: 16px;
  border: 2px solid ${props => props.active ? '#0088CC' : '#ddd'};
  border-radius: 12px;
  background: ${props => props.active ? '#e6f4ff' : 'white'};
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
  
  &:hover {
    border-color: #0088CC;
  }
`;

const CurrencyIcon = styled.div`
  font-size: 24px;
  margin-bottom: 4px;
`;

const CurrencyName = styled.div`
  font-weight: 600;
  color: #333;
`;

const CurrencyBalance = styled.div`
  font-size: 12px;
  color: #666;
  margin-top: 4px;
`;

const WalletInfo = styled.div`
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
`;

const WalletLabel = styled.div`
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
  text-transform: uppercase;
`;

const WalletAddress = styled.div`
  font-family: monospace;
  font-size: 14px;
  word-break: break-all;
  color: #333;
`;

const ConnectWalletButton = styled.button`
  background: linear-gradient(135deg, #0088CC 0%, #00AAFF 100%);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  width: 100%;
`;

const QuickAmounts = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
`;

const QuickAmountButton = styled.button`
  padding: 10px 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: ${props => props.active ? '#0088CC' : 'white'};
  color: ${props => props.active ? 'white' : '#333'};
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
  
  &:hover {
    border-color: #0088CC;
  }
`;

const MaxButton = styled.button`
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: #e0e0e0;
  border: none;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 12px;
  cursor: pointer;
  
  &:hover {
    background: #d0d0d0;
  }
`;

const InputWrapper = styled.div`
  position: relative;
`;

const WithdrawButton = styled.button`
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a5a 100%);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 16px 20px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  width: 100%;
  margin-top: 16px;
  transition: transform 0.2s, box-shadow 0.2s;
  
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(255, 107, 107, 0.3);
  }
  
  &:disabled {
    background: #cccccc;
    cursor: not-allowed;
    transform: none;
  }
`;

const FeeInfo = styled.div`
  background: #fff8e1;
  border: 1px solid #ffe082;
  border-radius: 8px;
  padding: 12px;
  margin-top: 16px;
  font-size: 13px;
  color: #f57f17;
`;

const SuccessMessage = styled.div`
  background-color: #e8f5e9;
  color: #2e7d32;
  padding: 16px;
  border-radius: 8px;
  margin: 16px 0;
  text-align: center;
  
  h3 {
    margin: 0 0 8px;
  }
`;

const ErrorMessage = styled.div`
  background-color: #ffebee;
  color: #c62828;
  padding: 12px;
  border-radius: 6px;
  margin: 16px 0;
  font-size: 14px;
`;

const HistorySection = styled.div`
  margin-top: 24px;
`;

const HistoryTitle = styled.h3`
  font-size: 16px;
  margin: 0 0 16px;
  color: #333;
`;

const HistoryItem = styled.div`
  background: white;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid #e0e0e0;
`;

const HistoryAmount = styled.div`
  font-weight: 600;
  color: #333;
`;

const HistoryStatus = styled.span`
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  background: ${props => {
    switch (props.status) {
      case 'completed': return '#e8f5e9';
      case 'pending': return '#fff8e1';
      case 'failed': return '#ffebee';
      default: return '#f5f5f5';
    }
  }};
  color: ${props => {
    switch (props.status) {
      case 'completed': return '#2e7d32';
      case 'pending': return '#f57f17';
      case 'failed': return '#c62828';
      default: return '#666';
    }
  }};
`;

const WithdrawPage = () => {
    const navigate = useNavigate();
    const { user, refreshUserData } = useAuth();
    const { isConnected, address, connectWallet } = useTonConnect();

    const [currency, setCurrency] = useState('TON');
    const [amount, setAmount] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(null);
    const [history, setHistory] = useState([]);

    const balance = currency === 'TON' ? (user?.balance_ton || 0) : (user?.balance_usdt || 0);
    const minWithdraw = currency === 'TON' ? 0.1 : 1;
    const fee = currency === 'TON' ? 0.01 : 0.5;

    // Загрузка истории выводов
    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const response = await userAPI.getWithdrawHistory();
                setHistory(response.data || []);
            } catch (err) {
                console.error('Error fetching history:', err);
            }
        };

        fetchHistory();
    }, [success]);

    // Получение адреса для вывода
    const withdrawAddress = isConnected ? address : user?.wallet;

    // Быстрые суммы
    const getQuickAmounts = () => {
        const maxAmount = Math.max(0, balance - fee);
        if (currency === 'TON') {
            return [0.5, 1, 2, maxAmount].filter(a => a <= maxAmount && a >= minWithdraw);
        }
        return [5, 10, 20, maxAmount].filter(a => a <= maxAmount && a >= minWithdraw);
    };

    // Установка максимальной суммы
    const handleMaxAmount = () => {
        const maxAmount = Math.max(0, balance - fee);
        setAmount(maxAmount.toFixed(currency === 'TON' ? 4 : 2));
    };

    // Обработка отправки формы
    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!withdrawAddress) {
            setError('Подключите кошелек для вывода средств');
            return;
        }

        const withdrawAmount = parseFloat(amount);
        
        if (!withdrawAmount || withdrawAmount <= 0) {
            setError('Укажите сумму для вывода');
            return;
        }

        if (withdrawAmount < minWithdraw) {
            setError(`Минимальная сумма вывода: ${minWithdraw} ${currency}`);
            return;
        }

        if (withdrawAmount + fee > balance) {
            setError(`Недостаточно средств. Доступно: ${balance.toFixed(4)} ${currency}`);
            return;
        }

        try {
            setLoading(true);
            setError(null);
            setSuccess(null);

            await userAPI.requestWithdraw(withdrawAmount, currency, withdrawAddress);

            setSuccess({
                amount: withdrawAmount,
                currency,
                address: withdrawAddress
            });
            setAmount('');

            // Обновляем данные пользователя
            await refreshUserData();
        } catch (err) {
            console.error('Error requesting withdraw:', err);
            setError(err.response?.data?.error || 'Не удалось создать запрос на вывод');
        } finally {
            setLoading(false);
        }
    };

    // Возврат назад
    const handleBack = () => {
        navigate('/profile');
    };

    return (
        <Container>
            <BackButton onClick={handleBack}>← Профиль</BackButton>

            <Header>
                <Title>💸 Вывод средств</Title>
                <Subtitle>Выведите средства на свой TON кошелек</Subtitle>
            </Header>

            {/* Текущий баланс */}
            <BalanceSection>
                <BalanceLabel>Доступно для вывода</BalanceLabel>
                <BalanceRow>
                    <BalanceItem>
                        <BalanceAmount>{(user?.balance_ton || 0).toFixed(4)}</BalanceAmount>
                        <BalanceLabel>TON</BalanceLabel>
                    </BalanceItem>
                    <BalanceItem>
                        <BalanceAmount>{(user?.balance_usdt || 0).toFixed(2)}</BalanceAmount>
                        <BalanceLabel>USDT</BalanceLabel>
                    </BalanceItem>
                </BalanceRow>
            </BalanceSection>

            {error && <ErrorMessage>{error}</ErrorMessage>}
            
            {success && (
                <SuccessMessage>
                    <h3>✅ Запрос на вывод создан!</h3>
                    <p>{success.amount} {success.currency} будет отправлено на</p>
                    <p style={{ fontFamily: 'monospace', fontSize: '12px' }}>
                        {success.address}
                    </p>
                </SuccessMessage>
            )}

            <Card>
                {/* Адрес кошелька */}
                <FormGroup>
                    <Label>Кошелек для вывода</Label>
                    {withdrawAddress ? (
                        <WalletInfo>
                            <WalletLabel>TON адрес</WalletLabel>
                            <WalletAddress>{withdrawAddress}</WalletAddress>
                        </WalletInfo>
                    ) : (
                        <ConnectWalletButton onClick={connectWallet}>
                            🔗 Подключить кошелек
                        </ConnectWalletButton>
                    )}
                </FormGroup>

                {/* Выбор валюты */}
                <FormGroup>
                    <Label>Валюта</Label>
                    <CurrencySelector>
                        <CurrencyButton 
                            active={currency === 'TON'} 
                            onClick={() => { setCurrency('TON'); setAmount(''); }}
                        >
                            <CurrencyIcon>💎</CurrencyIcon>
                            <CurrencyName>TON</CurrencyName>
                            <CurrencyBalance>
                                {(user?.balance_ton || 0).toFixed(4)}
                            </CurrencyBalance>
                        </CurrencyButton>
                        <CurrencyButton 
                            active={currency === 'USDT'} 
                            onClick={() => { setCurrency('USDT'); setAmount(''); }}
                        >
                            <CurrencyIcon>💵</CurrencyIcon>
                            <CurrencyName>USDT</CurrencyName>
                            <CurrencyBalance>
                                {(user?.balance_usdt || 0).toFixed(2)}
                            </CurrencyBalance>
                        </CurrencyButton>
                    </CurrencySelector>
                </FormGroup>

                {/* Сумма */}
                <FormGroup>
                    <Label>Сумма вывода ({currency})</Label>
                    <QuickAmounts>
                        {getQuickAmounts().map((amt, idx) => (
                            <QuickAmountButton
                                key={idx}
                                active={parseFloat(amount) === amt}
                                onClick={() => setAmount(amt.toFixed(currency === 'TON' ? 4 : 2))}
                            >
                                {idx === getQuickAmounts().length - 1 ? 'MAX' : amt}
                            </QuickAmountButton>
                        ))}
                    </QuickAmounts>
                    <InputWrapper>
                        <Input
                            type="number"
                            value={amount}
                            onChange={(e) => setAmount(e.target.value)}
                            placeholder={`Мин. ${minWithdraw} ${currency}`}
                            step={currency === 'TON' ? '0.0001' : '0.01'}
                            min={minWithdraw}
                            max={balance - fee}
                        />
                        <MaxButton type="button" onClick={handleMaxAmount}>MAX</MaxButton>
                    </InputWrapper>
                </FormGroup>

                <FeeInfo>
                    ⚠️ Комиссия сети: {fee} {currency}
                    {amount && parseFloat(amount) > 0 && (
                        <> • К получению: {Math.max(0, parseFloat(amount)).toFixed(currency === 'TON' ? 4 : 2)} {currency}</>
                    )}
                </FeeInfo>

                <WithdrawButton 
                    onClick={handleSubmit} 
                    disabled={loading || !withdrawAddress || !amount}
                >
                    {loading ? 'Обработка...' : `Вывести ${amount || '0'} ${currency}`}
                </WithdrawButton>
            </Card>

            {/* История выводов */}
            {history.length > 0 && (
                <HistorySection>
                    <HistoryTitle>История выводов</HistoryTitle>
                    {history.slice(0, 5).map((item, idx) => (
                        <HistoryItem key={idx}>
                            <div>
                                <HistoryAmount>
                                    -{item.amount} {item.currency}
                                </HistoryAmount>
                                <div style={{ fontSize: '12px', color: '#666' }}>
                                    {new Date(item.created_at).toLocaleDateString()}
                                </div>
                            </div>
                            <HistoryStatus status={item.status}>
                                {item.status === 'completed' && '✓ Выполнен'}
                                {item.status === 'pending' && '⏳ В обработке'}
                                {item.status === 'failed' && '✗ Ошибка'}
                            </HistoryStatus>
                        </HistoryItem>
                    ))}
                </HistorySection>
            )}
        </Container>
    );
};

export default WithdrawPage;
