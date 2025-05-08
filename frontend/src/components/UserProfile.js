import React from 'react';
import styled from 'styled-components';
import { useAuth } from '../context/AuthContext';

const ProfileCard = styled.div`
  background-color: #ffffff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 20px;
  margin-bottom: 20px;
`;

const UserInfo = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: 16px;
`;

const Avatar = styled.div`
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background-color: #0077cc;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: bold;
  margin-right: 16px;
`;

const UserDetails = styled.div`
  flex: 1;
`;

const Username = styled.h3`
  margin: 0 0 4px;
  font-size: 18px;
`;

const UserID = styled.p`
  margin: 0;
  font-size: 14px;
  color: #666;
`;

const BalanceSection = styled.div`
  margin-top: 16px;
`;

const BalanceTitle = styled.h4`
  margin: 0 0 12px;
  font-size: 16px;
  color: #333;
`;

const BalanceRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
  
  &:last-child {
    border-bottom: none;
  }
`;

const CurrencyName = styled.span`
  font-size: 14px;
  color: #333;
`;

const BalanceAmount = styled.span`
  font-size: 16px;
  font-weight: bold;
  color: #0077cc;
`;

const StatsSection = styled.div`
  margin-top: 16px;
`;

const StatsTitle = styled.h4`
  margin: 0 0 12px;
  font-size: 16px;
  color: #333;
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
`;

const StatCard = styled.div`
  background-color: #f9f9f9;
  border-radius: 8px;
  padding: 12px;
  text-align: center;
`;

const StatValue = styled.div`
  font-size: 20px;
  font-weight: bold;
  color: #333;
`;

const StatLabel = styled.div`
  font-size: 12px;
  color: #666;
  margin-top: 4px;
`;

const ActionButton = styled.button`
  background-color: #0077cc;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 10px 16px;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  margin-top: 16px;
  width: 100%;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: #0066b3;
  }
`;

const UserProfile = ({ onWithdraw, onDeposit }) => {
    const { user, refreshUserData } = useAuth();

    if (!user) {
        return <div>Загрузка профиля...</div>;
    }

    // Получаем первую букву имени пользователя для аватара
    const getInitial = () => {
        if (user.first_name && user.first_name.length > 0) {
            return user.first_name[0].toUpperCase();
        }
        if (user.username && user.username.length > 0) {
            return user.username[0].toUpperCase();
        }
        return 'U';
    };

    // Форматируем сумму с двумя знаками после запятой
    const formatBalance = (balance) => {
        return parseFloat(balance).toFixed(2);
    };

    return (
        <ProfileCard>
            <UserInfo>
                <Avatar>{getInitial()}</Avatar>
                <UserDetails>
                    <Username>{user.first_name} {user.last_name}</Username>
                    <UserID>@{user.username}</UserID>
                </UserDetails>
            </UserInfo>

            <BalanceSection>
                <BalanceTitle>Баланс</BalanceTitle>
                <BalanceRow>
                    <CurrencyName>TON</CurrencyName>
                    <BalanceAmount>{formatBalance(user.balance_ton)} TON</BalanceAmount>
                </BalanceRow>
                <BalanceRow>
                    <CurrencyName>USDT</CurrencyName>
                    <BalanceAmount>{formatBalance(user.balance_usdt)} USDT</BalanceAmount>
                </BalanceRow>
            </BalanceSection>

            <StatsSection>
                <StatsTitle>Статистика</StatsTitle>
                <StatsGrid>
                    <StatCard>
                        <StatValue>{user.wins || 0}</StatValue>
                        <StatLabel>Побед</StatLabel>
                    </StatCard>
                    <StatCard>
                        <StatValue>{user.losses || 0}</StatValue>
                        <StatLabel>Поражений</StatLabel>
                    </StatCard>
                    <StatCard>
                        <StatValue>{(user.wins && user.losses) ? ((user.wins / (user.wins + user.losses) * 100).toFixed(1) + '%') : '0%'}</StatValue>
                        <StatLabel>Винрейт</StatLabel>
                    </StatCard>
                    <StatCard>
                        <StatValue>{user.wins + user.losses || 0}</StatValue>
                        <StatLabel>Всего игр</StatLabel>
                    </StatCard>
                </StatsGrid>
            </StatsSection>

            <ActionButton onClick={onDeposit}>Пополнить баланс</ActionButton>
            <ActionButton onClick={onWithdraw} style={{ marginTop: '8px', backgroundColor: '#f0f0f0', color: '#333' }}>
                Вывести средства
            </ActionButton>
        </ProfileCard>
    );
};

export default UserProfile; 