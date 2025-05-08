import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
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

// Профиль пользователя
const ProfileSection = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 24px;
`;

// Информация о пользователе
const UserInfo = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`;

// Имя пользователя
const UserName = styled.h2`
  font-size: 20px;
  font-weight: 600;
  margin: 0;
  color: var(--tg-theme-text-color, #000000);
`;

// Статистика пользователя
const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-top: 16px;
`;

// Элемент статистики
const StatItem = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px;
  background-color: var(--tg-theme-secondary-bg-color, #f1f1f1);
  border-radius: 8px;
`;

// Значение статистики
const StatValue = styled.span`
  font-size: 24px;
  font-weight: 700;
  color: var(--tg-theme-text-color, #000000);
`;

// Метка статистики
const StatLabel = styled.span`
  font-size: 12px;
  color: var(--tg-theme-hint-color, #999999);
  margin-top: 4px;
`;

// Секция баланса
const BalanceSection = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
`;

// Элемент баланса
const BalanceItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: var(--tg-theme-secondary-bg-color, #f1f1f1);
  border-radius: 8px;
`;

// Валюта баланса
const BalanceCurrency = styled.span`
  font-size: 16px;
  color: var(--tg-theme-text-color, #000000);
`;

// Значение баланса
const BalanceValue = styled.span`
  font-size: 18px;
  font-weight: 600;
  color: var(--tg-theme-text-color, #000000);
`;

// Форматирование денежных значений
const formatCurrency = (amount, currency) => {
    return `${amount.toFixed(2)} ${currency}`;
};

const ProfilePage = () => {
    const navigate = useNavigate();
    const { tg, isReady } = useTelegram();
    const { user, isAuthenticated, logout } = useAuth();

    const [myGames, setMyGames] = useState([]);
    const [loading, setLoading] = useState(true);

    // Проверка аутентификации
    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/');
        }
    }, [isAuthenticated, navigate]);

    // Загрузка данных пользователя
    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);

                // Получаем список игр пользователя
                const gamesResponse = await api.game.getMy();
                setMyGames(gamesResponse.data || []);
            } catch (err) {
                console.error('Ошибка при загрузке данных:', err);
            } finally {
                setLoading(false);
            }
        };

        if (isAuthenticated) {
            fetchData();
        }
    }, [isAuthenticated]);

    // Настройка кнопок Telegram
    useEffect(() => {
        if (isReady && tg) {
            tg.BackButton.show();
            tg.BackButton.onClick(() => navigate(-1));
        }

        return () => {
            if (isReady && tg) {
                tg.BackButton.hide();
            }
        };
    }, [isReady, tg, navigate]);

    // Обработка выхода из аккаунта
    const handleLogout = async () => {
        try {
            await logout();
            navigate('/');
        } catch (err) {
            console.error('Ошибка при выходе из аккаунта:', err);
        }
    };

    // Переход на страницу создания игры
    const handleCreateGame = () => {
        navigate('/create-game');
    };

    if (loading) {
        return <LoadingScreen text="Загрузка профиля..." />;
    }

    if (!user) {
        return (
            <PageContainer>
                <div style={{ textAlign: 'center', margin: '40px 0' }}>
                    <p>Вы не авторизованы</p>
                    <Button onClick={() => navigate('/')}>Вернуться на главную</Button>
                </div>
            </PageContainer>
        );
    }

    return (
        <PageContainer>
            <PageTitle>Профиль</PageTitle>

            <ProfileSection>
                <Card>
                    <UserInfo>
                        <UserName>{user.first_name} {user.last_name}</UserName>
                        <span style={{ color: 'var(--tg-theme-hint-color, #999999)' }}>
                            @{user.username || 'пользователь'}
                        </span>
                    </UserInfo>

                    <StatsGrid>
                        <StatItem>
                            <StatValue>{user.wins}</StatValue>
                            <StatLabel>Побед</StatLabel>
                        </StatItem>
                        <StatItem>
                            <StatValue>{user.losses}</StatValue>
                            <StatLabel>Поражений</StatLabel>
                        </StatItem>
                        <StatItem>
                            <StatValue>{((user.wins / (user.wins + user.losses)) * 100 || 0).toFixed(1)}%</StatValue>
                            <StatLabel>Процент побед</StatLabel>
                        </StatItem>
                        <StatItem>
                            <StatValue>{user.wins + user.losses}</StatValue>
                            <StatLabel>Всего игр</StatLabel>
                        </StatItem>
                    </StatsGrid>
                </Card>
            </ProfileSection>

            <Card title="Баланс">
                <BalanceSection>
                    <BalanceItem>
                        <BalanceCurrency>TON</BalanceCurrency>
                        <BalanceValue>{formatCurrency(user.balance_ton, '')}</BalanceValue>
                    </BalanceItem>
                    <BalanceItem>
                        <BalanceCurrency>USDT</BalanceCurrency>
                        <BalanceValue>{formatCurrency(user.balance_usdt, '')}</BalanceValue>
                    </BalanceItem>
                </BalanceSection>
            </Card>

            <Card
                title="Мои игры"
                footer={
                    <Button variant="primary" onClick={handleCreateGame}>
                        Создать новую игру
                    </Button>
                }
            >
                {myGames.length > 0 ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                        {myGames.map(game => (
                            <div
                                key={game.id}
                                style={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    padding: '8px',
                                    borderBottom: '1px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.1))'
                                }}
                            >
                                <div>
                                    <div style={{ fontWeight: 500 }}>{game.title}</div>
                                    <div style={{ fontSize: '12px', color: 'var(--tg-theme-hint-color, #999999)' }}>
                                        Слово: {game.length} букв, Статус: {game.status}
                                    </div>
                                </div>
                                <Button
                                    variant="secondary"
                                    onClick={() => navigate(`/game/${game.id}`)}
                                >
                                    Открыть
                                </Button>
                            </div>
                        ))}
                    </div>
                ) : (
                    <p style={{ textAlign: 'center', color: 'var(--tg-theme-hint-color, #999999)' }}>
                        У вас пока нет созданных игр
                    </p>
                )}
            </Card>

            <div style={{ marginTop: '24px' }}>
                <Button variant="secondary" fullWidth onClick={handleLogout}>
                    Выйти из аккаунта
                </Button>
            </div>
        </PageContainer>
    );
};

export default ProfilePage; 