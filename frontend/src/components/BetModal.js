import React, { useState } from 'react';
import styled from 'styled-components';

const ModalOverlay = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background-color: #fff;
  border-radius: 12px;
  padding: 24px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`;

const ModalTitle = styled.h3`
  margin: 0;
  font-size: 20px;
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #999;
  
  &:hover {
    color: #333;
  }
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
`;

const FormGroup = styled.div`
  margin-bottom: 16px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  font-weight: bold;
  font-size: 14px;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  
  &:focus {
    outline: none;
    border-color: #0077cc;
  }
`;

const RangeContainer = styled.div`
  margin-top: 8px;
`;

const RangeInput = styled.input`
  width: 100%;
  margin: 8px 0;
`;

const RangeLabels = styled.div`
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
`;

const InfoText = styled.p`
  font-size: 14px;
  color: #666;
  margin: 12px 0;
`;

const RewardText = styled.div`
  font-size: 16px;
  font-weight: bold;
  color: #0077cc;
  text-align: center;
  margin: 16px 0;
  padding: 8px;
  background-color: #f0f9ff;
  border-radius: 6px;
`;

const Button = styled.button`
  background-color: #0077cc;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 12px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: #0066b3;
  }
  
  &:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
  }
`;

const ErrorText = styled.div`
  color: #e53935;
  font-size: 14px;
  margin-top: 8px;
`;

const BetModal = ({ game, onClose, onSubmit, userBalance }) => {
    const [betAmount, setBetAmount] = useState(game.min_bet);
    const [error, setError] = useState('');

    const { min_bet, max_bet, reward_multiplier, currency } = game;

    // Рассчитываем потенциальный выигрыш
    const potentialReward = betAmount * reward_multiplier;

    const handleSubmit = (e) => {
        e.preventDefault();

        // Проверка на минимальную и максимальную ставку
        if (betAmount < min_bet) {
            setError(`Минимальная ставка: ${min_bet} ${currency}`);
            return;
        }

        if (betAmount > max_bet) {
            setError(`Максимальная ставка: ${max_bet} ${currency}`);
            return;
        }

        // Проверка на достаточность баланса
        const balance = currency === 'TON' ? userBalance.balance_ton : userBalance.balance_usdt;
        if (betAmount > balance) {
            setError(`Недостаточно средств. Ваш баланс: ${balance} ${currency}`);
            return;
        }

        // Отправляем ставку
        onSubmit(betAmount);
    };

    const handleBetChange = (e) => {
        setBetAmount(parseFloat(e.target.value));
        setError('');
    };

    return (
        <ModalOverlay>
            <ModalContent>
                <ModalHeader>
                    <ModalTitle>Сделать ставку</ModalTitle>
                    <CloseButton onClick={onClose}>&times;</CloseButton>
                </ModalHeader>

                <Form onSubmit={handleSubmit}>
                    <FormGroup>
                        <Label>Сумма ставки ({currency})</Label>
                        <Input
                            type="number"
                            min={min_bet}
                            max={max_bet}
                            step="0.1"
                            value={betAmount}
                            onChange={handleBetChange}
                        />

                        <RangeContainer>
                            <RangeInput
                                type="range"
                                min={min_bet}
                                max={max_bet}
                                step="0.1"
                                value={betAmount}
                                onChange={handleBetChange}
                            />
                            <RangeLabels>
                                <span>{min_bet} {currency}</span>
                                <span>{max_bet} {currency}</span>
                            </RangeLabels>
                        </RangeContainer>

                        {error && <ErrorText>{error}</ErrorText>}
                    </FormGroup>

                    <InfoText>
                        Выберите сумму ставки. В случае победы вы получите выигрыш с учетом множителя x{reward_multiplier}.
                    </InfoText>

                    <RewardText>
                        Потенциальный выигрыш: {potentialReward.toFixed(2)} {currency}
                    </RewardText>

                    <Button type="submit">Сделать ставку</Button>
                </Form>
            </ModalContent>
        </ModalOverlay>
    );
};

export default BetModal; 