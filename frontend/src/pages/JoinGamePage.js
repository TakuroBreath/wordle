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

// Форма
const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: 16px;
`;

// Группа формы
const FormGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`;

// Метка для поля
const Label = styled.label`
  font-size: 14px;
  font-weight: 500;
  color: var(--tg-theme-text-color, #000000);
`;

// Поле ввода
const Input = styled.input`
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.2));
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
  font-size: 16px;
  
  &:focus {
    outline: none;
    border-color: var(--tg-theme-button-color, #50a8eb);
  }
`;

// Подсказка для поля
const FieldHint = styled.p`
  font-size: 12px;
  color: var(--tg-theme-hint-color, #999999);
  margin: 0;
`;

// Группа кнопок
const ButtonGroup = styled.div`
  display: flex;
  gap: 12px;
  margin-top: 8px;
`;

const JoinGamePage = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const { tg, isReady } = useTelegram();
    const { isAuthenticated } = useAuth();

    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [game, setGame] = useState(null);
    const [betAmount, setBetAmount] = useState('');

    // Загрузка данных игры
    useEffect(() => {
        const fetchGameData = async () => {
            try {
                setLoading(true);
                const response = await api.game.get(id);
                setGame(response.data);
                setBetAmount(response.data.min_bet.toString());
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
            tg.BackButton.show();
            tg.BackButton.onClick(() => navigate(-1));
        }

        return () => {
            if (isReady && tg) {
                tg.BackButton.hide();
            }
        };
    }, [isReady, tg, navigate]);

    // Проверка авторизации при загрузке страницы
    useEffect(() => {
        if (!isAuthenticated) {
            console.log('Пользователь не авторизован, перенаправление на главную страницу');
            // Можно добавить сообщение о необходимости авторизации
            // setError('Для присоединения к игре необходимо авторизоваться');
        }
    }, [isAuthenticated]);

    // Присоединение к игре
    const handleSubmit = async (e) => {
        e.preventDefault();

        try {
            setLoading(true);
            setError(null);

            // Валидация данных
            const betAmountNum = parseFloat(betAmount);
            if (isNaN(betAmountNum)) {
                throw new Error('Некорректная сумма ставки');
            }

            if (betAmountNum < game.min_bet || betAmountNum > game.max_bet) {
                throw new Error(`Сумма ставки должна быть от ${game.min_bet} до ${game.max_bet} ${game.currency}`);
            }

            // Проверяем, что у нас есть все необходимые данные
            if (!id || !game) {
                throw new Error('Отсутствуют необходимые данные игры');
            }

            // Проверяем, что игра активна
            if (game.status !== 'active') {
                throw new Error('Игра неактивна');
            }

            const requestData = {
                game_id: id,
                bet_amount: betAmountNum
            };

            console.log('Подготовка данных для создания лобби:', {
                request: requestData,
                game: {
                    id: game.id,
                    title: game.title,
                    status: game.status,
                    min_bet: game.min_bet,
                    max_bet: game.max_bet,
                    currency: game.currency,
                    reward_multiplier: game.reward_multiplier
                }
            });

            // Проверяем авторизацию
            const token = localStorage.getItem('token');
            console.log('Токен авторизации:', token ? `${token.substring(0, 15)}...` : 'Отсутствует');

            // Проверяем заголовки API клиента
            const authHeader = api.apiClient?.defaults?.headers?.common?.['Authorization'];
            console.log('Заголовок авторизации в API клиенте:',
                authHeader ? `${authHeader.substring(0, 20)}...` : 'Отсутствует');

            // Если токен есть, но заголовок отсутствует, устанавливаем его
            if (token && !authHeader) {
                console.log('Установка токена в заголовки API перед запросом');
                api.setAuthToken(token);
            }

            // Создаем лобби
            console.log('Отправка запроса на создание лобби...');
            const response = await api.lobby.join(requestData);

            console.log('Ответ сервера при создании лобби:', {
                status: response.status,
                data: response.data,
                headers: response.headers
            });

            // Проверяем формат ответа
            const lobbyId = response.data?.id || response.data?.lobby?.id;
            if (lobbyId) {
                console.log('Лобби успешно создано, переход на страницу лобби:', lobbyId);
                // Переходим на страницу лобби
                navigate(`/lobby/${lobbyId}`);
            } else {
                console.error('Ошибка: отсутствует ID лобби в ответе сервера:', response.data);
                throw new Error('Не удалось создать лобби: отсутствует ID в ответе сервера');
            }
        } catch (err) {
            console.error('Ошибка при присоединении к игре:', err);
            let errorMessage = 'Не удалось присоединиться к игре. ';

            if (err.response?.data?.error) {
                console.error('Детали ошибки от сервера:', {
                    status: err.response.status,
                    statusText: err.response.statusText,
                    data: err.response.data,
                    url: err.response.config?.url,
                    method: err.response.config?.method
                });

                if (err.response.data.error.includes('active lobby not found')) {
                    errorMessage = 'Пожалуйста, попробуйте еще раз.';
                } else {
                    errorMessage += err.response.data.error;
                }
            } else if (err.message) {
                errorMessage += err.message;
            } else {
                errorMessage += 'Пожалуйста, попробуйте еще раз.';
            }

            setError(errorMessage);
            setLoading(false);
        }
    };

    if (loading) {
        return <LoadingScreen text="Загрузка игры..." />;
    }

    if (error) {
        return (
            <PageContainer>
                <div style={{ textAlign: 'center', margin: '40px 0' }}>
                    <p>{error}</p>
                    <Button onClick={() => navigate(-1)}>Назад</Button>
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
            <PageTitle>Присоединение к игре</PageTitle>

            <Card title="Информация об игре">
                <div style={{ marginBottom: '16px' }}>
                    <div style={{ fontWeight: 500, marginBottom: '8px' }}>{game.title}</div>
                    <div style={{ fontSize: '14px', color: 'var(--tg-theme-hint-color, #999999)' }}>
                        <div>Длина слова: {game.length} букв</div>
                        <div>Попыток: {game.max_tries}</div>
                        <div>Множитель: x{game.reward_multiplier}</div>
                    </div>
                </div>
            </Card>

            <Card title="Ваша ставка">
                <Form onSubmit={handleSubmit}>
                    <FormGroup>
                        <Label htmlFor="bet">Сумма ставки ({game.currency})</Label>
                        <Input
                            type="number"
                            id="bet"
                            value={betAmount}
                            onChange={(e) => setBetAmount(e.target.value)}
                            required
                            min={game.min_bet}
                            max={game.max_bet}
                            step="0.1"
                        />
                        <FieldHint>
                            Минимальная ставка: {game.min_bet} {game.currency}<br />
                            Максимальная ставка: {game.max_bet} {game.currency}<br />
                            Возможный выигрыш: {(parseFloat(betAmount) * game.reward_multiplier).toFixed(2)} {game.currency}
                        </FieldHint>
                    </FormGroup>

                    <ButtonGroup>
                        <Button type="button" variant="secondary" onClick={() => navigate(-1)}>
                            Отмена
                        </Button>
                        <Button type="submit" variant="primary">
                            Присоединиться
                        </Button>
                    </ButtonGroup>
                </Form>
            </Card>
        </PageContainer>
    );
};

export default JoinGamePage; 