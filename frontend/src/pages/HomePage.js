import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { gameAPI, lobbyAPI } from '../api';
import { useAuth } from '../context/AuthContext';
import GameCard from '../components/GameCard';
import BetModal from '../components/BetModal';
import UserProfile from '../components/UserProfile';

const Container = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: 16px;
`;

const Header = styled.header`
  margin-bottom: 24px;
`;

const Title = styled.h1`
  font-size: 24px;
  margin: 0;
  color: #333;
`;

const Subtitle = styled.p`
  font-size: 16px;
  color: #666;
  margin: 8px 0 0;
`;

const TabContainer = styled.div`
  display: flex;
  margin: 24px 0 16px;
  border-bottom: 1px solid #ddd;
`;

const Tab = styled.button`
  padding: 12px 16px;
  background: none;
  border: none;
  font-size: 16px;
  font-weight: ${props => (props.active ? 'bold' : 'normal')};
  color: ${props => (props.active ? '#0077cc' : '#333')};
  cursor: pointer;
  position: relative;
  
  &::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 3px;
    background-color: ${props => (props.active ? '#0077cc' : 'transparent')};
  }
`;

const GamesList = styled.div`
  margin-top: 24px;
`;

const EmptyState = styled.div`
  text-align: center;
  padding: 40px 0;
  color: #666;
`;

const EmptyStateText = styled.p`
  font-size: 16px;
  margin: 16px 0;
`;

const LoadingIndicator = styled.div`
  text-align: center;
  padding: 40px 0;
  color: #666;
`;

const CreateGameButton = styled.button`
  background-color: #0077cc;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 12px 20px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  margin-top: 16px;
  width: 100%;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: #0066b3;
  }
`;

const ErrorMessage = styled.div`
  background-color: #ffebee;
  color: #c62828;
  padding: 12px;
  border-radius: 6px;
  margin: 16px 0;
  font-size: 14px;
`;

const HomePage = () => {
    const { user, isAuthenticated, loading: authLoading } = useAuth();
    const navigate = useNavigate();

    const [activeTab, setActiveTab] = useState('active');
    const [games, setGames] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [selectedGame, setSelectedGame] = useState(null);
    const [showBetModal, setShowBetModal] = useState(false);

    // Загрузка игр в зависимости от активной вкладки
    useEffect(() => {
        const fetchGames = async () => {
            try {
                setLoading(true);
                setError(null);

                let response;
                if (activeTab === 'active') {
                    response = await gameAPI.getActiveGames();
                } else if (activeTab === 'my') {
                    response = await gameAPI.getUserGames();
                }

                setGames(response.data || []);
            } catch (err) {
                console.error('Error fetching games:', err);
                setError('Не удалось загрузить игры. Пожалуйста, попробуйте позже.');
            } finally {
                setLoading(false);
            }
        };

        fetchGames();
    }, [activeTab]);

    // Обработка нажатия на кнопку "Играть"
    const handleJoinGame = (gameId) => {
        if (!isAuthenticated) {
            setError('Необходимо авторизоваться для участия в игре');
            return;
        }

        const game = games.find(g => g.id === gameId);
        if (game) {
            setSelectedGame(game);
            setShowBetModal(true);
        }
    };

    // Обработка нажатия на кнопку "Подробнее"
    const handleGameDetails = (gameId) => {
        navigate(`/games/${gameId}`);
    };

    // Обработка подтверждения ставки
    const handleBetSubmit = async (betAmount) => {
        try {
            setLoading(true);
            const response = await lobbyAPI.joinGame(selectedGame.id, betAmount);
            setShowBetModal(false);

            // Переходим на страницу лобби
            navigate(`/lobbies/${response.data.id}`);
        } catch (err) {
            console.error('Error joining game:', err);
            setError(err.response?.data?.error || 'Не удалось присоединиться к игре');
        } finally {
            setLoading(false);
        }
    };

    // Обработка нажатия на кнопку "Создать игру"
    const handleCreateGame = () => {
        navigate('/games/create');
    };

    // Обработка нажатия на кнопку "Пополнить баланс"
    const handleDeposit = () => {
        navigate('/deposit');
    };

    // Обработка нажатия на кнопку "Вывести средства"
    const handleWithdraw = () => {
        navigate('/withdraw');
    };

    // Проверка активного лобби
    useEffect(() => {
        const checkActiveLobby = async () => {
            if (!isAuthenticated) return;

            try {
                const response = await lobbyAPI.getActiveLobby();
                if (response.data) {
                    // Если есть активное лобби, перенаправляем на его страницу
                    navigate(`/lobbies/${response.data.id}`);
                }
            } catch (err) {
                // Если нет активного лобби, ничего не делаем
                console.log('No active lobby found');
            }
        };

        checkActiveLobby();
    }, [isAuthenticated, navigate]);

    if (authLoading) {
        return (
            <Container>
                <LoadingIndicator>Загрузка...</LoadingIndicator>
            </Container>
        );
    }

    return (
        <Container>
            {isAuthenticated && <UserProfile onDeposit={handleDeposit} onWithdraw={handleWithdraw} />}

            <Header>
                <Title>Wordle Game</Title>
                <Subtitle>Угадывай слова и выигрывай TON!</Subtitle>
            </Header>

            {error && <ErrorMessage>{error}</ErrorMessage>}

            <TabContainer>
                <Tab
                    active={activeTab === 'active'}
                    onClick={() => setActiveTab('active')}
                >
                    Активные игры
                </Tab>

                {isAuthenticated && (
                    <Tab
                        active={activeTab === 'my'}
                        onClick={() => setActiveTab('my')}
                    >
                        Мои игры
                    </Tab>
                )}
            </TabContainer>

            {loading ? (
                <LoadingIndicator>Загрузка игр...</LoadingIndicator>
            ) : (
                <GamesList>
                    {games.length > 0 ? (
                        games.map(game => (
                            <GameCard
                                key={game.id}
                                game={game}
                                onJoin={handleJoinGame}
                                onDetails={handleGameDetails}
                            />
                        ))
                    ) : (
                        <EmptyState>
                            <EmptyStateText>
                                {activeTab === 'active'
                                    ? 'Нет активных игр. Создайте свою игру!'
                                    : 'У вас пока нет созданных игр.'}
                            </EmptyStateText>
                            {isAuthenticated && (
                                <CreateGameButton onClick={handleCreateGame}>
                                    Создать игру
                                </CreateGameButton>
                            )}
                        </EmptyState>
                    )}

                    {activeTab === 'active' && games.length > 0 && isAuthenticated && (
                        <CreateGameButton onClick={handleCreateGame}>
                            Создать свою игру
                        </CreateGameButton>
                    )}
                </GamesList>
            )}

            {showBetModal && selectedGame && (
                <BetModal
                    game={selectedGame}
                    onClose={() => setShowBetModal(false)}
                    onSubmit={handleBetSubmit}
                    userBalance={user}
                />
            )}
        </Container>
    );
};

export default HomePage; 