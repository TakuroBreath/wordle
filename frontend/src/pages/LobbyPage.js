import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { useTelegram } from '../contexts/TelegramContext';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import WordleBoard from '../components/WordleGame/WordleBoard';
import WordleKeyboard from '../components/WordleGame/WordleKeyboard';
import Button from '../components/UI/Button';
import LoadingScreen from '../components/LoadingScreen';

// Контейнер страницы
const PageContainer = styled.div`
  padding: 16px;
  max-width: 600px;
  margin: 0 auto;
  width: 100%;
`;

// Шапка с информацией
const Header = styled.div`
  margin-bottom: 20px;
`;

// Заголовок игры
const GameTitle = styled.h1`
  font-size: 20px;
  font-weight: 700;
  margin: 0 0 8px 0;
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

// Блок с результатом игры
const ResultPanel = styled.div`
  background-color: ${props => props.success ? 'rgba(106, 170, 100, 0.2)' : 'rgba(231, 87, 87, 0.2)'};
  border-radius: 12px;
  padding: 16px;
  margin: 20px 0;
  text-align: center;
`;

// Текст результата
const ResultText = styled.p`
  font-size: 18px;
  font-weight: 600;
  color: ${props => props.success ? '#6aaa64' : '#e75757'};
  margin: 0 0 10px 0;
`;

// Текст награды
const RewardText = styled.p`
  font-size: 16px;
  margin: 0;
  color: var(--tg-theme-text-color, #000000);
`;

// Таймер до истечения времени лобби
const Timer = styled.div`
  font-size: 14px;
  color: ${props => props.warning ? '#ff3b30' : 'var(--tg-theme-hint-color, #999999)'};
  text-align: center;
  margin-bottom: 16px;
`;

// Панель с действиями
const ActionPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 20px;
`;

const LobbyPage = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const { tg, isReady } = useTelegram();
    const { user } = useAuth();

    // Состояния компонента
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [lobby, setLobby] = useState(null);
    const [game, setGame] = useState(null);
    const [attempts, setAttempts] = useState([]);
    const [currentAttempt, setCurrentAttempt] = useState('');
    const [letterStatuses, setLetterStatuses] = useState({});
    const [timeLeft, setTimeLeft] = useState(null);
    const [gameEnded, setGameEnded] = useState(false);
    const [success, setSuccess] = useState(false);

    // Загрузка данных лобби и игры
    useEffect(() => {
        const fetchLobbyData = async () => {
            try {
                setLoading(true);

                // Получаем данные о лобби
                const lobbyResponse = await api.lobby.get(id);
                const lobbyData = lobbyResponse.data;
                setLobby(lobbyData);

                // Получаем данные об игре
                const gameResponse = await api.game.get(lobbyData.game_id);
                const gameData = gameResponse.data;
                setGame(gameData);

                // Получаем историю попыток
                const attemptsResponse = await api.lobby.getAttempts(id);
                setAttempts(attemptsResponse.data || []);

                // Обновляем статусы букв на основе попыток
                updateLetterStatuses(attemptsResponse.data || []);

                // Проверяем статус лобби
                if (lobbyData.status === 'success') {
                    setGameEnded(true);
                    setSuccess(true);
                } else if (lobbyData.status === 'failed') {
                    setGameEnded(true);
                    setSuccess(false);
                }

                setLoading(false);
            } catch (err) {
                console.error('Ошибка при загрузке данных лобби:', err);
                setError('Не удалось загрузить данные игры');
                setLoading(false);
            }
        };

        fetchLobbyData();
    }, [id]);

    // Обновление статусов букв
    const updateLetterStatuses = (attempts) => {
        const statuses = {};

        attempts.forEach(attempt => {
            const word = attempt.word.toLowerCase();
            const result = attempt.result;

            for (let i = 0; i < word.length; i++) {
                const letter = word[i];
                const currentStatus = statuses[letter];
                const newStatus = result[i] === 2 ? 'correct' :
                    result[i] === 1 ? 'present' :
                        'absent';

                // Приоритет статусов: correct > present > absent
                if (currentStatus === 'correct') {
                    continue;
                } else if (currentStatus === 'present' && newStatus !== 'correct') {
                    continue;
                }

                statuses[letter] = newStatus;
            }
        });

        setLetterStatuses(statuses);
    };

    // Отслеживание времени до истечения лобби
    useEffect(() => {
        if (!lobby || gameEnded) return;

        const calculateTimeLeft = () => {
            const now = new Date();
            const expiresAt = new Date(lobby.expires_at);
            const difference = expiresAt - now;

            if (difference <= 0) {
                // Время истекло
                setTimeLeft(0);
                setGameEnded(true);
                setSuccess(false);
                return;
            }

            // Формат времени: MM:SS
            const minutes = Math.floor((difference / 1000 / 60) % 60);
            const seconds = Math.floor((difference / 1000) % 60);

            return {
                minutes,
                seconds,
                total: difference
            };
        };

        const timer = setInterval(() => {
            const timeLeftValue = calculateTimeLeft();
            if (!timeLeftValue) {
                clearInterval(timer);
            } else {
                setTimeLeft(timeLeftValue);
            }
        }, 1000);

        // Начальный расчет времени
        setTimeLeft(calculateTimeLeft());

        return () => clearInterval(timer);
    }, [lobby, gameEnded]);

    // Обработка нажатия на клавишу
    const handleKeyPress = useCallback((key) => {
        if (gameEnded || !game) return;

        if (key === 'Backspace') {
            // Удаляем последнюю букву
            setCurrentAttempt(prev => prev.slice(0, -1));
        } else if (key === 'Enter') {
            // Проверяем и отправляем попытку
            if (currentAttempt.length === game.length) {
                submitAttempt();
            }
        } else if (/^[а-яё]$/.test(key.toLowerCase())) {
            // Добавляем букву (если не достигнута максимальная длина)
            if (currentAttempt.length < game.length) {
                setCurrentAttempt(prev => prev + key.toLowerCase());
            }
        }
    }, [currentAttempt, game, gameEnded]);

    // Отправка попытки на сервер
    const submitAttempt = async () => {
        try {
            const response = await api.lobby.makeAttempt(id, { word: currentAttempt });

            // Добавляем новую попытку в список
            const newAttempt = {
                word: currentAttempt,
                result: response.data.result
            };

            setAttempts(prev => [...prev, newAttempt]);
            updateLetterStatuses([...attempts, newAttempt]);
            setCurrentAttempt('');

            // Проверка на победу или проигрыш
            if (response.data.status === 'success') {
                setGameEnded(true);
                setSuccess(true);
            } else if (response.data.status === 'failed' ||
                attempts.length + 1 >= lobby.max_tries) {
                setGameEnded(true);
                setSuccess(false);
            }
        } catch (err) {
            console.error('Ошибка при отправке попытки:', err);
            // Можно показать сообщение об ошибке
        }
    };

    // Обработка продления времени
    const handleExtendTime = async () => {
        try {
            await api.lobby.extendTime(id, { duration: 5 }); // Продление на 5 минут

            // Обновляем данные лобби для обновления времени
            const response = await api.lobby.get(id);
            setLobby(response.data);
        } catch (err) {
            console.error('Ошибка при продлении времени:', err);
        }
    };

    // Настройка кнопок Telegram
    useEffect(() => {
        if (!isReady || !tg) return;

        if (gameEnded) {
            // Если игра завершена, предлагаем вернуться к списку игр
            tg.MainButton.setText('Вернуться к играм');
            tg.MainButton.show();
            tg.MainButton.onClick(() => navigate('/'));
        } else {
            tg.MainButton.hide();
        }

        // Показываем кнопку "Назад"
        tg.BackButton.show();
        tg.BackButton.onClick(() => navigate(-1));

        return () => {
            tg.MainButton.hide();
            tg.BackButton.hide();
        };
    }, [isReady, tg, gameEnded, navigate]);

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

    return (
        <PageContainer>
            <Header>
                <GameTitle>{game.title}</GameTitle>

                <GameInfo>
                    <InfoItem>
                        <InfoLabel>Ставка</InfoLabel>
                        <InfoValue>{lobby.bet_amount} {game.currency}</InfoValue>
                    </InfoItem>

                    <InfoItem>
                        <InfoLabel>Возможный выигрыш</InfoLabel>
                        <InfoValue>{lobby.potential_reward} {game.currency}</InfoValue>
                    </InfoItem>

                    <InfoItem>
                        <InfoLabel>Попытки</InfoLabel>
                        <InfoValue>{attempts.length}/{lobby.max_tries}</InfoValue>
                    </InfoItem>
                </GameInfo>

                {!gameEnded && timeLeft && (
                    <Timer warning={timeLeft.total < 60000}>
                        Осталось времени: {timeLeft.minutes.toString().padStart(2, '0')}:
                        {timeLeft.seconds.toString().padStart(2, '0')}
                    </Timer>
                )}
            </Header>

            {/* Игровое поле */}
            <WordleBoard
                attempts={attempts}
                maxTries={lobby.max_tries}
                wordLength={game.length}
                currentAttempt={currentAttempt}
            />

            {/* Результат игры */}
            {gameEnded && (
                <ResultPanel success={success}>
                    <ResultText success={success}>
                        {success ? 'Вы выиграли!' : 'Вы проиграли!'}
                    </ResultText>
                    {success && (
                        <RewardText>
                            Ваш выигрыш: {lobby.potential_reward} {game.currency}
                        </RewardText>
                    )}
                    {!success && (
                        <RewardText>
                            Загаданное слово: {game.word.toUpperCase()}
                        </RewardText>
                    )}
                </ResultPanel>
            )}

            {/* Клавиатура */}
            {!gameEnded && (
                <>
                    <WordleKeyboard
                        onKeyPress={handleKeyPress}
                        letterStatuses={letterStatuses}
                    />

                    <ActionPanel>
                        {timeLeft && timeLeft.total < 300000 && ( // Менее 5 минут
                            <Button onClick={handleExtendTime}>
                                Продлить время (+5 минут)
                            </Button>
                        )}
                    </ActionPanel>
                </>
            )}

            {gameEnded && (
                <ActionPanel>
                    <Button onClick={() => navigate('/')}>
                        Вернуться к списку игр
                    </Button>
                </ActionPanel>
            )}
        </PageContainer>
    );
};

export default LobbyPage; 