import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { gameAPI } from '../api';

const Container = styled.div`
  padding: 20px;
  padding-bottom: 80px;
`;

const CreateButton = styled.button`
  position: fixed;
  bottom: 80px;
  right: 20px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #2196F3;
  color: white;
  border: none;
  box-shadow: 0 2px 5px rgba(0,0,0,0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  cursor: pointer;
  
  &:hover {
    background: #1976D2;
  }
`;

const GameList = styled.div`
  display: grid;
  gap: 16px;
  margin-top: 20px;
`;

const GameCard = styled.div`
  background: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
`;

const MyGamesPage = () => {
    const [games, setGames] = useState([]);

    useEffect(() => {
        const fetchGames = async () => {
            try {
                const response = await gameAPI.getMyGames();
                setGames(response.data);
            } catch (error) {
                console.error('Error fetching games:', error);
            }
        };

        fetchGames();
    }, []);

    return (
        <Container>
            <h1>Мои игры</h1>
            <GameList>
                {games.map(game => (
                    <GameCard key={game.id}>
                        <h3>{game.name}</h3>
                        <p>Статус: {game.status}</p>
                        <p>Участников: {game.participants_count}</p>
                    </GameCard>
                ))}
            </GameList>
            <CreateButton onClick={() => window.location.href = '/games/create'}>
                +
            </CreateButton>
        </Container>
    );
};

export default MyGamesPage; 