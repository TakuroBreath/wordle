import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useTonConnect } from '../context/TonConnectContext';
import { userAPI } from '../api';

const Container = styled.div`
  padding: 20px;
  padding-bottom: 80px;
  max-width: 600px;
  margin: 0 auto;
`;

const ProfileCard = styled.div`
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  margin-top: 20px;
`;

const Avatar = styled.div`
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  margin: 0 auto 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: white;
  font-weight: bold;
`;

const UserName = styled.h2`
  text-align: center;
  margin: 0 0 8px;
  color: #333;
`;

const UserId = styled.p`
  text-align: center;
  color: #888;
  font-size: 14px;
  margin: 0 0 20px;
`;

const WalletSection = styled.div`
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  margin: 20px 0;
`;

const WalletLabel = styled.div`
  font-size: 12px;
  color: #666;
  margin-bottom: 8px;
  text-transform: uppercase;
  font-weight: 500;
`;

const WalletAddress = styled.div`
  font-family: monospace;
  font-size: 14px;
  color: #333;
  word-break: break-all;
  background: white;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #e0e0e0;
`;

const ConnectButton = styled.button`
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #0088CC 0%, #00AAFF 100%);
  color: white;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 136, 204, 0.3);
  }

  &:active {
    transform: translateY(0);
  }

  &:disabled {
    background: #ccc;
    cursor: not-allowed;
  }
`;

const DisconnectButton = styled(ConnectButton)`
  background: #ff4444;
  margin-top: 10px;
  
  &:hover {
    box-shadow: 0 4px 12px rgba(255, 68, 68, 0.3);
  }
`;

const BalanceSection = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin: 20px 0;
`;

const BalanceCard = styled.div`
  background: ${props => props.primary ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : '#f5f5f5'};
  color: ${props => props.primary ? 'white' : '#333'};
  padding: 16px;
  border-radius: 12px;
  text-align: center;
`;

const BalanceLabel = styled.div`
  font-size: 12px;
  opacity: 0.8;
  margin-bottom: 4px;
`;

const BalanceValue = styled.div`
  font-size: 20px;
  font-weight: 700;
`;

const Stats = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 20px;
`;

const StatItem = styled.div`
  background: #f5f5f5;
  padding: 16px 12px;
  border-radius: 8px;
  text-align: center;
`;

const StatLabel = styled.div`
  font-size: 11px;
  color: #666;
  margin-bottom: 4px;
  text-transform: uppercase;
`;

const StatValue = styled.div`
  font-size: 18px;
  font-weight: 600;
  color: #333;
`;

const ActionButtons = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-top: 20px;
`;

const ActionButton = styled.button`
  padding: 14px;
  border: none;
  border-radius: 8px;
  background: ${props => props.primary ? 'linear-gradient(135deg, #4CAF50 0%, #45a049 100%)' : '#e0e0e0'};
  color: ${props => props.primary ? 'white' : '#333'};
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-1px);
  }
`;

const ProfilePage = () => {
    const { user, refreshUserData } = useAuth();
    const { isConnected, address, connectWallet, disconnectWallet } = useTonConnect();
    const navigate = useNavigate();
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const response = await userAPI.getStats();
                setStats(response.data);
            } catch (error) {
                console.error('Ошибка загрузки статистики:', error);
            }
        };

        if (user) {
            fetchStats();
        }
    }, [user]);

    const handleConnectWallet = async () => {
        setLoading(true);
        try {
            await connectWallet();
            await refreshUserData();
        } catch (error) {
            console.error('Ошибка подключения кошелька:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleDisconnectWallet = async () => {
        setLoading(true);
        try {
            await disconnectWallet();
        } catch (error) {
            console.error('Ошибка отключения кошелька:', error);
        } finally {
            setLoading(false);
        }
    };

    if (!user) {
        return <Container>Загрузка...</Container>;
    }

    return (
        <Container>
            <ProfileCard>
                <Avatar>
                    {user.username ? user.username[0].toUpperCase() : user.first_name ? user.first_name[0].toUpperCase() : '?'}
                </Avatar>
                <UserName>
                    {user.first_name} {user.last_name}
                </UserName>
                <UserId>
                    @{user.username || 'user'} • ID: {user.telegram_id}
                </UserId>

                {/* Секция кошелька */}
                <WalletSection>
                    <WalletLabel>TON Кошелек</WalletLabel>
                    {isConnected && address ? (
                        <>
                            <WalletAddress>{address}</WalletAddress>
                            <DisconnectButton onClick={handleDisconnectWallet} disabled={loading}>
                                {loading ? 'Отключение...' : 'Отключить кошелек'}
                            </DisconnectButton>
                        </>
                    ) : user.wallet ? (
                        <>
                            <WalletAddress>{user.wallet}</WalletAddress>
                            <ConnectButton onClick={handleConnectWallet} disabled={loading}>
                                {loading ? 'Подключение...' : 'Переподключить кошелек'}
                            </ConnectButton>
                        </>
                    ) : (
                        <ConnectButton onClick={handleConnectWallet} disabled={loading}>
                            {loading ? 'Подключение...' : '🔗 Подключить TON Кошелек'}
                        </ConnectButton>
                    )}
                </WalletSection>

                {/* Балансы */}
                <BalanceSection>
                    <BalanceCard primary>
                        <BalanceLabel>TON</BalanceLabel>
                        <BalanceValue>{(user.balance_ton || 0).toFixed(4)}</BalanceValue>
                    </BalanceCard>
                    <BalanceCard>
                        <BalanceLabel>USDT</BalanceLabel>
                        <BalanceValue>{(user.balance_usdt || 0).toFixed(2)}</BalanceValue>
                    </BalanceCard>
                </BalanceSection>

                {/* Статистика */}
                <Stats>
                    <StatItem>
                        <StatLabel>Побед</StatLabel>
                        <StatValue>{stats?.wins || user.wins || 0}</StatValue>
                    </StatItem>
                    <StatItem>
                        <StatLabel>Поражений</StatLabel>
                        <StatValue>{stats?.losses || user.losses || 0}</StatValue>
                    </StatItem>
                    <StatItem>
                        <StatLabel>Игр</StatLabel>
                        <StatValue>{stats?.total_games || (user.wins || 0) + (user.losses || 0)}</StatValue>
                    </StatItem>
                </Stats>

                {/* Кнопки действий */}
                <ActionButtons>
                    <ActionButton primary onClick={() => navigate('/deposit')}>
                        💰 Пополнить
                    </ActionButton>
                    <ActionButton onClick={() => navigate('/withdraw')}>
                        💸 Вывести
                    </ActionButton>
                </ActionButtons>
            </ProfileCard>
        </Container>
    );
};

export default ProfilePage;
