import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { lobbyAPI } from '../api';
import WordleGrid from '../components/WordleGrid';
import Keyboard from '../components/Keyboard';

const Container = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: 16px;
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

const InfoBar = styled.div`
  display: flex;
  justify-content: space-between;
  background-color: #f9f9f9;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 24px;
`;

const InfoItem = styled.div`
  text-align: center;
`;

const InfoLabel = styled.div`
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
`;

const InfoValue = styled.div`
  font-size: 16px;
  font-weight: bold;
  color: #333;
`;

const TimerValue = styled(InfoValue)`
  color: ${props => props.isExpiring ? '#e53935' : '#333'};
`;

const GameSection = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24px;
`;

const InputRow = styled.div`
  display: flex;
  margin-bottom: 16px;
  width: 100%;
  max-width: 400px;
`;

const WordInput = styled.input`
  flex: 1;
  padding: 12px;
  font-size: 18px;
  border: 2px solid #d3d6da;
  border-radius: 6px;
  text-transform: uppercase;
  
  &:focus {
    outline: none;
    border-color: #0077cc;
  }
  
  &:disabled {
    background-color: #f9f9f9;
    cursor: not-allowed;
  }
`;

const SubmitButton = styled.button`
  background-color: #0077cc;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 12px 20px;
  font-size: 16px;
  font-weight: bold;
  margin-left: 8px;
  cursor: pointer;
  
  &:hover {
    background-color: #0066b3;
  }
  
  &:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
  }
`;

const ResultSection = styled.div`
  text-align: center;
  padding: 24px;
  border-radius: 12px;
  margin-bottom: 24px;
  background-color: ${props => props.success ? '#e8f5e9' : '#ffebee'};
`;

const ResultTitle = styled.h2`
  font-size: 24px;
  margin: 0 0 16px;
  color: ${props => props.success ? '#2e7d32' : '#c62828'};
`;

const ResultMessage = styled.p`
  font-size: 16px;
  margin: 0 0 16px;
  color: #666;
`;

const RewardAmount = styled.div`
  font-size: 24px;
  font-weight: bold;
  color: #0077cc;
  margin: 16px 0;
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
  margin-top: 16px;
  
  &:hover {
    background-color: #0066b3;
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

const LobbyPage = () => {
    const { id } = useParams();
    const navigate = useNavigate();

    const [lobby, setLobby] = useState(null);
    const [game, setGame] = useState(null);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState(null);
    const [currentWord, setCurrentWord] = useState('');
    const [timeLeft, setTimeLeft] = useState(0);
    const [attempts, setAttempts] = useState([]);
    const [letterStates, setLetterStates] = useState({});

    // Загрузка данных лобби
    const fetchLobbyData = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);

            const response = await lobbyAPI.getLobby(id);
            const lobbyData = response.data;

            setLobby(lobbyData);
            setGame(lobbyData.game);

            // Преобразуем попытки в нужный формат
            if (lobbyData.attempts && lobbyData.attempts.length > 0) {
                const formattedAttempts = lobbyData.attempts.map(attempt => ({
                    word: attempt.word,
                    result: attempt.result
                }));
                setAttempts(formattedAttempts);

                // Обновляем состояние клавиатуры
                updateKeyboardState(formattedAttempts);
            }

            // Вычисляем оставшееся время
            if (lobbyData.expires_at) {
                const expiresAt = new Date(lobbyData.expires_at).getTime();
                const now = Date.now();
                const remaining = Math.max(0, Math.floor((expiresAt - now) / 1000));
                setTimeLeft(remaining);
            }
        } catch (err) {
            console.error('Error fetching lobby:', err);
            setError('Не удалось загрузить данные лобби');
        } finally {
            setLoading(false);
        }
    }, [id]);

    // Обновление состояния клавиатуры
    const updateKeyboardState = (attemptsList) => {
        const states = {};

        // Проходим по всем попыткам и обновляем состояние каждой буквы
        attemptsList.forEach(attempt => {
            const { word, result } = attempt;

            for (let i = 0; i < word.length; i++) {
                const letter = word[i].toLowerCase();
                const currentState = states[letter];
                const newState = result[i];

                // Обновляем состояние буквы, если новое состояние лучше
                if (currentState === undefined || newState > currentState) {
                    states[letter] = newState;
                }
            }
        });

        setLetterStates(states);
    };

    // Первоначальная загрузка данных
    useEffect(() => {
        fetchLobbyData();
    }, [fetchLobbyData]);

    // Таймер для обратного отсчета
    useEffect(() => {
        if (!lobby || lobby.status !== 'active') return;

        const timer = setInterval(() => {
            setTimeLeft(prevTime => {
                if (prevTime <= 1) {
                    clearInterval(timer);
                    // Перезагружаем данные лобби, если время истекло
                    fetchLobbyData();
                    return 0;
                }
                return prevTime - 1;
            });
        }, 1000);

        return () => clearInterval(timer);
    }, [lobby, fetchLobbyData]);

    // Форматирование времени
    const formatTime = (seconds) => {
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = seconds % 60;
        return `${minutes}:${remainingSeconds < 10 ? '0' : ''}${remainingSeconds}`;
    };

    // Возврат на главную страницу
    const handleBack = () => {
        navigate('/');
    };

    // Обработка ввода буквы
    const handleKeyPress = (key) => {
        if (currentWord.length < (game?.length || 5)) {
            setCurrentWord(prev => prev + key);
        }
    };

    // Обработка удаления буквы
    const handleDelete = () => {
        setCurrentWord(prev => prev.slice(0, -1));
    };

    // Обработка отправки слова
    const handleSubmit = async () => {
        if (currentWord.length !== (game?.length || 5)) {
            setError(`Слово должно быть длиной ${game?.length || 5} букв`);
            return;
        }

        try {
            setSubmitting(true);
            setError(null);

            const response = await lobbyAPI.makeAttempt(id, currentWord);

            // Добавляем новую попытку
            const newAttempt = {
                word: currentWord,
                result: response.data
            };

            const updatedAttempts = [...attempts, newAttempt];
            setAttempts(updatedAttempts);
            updateKeyboardState(updatedAttempts);

            // Очищаем поле ввода
            setCurrentWord('');

            // Перезагружаем данные лобби
            fetchLobbyData();
        } catch (err) {
            console.error('Error submitting attempt:', err);
            setError(err.response?.data?.error || 'Не удалось отправить попытку');
        } finally {
            setSubmitting(false);
        }
    };

    // Продление времени лобби
    const handleExtendTime = async () => {
        try {
            setLoading(true);
            await lobbyAPI.extendLobbyTime(id);
            fetchLobbyData();
        } catch (err) {
            console.error('Error extending lobby time:', err);
            setError('Не удалось продлить время игры');
        } finally {
            setLoading(false);
        }
    };

    // Возврат на главную страницу
    const handleReturnHome = () => {
        navigate('/');
    };

    // Начать новую игру
    const handlePlayAgain = () => {
        navigate('/');
    };

    if (loading && !lobby) {
        return (
            <Container>
                <LoadingIndicator>Загрузка лобби...</LoadingIndicator>
            </Container>
        );
    }

    if (error && !lobby) {
        return (
            <Container>
                <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>
                <ErrorMessage>{error}</ErrorMessage>
            </Container>
        );
    }

    if (!lobby) {
        return (
            <Container>
                <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>
                <ErrorMessage>Лобби не найдено</ErrorMessage>
            </Container>
        );
    }

    // Проверяем, закончена ли игра
    const isGameFinished = lobby.status !== 'active';
    const isSuccess = lobby.status === 'success';
    const isExpired = lobby.status === 'failed_expired';
    const isOutOfTries = lobby.status === 'failed_tries';

    // Проверяем, мало ли времени осталось
    const isTimeExpiring = timeLeft < 60;

    return (
        <Container>
            <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>

            <Header>
                <Title>Игра "{game?.title || 'Wordle'}"</Title>
                <Subtitle>
                    {isGameFinished
                        ? (isSuccess
                            ? 'Поздравляем! Вы угадали слово!'
                            : 'Игра завершена')
                        : 'Угадайте слово'}
                </Subtitle>
            </Header>

            {error && <ErrorMessage>{error}</ErrorMessage>}

            <InfoBar>
                <InfoItem>
                    <InfoLabel>Попытки</InfoLabel>
                    <InfoValue>{lobby.tries_used} / {lobby.max_tries}</InfoValue>
                </InfoItem>
                <InfoItem>
                    <InfoLabel>Ставка</InfoLabel>
                    <InfoValue>{lobby.bet_amount} {game?.currency}</InfoValue>
                </InfoItem>
                <InfoItem>
                    <InfoLabel>Потенциальный выигрыш</InfoLabel>
                    <InfoValue>{lobby.potential_reward} {game?.currency}</InfoValue>
                </InfoItem>
                {!isGameFinished && (
                    <InfoItem>
                        <InfoLabel>Время</InfoLabel>
                        <TimerValue isExpiring={isTimeExpiring}>{formatTime(timeLeft)}</TimerValue>
                    </InfoItem>
                )}
            </InfoBar>

            <GameSection>
                <WordleGrid
                    attempts={attempts}
                    wordLength={game?.length || 5}
                    maxTries={lobby.max_tries}
                />

                {!isGameFinished && (
                    <>
                        <InputRow>
                            <WordInput
                                type="text"
                                value={currentWord}
                                onChange={(e) => setCurrentWord(e.target.value.toLowerCase())}
                                maxLength={game?.length || 5}
                                placeholder={`Введите слово (${game?.length || 5} букв)`}
                                disabled={submitting}
                            />
                            <SubmitButton
                                onClick={handleSubmit}
                                disabled={currentWord.length !== (game?.length || 5) || submitting}
                            >
                                Отправить
                            </SubmitButton>
                        </InputRow>

                        <Keyboard
                            onKeyPress={handleKeyPress}
                            onEnter={handleSubmit}
                            onDelete={handleDelete}
                            letterStates={letterStates}
                            disabled={submitting}
                        />

                        {timeLeft < 120 && timeLeft > 0 && (
                            <ActionButton onClick={handleExtendTime}>
                                Продлить время (+5 минут)
                            </ActionButton>
                        )}
                    </>
                )}
            </GameSection>

            {isGameFinished && (
                <ResultSection success={isSuccess}>
                    <ResultTitle success={isSuccess}>
                        {isSuccess ? 'Победа!' : 'Игра завершена'}
                    </ResultTitle>

                    <ResultMessage>
                        {isSuccess && 'Поздравляем! Вы угадали слово и выиграли награду!'}
                        {isExpired && 'Время игры истекло. Попробуйте еще раз!'}
                        {isOutOfTries && 'Вы использовали все попытки. Попробуйте еще раз!'}
                    </ResultMessage>

                    {isSuccess && (
                        <RewardAmount>
                            Выигрыш: {lobby.reward || lobby.potential_reward} {game?.currency}
                        </RewardAmount>
                    )}

                    <ActionButton onClick={handlePlayAgain}>
                        Играть снова
                    </ActionButton>

                    <ActionButton
                        onClick={handleReturnHome}
                        style={{ marginTop: '8px', backgroundColor: '#f0f0f0', color: '#333' }}
                    >
                        Вернуться на главную
                    </ActionButton>
                </ResultSection>
            )}
        </Container>
    );
};

export default LobbyPage; 