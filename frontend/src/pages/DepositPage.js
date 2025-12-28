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

const BalanceAmount = styled.div`
  font-size: 32px;
  font-weight: 700;
`;

const BalanceRow = styled.div`
  display: flex;
  justify-content: center;
  gap: 32px;
  margin-top: 16px;
`;

const BalanceItem = styled.div`
  text-align: center;
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

const DepositInfo = styled.div`
  background: #f8f9fa;
  border-radius: 12px;
  padding: 20px;
  margin-top: 20px;
`;

const DepositLabel = styled.div`
  font-size: 12px;
  color: #666;
  text-transform: uppercase;
  margin-bottom: 8px;
`;

const DepositAddress = styled.div`
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  font-family: monospace;
  font-size: 14px;
  word-break: break-all;
  margin-bottom: 12px;
`;

const CopyButton = styled.button`
  background: #e0e0e0;
  border: none;
  border-radius: 6px;
  padding: 10px 16px;
  font-size: 14px;
  cursor: pointer;
  width: 100%;
  transition: background 0.2s;
  
  &:hover {
    background: #d0d0d0;
  }
`;

const QuickAmounts = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
`;

const QuickAmountButton = styled.button`
  padding: 12px 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: ${props => props.active ? '#0088CC' : 'white'};
  color: ${props => props.active ? 'white' : '#333'};
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
  
  &:hover {
    border-color: #0088CC;
  }
`;

const PayButton = styled.button`
  background: linear-gradient(135deg, #0088CC 0%, #00AAFF 100%);
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
    box-shadow: 0 4px 12px rgba(0, 136, 204, 0.3);
  }
  
  &:disabled {
    background: #cccccc;
    cursor: not-allowed;
    transform: none;
  }
`;

const InfoBox = styled.div`
  background: #fff3e0;
  border: 1px solid #ffcc80;
  border-radius: 8px;
  padding: 16px;
  margin-top: 16px;
  font-size: 14px;
  line-height: 1.5;
  
  strong {
    display: block;
    color: #e65100;
    margin-bottom: 8px;
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

const DepositPage = () => {
    const navigate = useNavigate();
    const { user, refreshUserData } = useAuth();
    const { isConnected, sendTransaction, getPaymentDeepLink } = useTonConnect();

    const [currency, setCurrency] = useState('TON');
    const [amount, setAmount] = useState(1);
    const [depositInfo, setDepositInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [copied, setCopied] = useState(false);

    const quickAmounts = currency === 'TON' ? [0.5, 1, 2, 5] : [5, 10, 20, 50];

    // Загрузка информации о депозите
    useEffect(() => {
        const fetchDepositInfo = async () => {
            try {
                const response = await userAPI.getDepositInfo();
                setDepositInfo(response.data);
            } catch (err) {
                console.error('Error fetching deposit info:', err);
            }
        };

        fetchDepositInfo();
    }, []);

    // Копирование адреса
    const handleCopyAddress = async () => {
        if (!depositInfo?.address) return;
        
        try {
            await navigator.clipboard.writeText(depositInfo.address);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    // Оплата через TON Connect
    const handleTonConnectPayment = async () => {
        if (!depositInfo?.address || !isConnected) return;

        try {
            setLoading(true);
            setError(null);
            
            await sendTransaction(depositInfo.address, amount, `deposit_${user.telegram_id}`);
            
            // Обновляем данные пользователя
            await refreshUserData();
            
            // Показываем сообщение об успехе
            alert('Транзакция отправлена! Баланс обновится после подтверждения.');
        } catch (err) {
            console.error('Payment error:', err);
            setError('Ошибка при оплате. Попробуйте еще раз.');
        } finally {
            setLoading(false);
        }
    };

    // Открытие deep link
    const handleDeepLinkPayment = () => {
        if (!depositInfo?.address) return;
        
        const deepLink = getPaymentDeepLink(
            depositInfo.address,
            amount,
            `deposit_${user.telegram_id}`
        );
        
        window.open(deepLink, '_blank');
    };

    // Возврат назад
    const handleBack = () => {
        navigate('/profile');
    };

    return (
        <Container>
            <BackButton onClick={handleBack}>← Профиль</BackButton>

            <Header>
                <Title>💰 Пополнение баланса</Title>
                <Subtitle>Пополните баланс для участия в играх</Subtitle>
            </Header>

            {/* Текущий баланс */}
            <BalanceSection>
                <BalanceLabel>Ваш баланс</BalanceLabel>
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

            <Card>
                {/* Выбор валюты */}
                <FormGroup>
                    <Label>Выберите валюту</Label>
                    <CurrencySelector>
                        <CurrencyButton 
                            active={currency === 'TON'} 
                            onClick={() => setCurrency('TON')}
                        >
                            <CurrencyIcon>💎</CurrencyIcon>
                            <CurrencyName>TON</CurrencyName>
                        </CurrencyButton>
                        <CurrencyButton 
                            active={currency === 'USDT'} 
                            onClick={() => setCurrency('USDT')}
                        >
                            <CurrencyIcon>💵</CurrencyIcon>
                            <CurrencyName>USDT</CurrencyName>
                        </CurrencyButton>
                    </CurrencySelector>
                </FormGroup>

                {/* Быстрые суммы */}
                <FormGroup>
                    <Label>Сумма пополнения</Label>
                    <QuickAmounts>
                        {quickAmounts.map(amt => (
                            <QuickAmountButton
                                key={amt}
                                active={amount === amt}
                                onClick={() => setAmount(amt)}
                            >
                                {amt} {currency}
                            </QuickAmountButton>
                        ))}
                    </QuickAmounts>
                </FormGroup>

                {/* Информация о депозите */}
                {depositInfo && (
                    <DepositInfo>
                        <DepositLabel>Адрес для пополнения</DepositLabel>
                        <DepositAddress>{depositInfo.address}</DepositAddress>
                        <CopyButton onClick={handleCopyAddress}>
                            {copied ? '✓ Скопировано!' : '📋 Копировать адрес'}
                        </CopyButton>
                    </DepositInfo>
                )}

                {/* Кнопки оплаты */}
                {isConnected ? (
                    <PayButton onClick={handleTonConnectPayment} disabled={loading}>
                        {loading ? 'Отправка...' : `🔗 Пополнить ${amount} ${currency} через TON Connect`}
                    </PayButton>
                ) : (
                    <PayButton onClick={handleDeepLinkPayment}>
                        📱 Открыть кошелек для пополнения
                    </PayButton>
                )}

                <InfoBox>
                    <strong>ℹ️ Как пополнить баланс</strong>
                    1. Отправьте {currency} на указанный адрес<br/>
                    2. Дождитесь подтверждения транзакции (1-2 мин)<br/>
                    3. Баланс обновится автоматически
                </InfoBox>
            </Card>
        </Container>
    );
};

export default DepositPage;
