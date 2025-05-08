import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { gameAPI } from '../api';

const Container = styled.div`
  padding: 20px;
  padding-bottom: 80px;
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
  cursor: pointer;
  
  &:hover {
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  }
`;

const AllGamesPage = () => {
    const [games, setGames] = useState([]);

    useEffect(() => {
        const fetchGames = async () => {
            try {
                const response = await gameAPI.getAllGames();
                setGames(response.data);
            } catch (error) {
                console.error('Error fetching games:', error);
            }
        };

        fetchGames();
    }, []);

    return (
        <Container>
            <h1>Все игры</h1>
            <GameList>
                {games.map(game => (
                    <GameCard
                        key={game.id}
                        onClick={() => window.location.href = `/games/${game.id}`}
                    >
                        <h3>{game.name}</h3>
                        <p>Создатель: {game.creator_name}</p>
                        <p>Статус: {game.status}</p>
                        <p>Участников: {game.participants_count}</p>
                    </GameCard>
                ))}
            </GameList>
        </Container>
    );
};

export default AllGamesPage; 