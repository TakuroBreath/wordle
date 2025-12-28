import React from 'react';
import styled from 'styled-components';

const Card = styled.div`
  background-color: #ffffff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 16px;
  margin-bottom: 16px;
  width: 100%;
  transition: transform 0.2s ease;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }
`;

const Title = styled.h3`
  margin: 0 0 8px;
  font-size: 18px;
  color: #000;
`;

const Description = styled.p`
  margin: 0 0 12px;
  font-size: 14px;
  color: #666;
`;

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
`;

const InfoItem = styled.div`
  font-size: 14px;
  color: #444;
  
  span {
    font-weight: bold;
    color: #000;
  }
`;

const RewardPool = styled.div`
  background-color: #f0f9ff;
  border-radius: 8px;
  padding: 8px 12px;
  margin: 8px 0;
  font-size: 16px;
  font-weight: bold;
  color: #0077cc;
  text-align: center;
`;

const ButtonRow = styled.div`
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  gap: 10px;
`;

const Button = styled.button`
  background-color: ${props => props.primary ? '#0077cc' : '#f0f0f0'};
  color: ${props => props.primary ? '#ffffff' : '#333333'};
  border: none;
  border-radius: 6px;
  padding: 10px 16px;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  flex: 1;
  transition: background-color 0.2s ease;
  
  &:hover {
    background-color: ${props => props.primary ? '#0066b3' : '#e0e0e0'};
  }
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const Badge = styled.span`
  background-color: ${props => props.color || '#e0e0e0'};
  color: #fff;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  margin-left: 8px;
`;

const GameCard = ({ game, onJoin, onDetails }) => {
    const {
        id,
        title,
        description,
        difficulty,
        // word is intentionally not exposed to prevent cheating
        length,
        max_tries: maxTries,
        min_bet: minBet,
        max_bet: maxBet,
        reward_multiplier: rewardMultiplier,
        reward_pool_ton: rewardPoolTon,
        reward_pool_usdt: rewardPoolUsdt,
        currency,
        status
    } = game;

    // Определяем цвет для сложности
    const getDifficultyColor = (difficulty) => {
        switch (difficulty.toLowerCase()) {
            case 'easy':
                return '#4caf50';
            case 'medium':
                return '#ff9800';
            case 'hard':
                return '#f44336';
            default:
                return '#9e9e9e';
        }
    };

    // Форматируем сумму с двумя знаками после запятой
    const formatAmount = (amount) => {
        return parseFloat(amount).toFixed(2);
    };

    return (
        <Card>
            <Title>
                {title}
                <Badge color={getDifficultyColor(difficulty)}>{difficulty}</Badge>
            </Title>
            <Description>{description || 'Без описания'}</Description>

            <InfoRow>
                <InfoItem>Длина слова: <span>{length}</span></InfoItem>
                <InfoItem>Попыток: <span>{maxTries}</span></InfoItem>
            </InfoRow>

            <InfoRow>
                <InfoItem>Ставка: <span>{formatAmount(minBet)} - {formatAmount(maxBet)} {currency}</span></InfoItem>
                <InfoItem>Множитель: <span>x{rewardMultiplier}</span></InfoItem>
            </InfoRow>

            <RewardPool>
                Призовой фонд: {currency === 'TON' ? formatAmount(rewardPoolTon) : formatAmount(rewardPoolUsdt)} {currency}
            </RewardPool>

            <ButtonRow>
                <Button onClick={() => onDetails(id)}>Подробнее</Button>
                <Button primary onClick={() => onJoin(id)} disabled={status !== 'active'}>
                    Играть
                </Button>
            </ButtonRow>
        </Card>
    );
};

export default GameCard;