import React, { useState, useMemo } from 'react';
import styled from 'styled-components';
import { useTonConnect } from '../context/TonConnectContext';

const ModalOverlay = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
`;

const ModalContent = styled.div`
  background-color: #fff;
  border-radius: 16px;
  padding: 24px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const ModalTitle = styled.h3`
  margin: 0;
  font-size: 20px;
  color: #333;
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #999;
  line-height: 1;
  padding: 0;
  
  &:hover {
    color: #333;
  }
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
`;

const FormGroup = styled.div`
  margin-bottom: 20px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  font-size: 14px;
  color: #333;
`;

const Input = styled.input`
  width: 100%;
  padding: 14px;
  border: 2px solid #e0e0e0;
  border-radius: 10px;
  font-size: 18px;
  text-align: center;
  font-weight: 600;
  box-sizing: border-box;
  
  &:focus {
    outline: none;
    border-color: #0088CC;
  }
`;

const QuickBets = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-top: 12px;
`;

const QuickBetButton = styled.button`
  padding: 10px;
  border: 1px solid ${props => props.active ? '#0088CC' : '#ddd'};
  border-radius: 8px;
  background: ${props => props.active ? '#e6f4ff' : 'white'};
  color: ${props => props.active ? '#0088CC' : '#666'};
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  
  &:hover {
    border-color: #0088CC;
  }
`;

const BalanceInfo = styled.div`
  background: #f5f5f5;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
`;

const BalanceLabel = styled.span`
  color: #666;
`;

const BalanceValue = styled.span`
  font-weight: 600;
  color: ${props => props.sufficient ? '#4caf50' : '#f44336'};
`;

const RewardSection = styled.div`
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f9ff 100%);
  border-radius: 12px;
  padding: 16px;
  text-align: center;
  margin: 16px 0;
`;

const RewardLabel = styled.div`
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
`;

const RewardValue = styled.div`
  font-size: 24px;
  font-weight: 700;
  color: #0088CC;
`;

const MultiplierBadge = styled.span`
  background: #0088CC;
  color: white;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  margin-left: 8px;
`;

const PaymentMethod = styled.div`
  background: ${props => props.isBlockchain ? '#fff8e1' : '#e8f5e9'};
  border: 1px solid ${props => props.isBlockchain ? '#ffcc80' : '#a5d6a7'};
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
`;

const Button = styled.button`
  background: linear-gradient(135deg, #0088CC 0%, #00AAFF 100%);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 16px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
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

const ErrorText = styled.div`
  color: #e53935;
  font-size: 14px;
  margin-top: 8px;
  text-align: center;
`;

const InfoText = styled.p`
  font-size: 13px;
  color: #666;
  text-align: center;
  margin: 12px 0 0;
`;

const BetModal = ({ game, onClose, onSubmit, userBalance, paymentInfo }) => {
    const [betAmount, setBetAmount] = useState(game.min_bet);
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    
    const { isConnected, sendTransaction, getPaymentDeepLink } = useTonConnect();

    const { min_bet, max_bet, reward_multiplier, currency, time_limit } = game;

    // Получаем баланс пользователя
    const userCurrencyBalance = useMemo(() => {
        if (!userBalance) return 0;
        return currency === 'TON' ? (userBalance.balance_ton || 0) : (userBalance.balance_usdt || 0);
    }, [userBalance, currency]);

    // Проверка достаточности баланса
    const hasSufficientBalance = userCurrencyBalance >= betAmount;

    // Рассчитываем потенциальный выигрыш
    const potentialReward = betAmount * reward_multiplier;

    // Быстрые ставки
    const quickBets = useMemo(() => {
        const range = max_bet - min_bet;
        return [
            min_bet,
            min_bet + range * 0.33,
            min_bet + range * 0.66,
            max_bet
        ].map(v => Math.round(v * 100) / 100);
    }, [min_bet, max_bet]);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        // Валидация ставки
        if (betAmount < min_bet) {
            setError(`Минимальная ставка: ${min_bet} ${currency}`);
            return;
        }

        if (betAmount > max_bet) {
            setError(`Максимальная ставка: ${max_bet} ${currency}`);
            return;
        }

        // Если достаточно баланса - используем баланс
        if (hasSufficientBalance) {
            onSubmit(betAmount);
            return;
        }

        // Если недостаточно баланса - используем блокчейн платеж
        if (paymentInfo) {
            try {
                setLoading(true);
                
                if (isConnected) {
                    // Оплата через TON Connect
                    await sendTransaction(
                        paymentInfo.address,
                        betAmount,
                        paymentInfo.comment
                    );
                } else {
                    // Открываем deep link
                    const deepLink = getPaymentDeepLink(
                        paymentInfo.address,
                        betAmount,
                        paymentInfo.comment
                    );
                    window.open(deepLink, '_blank');
                }
                
                // Вызываем onSubmit для обработки результата
                onSubmit(betAmount, true); // true означает blockchain payment
            } catch (err) {
                console.error('Payment error:', err);
                setError('Ошибка платежа. Попробуйте еще раз.');
            } finally {
                setLoading(false);
            }
        } else {
            setError(`Недостаточно средств. Ваш баланс: ${userCurrencyBalance.toFixed(4)} ${currency}`);
        }
    };

    const handleBetChange = (value) => {
        const numValue = parseFloat(value);
        if (!isNaN(numValue)) {
            setBetAmount(Math.min(Math.max(numValue, min_bet), max_bet));
        }
        setError('');
    };

    return (
        <ModalOverlay onClick={onClose}>
            <ModalContent onClick={e => e.stopPropagation()}>
                <ModalHeader>
                    <ModalTitle>🎮 Сделать ставку</ModalTitle>
                    <CloseButton onClick={onClose}>&times;</CloseButton>
                </ModalHeader>

                <Form onSubmit={handleSubmit}>
                    {/* Баланс пользователя */}
                    <BalanceInfo>
                        <BalanceLabel>Ваш баланс:</BalanceLabel>
                        <BalanceValue sufficient={hasSufficientBalance}>
                            {userCurrencyBalance.toFixed(currency === 'TON' ? 4 : 2)} {currency}
                        </BalanceValue>
                    </BalanceInfo>

                    {/* Сумма ставки */}
                    <FormGroup>
                        <Label>Сумма ставки ({currency})</Label>
                        <Input
                            type="number"
                            min={min_bet}
                            max={max_bet}
                            step="0.01"
                            value={betAmount}
                            onChange={(e) => handleBetChange(e.target.value)}
                        />
                        
                        <QuickBets>
                            {quickBets.map((amount, idx) => (
                                <QuickBetButton
                                    key={idx}
                                    type="button"
                                    active={betAmount === amount}
                                    onClick={() => handleBetChange(amount)}
                                >
                                    {idx === quickBets.length - 1 ? 'MAX' : amount}
                                </QuickBetButton>
                            ))}
                        </QuickBets>
                    </FormGroup>

                    {/* Способ оплаты */}
                    <PaymentMethod isBlockchain={!hasSufficientBalance}>
                        {hasSufficientBalance ? (
                            <>✅ Оплата с баланса аккаунта</>
                        ) : (
                            <>💎 Оплата через {isConnected ? 'TON Connect' : 'кошелек'}</>
                        )}
                    </PaymentMethod>

                    {/* Потенциальный выигрыш */}
                    <RewardSection>
                        <RewardLabel>Возможный выигрыш</RewardLabel>
                        <RewardValue>
                            {potentialReward.toFixed(currency === 'TON' ? 4 : 2)} {currency}
                            <MultiplierBadge>×{reward_multiplier}</MultiplierBadge>
                        </RewardValue>
                    </RewardSection>

                    {error && <ErrorText>{error}</ErrorText>}

                    <Button type="submit" disabled={loading}>
                        {loading ? 'Обработка...' : `Играть за ${betAmount} ${currency}`}
                    </Button>

                    <InfoText>
                        ⏱️ Время на игру: {time_limit || 5} мин • {game.max_tries} попыток
                    </InfoText>
                </Form>
            </ModalContent>
        </ModalOverlay>
    );
};

export default BetModal;
