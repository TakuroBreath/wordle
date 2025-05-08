import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { useTelegram } from '../contexts/TelegramContext';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import Button from '../components/UI/Button';
import Card from '../components/UI/Card';
import LoadingScreen from '../components/LoadingScreen';

// Контейнер страницы
const PageContainer = styled.div`
  padding: 16px;
  max-width: 600px;
  margin: 0 auto;
  width: 100%;
`;

// Заголовок страницы
const PageTitle = styled.h1`
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 16px 0;
  color: var(--tg-theme-text-color, #000000);
`;

// Информация об игре
const GameInfo = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
`;

// Блок с отдельной информацией
const InfoItem = styled.div`
  display: flex;
  flex-direction: column;
  min-width: 120px;
`;

// Заголовок информации
const InfoLabel = styled.span`
  font-size: 12px;
  color: var(--tg-theme-hint-color, #999999);
`;

// Значение информации
const InfoValue = styled.span`
  font-size: 16px;
  font-weight: 500;
  color: var(--tg-theme-text-color, #000000);
`;

// Форматирование денежных значений
const formatCurrency = (amount, currency) => {
    return `${amount.toFixed(2)} ${currency}`;
};

const GamePage = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const { tg, isReady } = useTelegram();
    const { isAuthenticated } = useAuth();

    const [loading, setLoading] = useState(true);
    const [game, setGame] = useState(null);
    const [error, setError] = useState(null);

    // Загрузка данных игры
    useEffect(() => {
        const fetchGameData = async () => {
            try {
                setLoading(true);
                const response = await api.game.get(id);
                setGame(response.data);
            } catch (err) {
                console.error('Ошибка при загрузке игры:', err);
                setError('Не удалось загрузить информацию об игре');
            } finally {
                setLoading(false);
            }
        };

        fetchGameData();
    }, [id]);

    // Настройка кнопок Telegram
    useEffect(() => {
        if (isReady && tg) {
            if (isAuthenticated && game) {
                tg.MainButton.setText('Присоединиться к игре');
                tg.MainButton.show();
                tg.MainButton.onClick(() => handleJoinGame());
            } else {
                tg.MainButton.hide();
            }

            tg.BackButton.show();
            tg.BackButton.onClick(() => navigate(-1));
        }

        return () => {
            if (isReady && tg) {
                tg.MainButton.hide();
                tg.BackButton.hide();
            }
        };
    }, [isReady, tg, isAuthenticated, game, navigate]);

    // Присоединение к игре
    const handleJoinGame = () => {
        navigate(`/join-game/${id}`);
    };

    if (loading) {
        return <LoadingScreen text="Загрузка игры..." />;
    }

    if (error) {
        return (
            <PageContainer>
                <div style={{ textAlign: 'center', margin: '40px 0' }}>
                    <p>{error}</p>
                    <Button onClick={() => navigate('/')}>Вернуться к играм</Button>
                </div>
            </PageContainer>
        );
    }

    if (!game) {
        return (
            <PageContainer>
                <div style={{ textAlign: 'center', margin: '40px 0' }}>
                    <p>Игра не найдена</p>
                    <Button onClick={() => navigate('/')}>Вернуться к играм</Button>
                </div>
            </PageContainer>
        );
    }

    return (
        <PageContainer>
            <PageTitle>{game.title}</PageTitle>

            <Card
                title="Информация об игре"
                badges={[
                    { text: game.difficulty, variant: 'info' },
                    { text: game.currency, variant: 'warning' }
                ]}
            >
                <GameInfo>
                    <InfoItem>
                        <InfoLabel>Длина слова</InfoLabel>
                        <InfoValue>{game.length} букв</InfoValue>
                    </InfoItem>

                    <InfoItem>
                        <InfoLabel>Попытки</InfoLabel>
                        <InfoValue>{game.max_tries}</InfoValue>
                    </InfoItem>

                    <InfoItem>
                        <InfoLabel>Ставка</InfoLabel>
                        <InfoValue>{game.min_bet} - {game.max_bet} {game.currency}</InfoValue>
                    </InfoItem>

                    <InfoItem>
                        <InfoLabel>Множитель</InfoLabel>
                        <InfoValue>x{game.reward_multiplier}</InfoValue>
                    </InfoItem>

                    <InfoItem>
                        <InfoLabel>Призовой фонд</InfoLabel>
                        <InfoValue>
                            {game.currency === 'TON'
                                ? formatCurrency(game.reward_pool_ton, 'TON')
                                : formatCurrency(game.reward_pool_usdt, 'USDT')}
                        </InfoValue>
                    </InfoItem>
                </GameInfo>
            </Card>

            <div style={{ marginTop: '24px' }}>
                <Button
                    variant="primary"
                    fullWidth
                    onClick={handleJoinGame}
                    disabled={!isAuthenticated}
                >
                    Присоединиться к игре
                </Button>

                {!isAuthenticated && (
                    <p style={{ textAlign: 'center', fontSize: '14px', color: 'var(--tg-theme-hint-color, #999999)' }}>
                        Для участия в игре необходимо авторизоваться
                    </p>
                )}
            </div>
        </PageContainer>
    );
};

export default GamePage; 