import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { gameAPI, lobbyAPI } from '../api';
import { useAuth } from '../context/AuthContext';
import BetModal from '../components/BetModal';

const Container = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: 16px;
`;

const GameHeader = styled.div`
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
  display: flex;
  align-items: center;
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

const Description = styled.p`
  font-size: 16px;
  color: #666;
  margin: 8px 0;
`;

const InfoSection = styled.div`
  background-color: #f9f9f9;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
`;

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
  
  &:last-child {
    border-bottom: none;
  }
`;

const InfoLabel = styled.span`
  color: #666;
  font-size: 14px;
`;

const InfoValue = styled.span`
  font-weight: bold;
  color: #333;
  font-size: 14px;
`;

const RewardPoolSection = styled.div`
  background-color: #f0f9ff;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  text-align: center;
`;

const RewardPoolAmount = styled.div`
  font-size: 24px;
  font-weight: bold;
  color: #0077cc;
  margin-bottom: 8px;
`;

const RewardPoolLabel = styled.div`
  font-size: 14px;
  color: #666;
`;

const ActionButton = styled.button`
  background-color: #0077cc;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 12px 20px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  width: 100%;
  margin-bottom: 12px;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: #0066b3;
  }
  
  &:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
  }
`;

const SecondaryButton = styled(ActionButton)`
  background-color: #f0f0f0;
  color: #333;
  
  &:hover {
    background-color: #e0e0e0;
  }
`;

const LoadingIndicator = styled.div`
  text-align: center;
  padding: 40px 0;
  color: #666;
`;

const ErrorMessage = styled.div`
  background-color: #ffebee;
  color: #c62828;
  padding: 12px;
  border-radius: 6px;
  margin: 16px 0;
  font-size: 14px;
`;

const GamePage = () => {
    const { id } = useParams();
    const { user, isAuthenticated } = useAuth();
    const navigate = useNavigate();

    const [game, setGame] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showBetModal, setShowBetModal] = useState(false);

    // Загрузка данных игры
    const fetchGameData = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);

            const response = await gameAPI.getGame(id);
            setGame(response.data);
        } catch (err) {
            console.error('Error fetching game:', err);
            setError('Не удалось загрузить информацию об игре');
        } finally {
            setLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchGameData();
    }, [fetchGameData]);

    // Возврат на главную страницу
    const handleBack = () => {
        navigate('/');
    };

    // Активация/деактивация игры (для создателя)
    const handleToggleGameStatus = async () => {
        try {
            setLoading(true);

            if (game.status === 'active') {
                await gameAPI.deactivateGame(id);
            } else {
                await gameAPI.activateGame(id);
            }

            // Обновляем данные игры
            fetchGameData();
        } catch (err) {
            console.error('Error toggling game status:', err);
            setError('Не удалось изменить статус игры');
        } finally {
            setLoading(false);
        }
    };

    // Удаление игры (для создателя)
    const handleDeleteGame = async () => {
        if (!window.confirm('Вы уверены, что хотите удалить эту игру?')) {
            return;
        }

        try {
            setLoading(true);
            await gameAPI.deleteGame(id);
            navigate('/');
        } catch (err) {
            console.error('Error deleting game:', err);
            setError('Не удалось удалить игру');
            setLoading(false);
        }
    };

    // Присоединение к игре
    const handleJoinGame = () => {
        if (!isAuthenticated) {
            setError('Необходимо авторизоваться для участия в игре');
            return;
        }

        setShowBetModal(true);
    };

    // Подтверждение ставки
    const handleBetSubmit = async (betAmount) => {
        try {
            setLoading(true);
            const response = await lobbyAPI.joinGame(id, betAmount);
            setShowBetModal(false);

            // Переходим на страницу лобби
            navigate(`/lobbies/${response.data.id}`);
        } catch (err) {
            console.error('Error joining game:', err);
            setError(err.response?.data?.error || 'Не удалось присоединиться к игре');
            setLoading(false);
        }
    };

    // Определяем цвет для сложности
    const getDifficultyColor = (difficulty) => {
        switch (difficulty?.toLowerCase()) {
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
        return parseFloat(amount || 0).toFixed(2);
    };

    // Проверяем, является ли текущий пользователь создателем игры
    const isCreator = game && user && game.creator_id === user.telegram_id;

    if (loading && !game) {
        return (
            <Container>
                <LoadingIndicator>Загрузка информации об игре...</LoadingIndicator>
            </Container>
        );
    }

    if (error && !game) {
        return (
            <Container>
                <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>
                <ErrorMessage>{error}</ErrorMessage>
            </Container>
        );
    }

    if (!game) {
        return (
            <Container>
                <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>
                <ErrorMessage>Игра не найдена</ErrorMessage>
            </Container>
        );
    }

    return (
        <Container>
            <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>

            {error && <ErrorMessage>{error}</ErrorMessage>}

            <GameHeader>
                <Title>
                    {game.title}
                    <Badge color={getDifficultyColor(game.difficulty)}>
                        {game.difficulty}
                    </Badge>
                    {isCreator && (
                        <Badge color="#9c27b0">Моя игра</Badge>
                    )}
                </Title>
                <Description>{game.description || 'Без описания'}</Description>
            </GameHeader>

            <InfoSection>
                <InfoRow>
                    <InfoLabel>Длина слова</InfoLabel>
                    <InfoValue>{game.length}</InfoValue>
                </InfoRow>
                <InfoRow>
                    <InfoLabel>Количество попыток</InfoLabel>
                    <InfoValue>{game.max_tries}</InfoValue>
                </InfoRow>
                <InfoRow>
                    <InfoLabel>Ставка</InfoLabel>
                    <InfoValue>{formatAmount(game.min_bet)} - {formatAmount(game.max_bet)} {game.currency}</InfoValue>
                </InfoRow>
                <InfoRow>
                    <InfoLabel>Множитель награды</InfoLabel>
                    <InfoValue>x{game.reward_multiplier}</InfoValue>
                </InfoRow>
                <InfoRow>
                    <InfoLabel>Статус</InfoLabel>
                    <InfoValue>{game.status === 'active' ? 'Активна' : 'Неактивна'}</InfoValue>
                </InfoRow>
            </InfoSection>

            <RewardPoolSection>
                <RewardPoolAmount>
                    {game.currency === 'TON'
                        ? formatAmount(game.reward_pool_ton)
                        : formatAmount(game.reward_pool_usdt)} {game.currency}
                </RewardPoolAmount>
                <RewardPoolLabel>Призовой фонд</RewardPoolLabel>
            </RewardPoolSection>

            {isCreator ? (
                <>
                    <ActionButton
                        onClick={handleToggleGameStatus}
                        disabled={loading}
                    >
                        {game.status === 'active' ? 'Деактивировать игру' : 'Активировать игру'}
                    </ActionButton>
                    <SecondaryButton
                        onClick={handleDeleteGame}
                        disabled={loading}
                    >
                        Удалить игру
                    </SecondaryButton>
                </>
            ) : (
                <ActionButton
                    onClick={handleJoinGame}
                    disabled={loading || game.status !== 'active' || (user && game.creator_id === user.telegram_id)}
                >
                    Играть
                </ActionButton>
            )}

            {showBetModal && (
                <BetModal
                    game={game}
                    onClose={() => setShowBetModal(false)}
                    onSubmit={handleBetSubmit}
                    userBalance={user}
                />
            )}
        </Container>
    );
};

export default GamePage; 