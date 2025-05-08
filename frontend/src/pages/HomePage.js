import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { useTelegram } from '../contexts/TelegramContext';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import Card from '../components/UI/Card';
import Button from '../components/UI/Button';
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

// Секция для кнопок действий
const ActionSection = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 8px;
`;

// Панель информации о пользователе
const UserInfoPanel = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  margin-bottom: 24px;
  background-color: var(--tg-theme-secondary-bg-color, #f1f1f1);
  border-radius: 12px;
`;

// Текст информации о пользователе
const UserInfoText = styled.p`
  margin: 0;
  font-size: 14px;
  color: var(--tg-theme-text-color, #000000);
`;

// Список игр
const GameList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
`;

// Сообщение об отсутствии игр
const EmptyMessage = styled.p`
  text-align: center;
  color: var(--tg-theme-hint-color, #999999);
  font-size: 16px;
  margin: 40px 0;
`;

// Форматирование денежных значений
const formatCurrency = (amount, currency) => {
    return `${amount.toFixed(2)} ${currency}`;
};

const HomePage = () => {
    const { tg, isReady } = useTelegram();
    const { isAuthenticated, user } = useAuth();
    const [games, setGames] = useState([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    // Загрузка списка активных игр
    useEffect(() => {
        const fetchGames = async () => {
            try {
                setLoading(true);
                const response = await api.game.getActive({ limit: 10, offset: 0 });
                setGames(response.data || []);
            } catch (err) {
                console.error('Ошибка при загрузке игр:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchGames();
    }, []);

    // Настройка главной кнопки Telegram
    useEffect(() => {
        if (isReady && tg) {
            if (isAuthenticated) {
                tg.MainButton.setText('Создать игру');
                tg.MainButton.show();
                tg.MainButton.onClick(() => navigate('/create-game'));
            } else {
                tg.MainButton.hide();
            }
        }

        return () => {
            if (isReady && tg) {
                tg.MainButton.hide();
            }
        };
    }, [isReady, tg, isAuthenticated, navigate]);

    // Переход на страницу игры
    const handleGameClick = (gameId) => {
        navigate(`/game/${gameId}`);
    };

    // Переход на страницу профиля
    const handleProfileClick = () => {
        navigate('/profile');
    };

    // Переход на страницу создания игры
    const handleCreateGameClick = () => {
        navigate('/create-game');
    };

    if (loading) {
        return <LoadingScreen text="Загрузка игр..." />;
    }

    return (
        <PageContainer>
            <PageTitle>Wordle Game</PageTitle>

            {isAuthenticated && (
                <UserInfoPanel>
                    <div>
                        <UserInfoText>Привет, {user.first_name}!</UserInfoText>
                        <UserInfoText>
                            Баланс: {formatCurrency(user.balance_ton, 'TON')} / {formatCurrency(user.balance_usdt, 'USDT')}
                        </UserInfoText>
                    </div>
                    <Button variant="secondary" onClick={handleProfileClick}>
                        Профиль
                    </Button>
                </UserInfoPanel>
            )}

            <ActionSection>
                <Button
                    variant="primary"
                    fullWidth
                    onClick={handleCreateGameClick}
                    disabled={!isAuthenticated}
                >
                    Создать игру
                </Button>
            </ActionSection>

            <GameList>
                {games.length > 0 ? (
                    games.map((game) => (
                        <Card
                            key={game.id}
                            title={game.title}
                            clickable
                            fullWidth
                            onClick={() => handleGameClick(game.id)}
                            badges={[
                                { text: game.difficulty, variant: 'info' },
                                { text: game.currency, variant: 'warning' }
                            ]}
                            footer={
                                <>
                                    <div>
                                        <div style={{ fontSize: '14px', marginBottom: '4px' }}>
                                            Ставка: {game.min_bet} - {game.max_bet} {game.currency}
                                        </div>
                                        <div style={{ fontSize: '14px', color: 'var(--tg-theme-hint-color, #999999)' }}>
                                            Множитель: x{game.reward_multiplier}
                                        </div>
                                    </div>
                                    <Button variant="primary">Играть</Button>
                                </>
                            }
                        >
                            <div style={{ fontSize: '14px' }}>
                                <div>Длина слова: {game.length} букв</div>
                                <div>Попыток: {game.max_tries}</div>
                                <div>Призовой фонд: {
                                    game.currency === 'TON'
                                        ? formatCurrency(game.reward_pool_ton, 'TON')
                                        : formatCurrency(game.reward_pool_usdt, 'USDT')
                                }</div>
                            </div>
                        </Card>
                    ))
                ) : (
                    <EmptyMessage>
                        Нет активных игр. Создайте свою первую игру!
                    </EmptyMessage>
                )}
            </GameList>
        </PageContainer>
    );
};

export default HomePage; 