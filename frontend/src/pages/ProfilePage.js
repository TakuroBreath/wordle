import React from 'react';
import styled from 'styled-components';
import { useAuth } from '../context/AuthContext';

const Container = styled.div`
  padding: 20px;
  padding-bottom: 80px;
`;

const ProfileCard = styled.div`
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-top: 20px;
`;

const Avatar = styled.div`
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #e0e0e0;
  margin: 0 auto 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: #666;
`;

const Stats = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-top: 20px;
`;

const StatItem = styled.div`
  background: #f5f5f5;
  padding: 16px;
  border-radius: 8px;
  text-align: center;
`;

const ProfilePage = () => {
    const { user } = useAuth();

    if (!user) {
        return <Container>Загрузка...</Container>;
    }

    return (
        <Container>
            <h1>Профиль</h1>
            <ProfileCard>
                <Avatar>
                    {user.username ? user.username[0].toUpperCase() : '?'}
                </Avatar>
                <h2>{user.username}</h2>
                <p>ID: {user.id}</p>

                <Stats>
                    <StatItem>
                        <h3>Баланс</h3>
                        <p>{user.balance || 0} ₽</p>
                    </StatItem>
                    <StatItem>
                        <h3>Игр создано</h3>
                        <p>{user.games_created || 0}</p>
                    </StatItem>
                    <StatItem>
                        <h3>Побед</h3>
                        <p>{user.games_won || 0}</p>
                    </StatItem>
                    <StatItem>
                        <h3>Участий</h3>
                        <p>{user.games_played || 0}</p>
                    </StatItem>
                </Stats>
            </ProfileCard>
        </Container>
    );
};

export default ProfilePage; 