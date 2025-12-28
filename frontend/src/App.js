import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { AuthProvider, useAuth } from './context/AuthContext';
import { TonConnectProvider } from './context/TonConnectContext';
// import HomePage from './pages/HomePage'; // Currently unused
import GamePage from './pages/GamePage';
import LobbyPage from './pages/LobbyPage';
import CreateGamePage from './pages/CreateGamePage';
import DepositPage from './pages/DepositPage';
import WithdrawPage from './pages/WithdrawPage';
import MyGamesPage from './pages/MyGamesPage';
import AllGamesPage from './pages/AllGamesPage';
import ProfilePage from './pages/ProfilePage';
import BottomNavigation from './components/BottomNavigation';

// Глобальные стили
const GlobalStyle = styled.div`
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
    Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  background-color: #f5f5f5;
  min-height: 100vh;
  color: #333;
  padding-bottom: 70px; // Добавляем отступ для навигации
`;

// Компонент для защищенных маршрутов
const ProtectedRoute = ({ children }) => {
    const { isAuthenticated, loading } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        if (!loading && !isAuthenticated) {
            navigate('/all-games');
        }
    }, [loading, isAuthenticated, navigate]);

    if (loading) {
        return <div>Загрузка...</div>;
    }

    if (!isAuthenticated) {
        return null;
    }

    return children;
};

const AppRoutes = () => {
    const { loading } = useAuth();

    if (loading) {
        return <div>Загрузка...</div>;
    }

    return (
        <>
            <Routes>
                <Route path="/" element={<Navigate to="/all-games" />} />
                <Route path="/games/:id" element={<GamePage />} />
                <Route path="/lobbies/:id" element={<LobbyPage />} />

                {/* Основные маршруты с навигацией */}
                <Route path="/my-games" element={
                    <ProtectedRoute>
                        <MyGamesPage />
                    </ProtectedRoute>
                } />
                <Route path="/all-games" element={<AllGamesPage />} />
                <Route path="/profile" element={
                    <ProtectedRoute>
                        <ProfilePage />
                    </ProtectedRoute>
                } />

                {/* Защищенные маршруты */}
                <Route
                    path="/games/create"
                    element={
                        <ProtectedRoute>
                            <CreateGamePage />
                        </ProtectedRoute>
                    }
                />
                <Route
                    path="/deposit"
                    element={
                        <ProtectedRoute>
                            <DepositPage />
                        </ProtectedRoute>
                    }
                />
                <Route
                    path="/withdraw"
                    element={
                        <ProtectedRoute>
                            <WithdrawPage />
                        </ProtectedRoute>
                    }
                />

                {/* Маршрут по умолчанию - перенаправление на главную */}
                <Route path="*" element={<Navigate to="/all-games" />} />
            </Routes>
            <BottomNavigation />
        </>
    );
};

const App = () => {
    // Настройка Telegram Mini App
    useEffect(() => {
        const initTelegram = async () => {
            if (window.Telegram && window.Telegram.WebApp) {
                try {
                    window.Telegram.WebApp.ready();
                    window.Telegram.WebApp.expand();
                    
                    // Устанавливаем тему
                    const colorScheme = window.Telegram.WebApp.colorScheme;
                    document.documentElement.setAttribute('data-theme', colorScheme);
                    
                    // Получаем данные инициализации
                    const initData = window.Telegram.WebApp.initData;
                    if (initData) {
                        // Сохраняем данные в localStorage для использования в API
                        localStorage.setItem('telegram_init_data', initData);
                    }

                    // Настраиваем MainButton
                    window.Telegram.WebApp.MainButton.hide();
                    
                    // Обработка закрытия приложения
                    window.Telegram.WebApp.onEvent('viewportChanged', () => {
                        if (!window.Telegram.WebApp.isExpanded) {
                            window.Telegram.WebApp.expand();
                        }
                    });
                } catch (error) {
                    console.error('Error initializing Telegram WebApp:', error);
                }
            }
        };

        initTelegram();
    }, []);

    return (
        <Router>
            <TonConnectProvider>
                <AuthProvider>
                    <GlobalStyle>
                        <AppRoutes />
                    </GlobalStyle>
                </AuthProvider>
            </TonConnectProvider>
        </Router>
    );
};

export default App; 