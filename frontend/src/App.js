import React, { useEffect } from 'react';
import { Route, Routes, Navigate } from 'react-router-dom';
import { useTelegram } from './contexts/TelegramContext';
import { useAuth } from './contexts/AuthContext';
import styled from 'styled-components';

// Импорт компонентов страниц
import HomePage from './pages/HomePage';
import GamePage from './pages/GamePage';
import LobbyPage from './pages/LobbyPage';
import CreateGamePage from './pages/CreateGamePage';
import ProfilePage from './pages/ProfilePage';
import JoinGamePage from './pages/JoinGamePage';
import LoadingScreen from './components/LoadingScreen';

// Контейнер приложения со стилями, соответствующими Telegram Mini App
const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
`;

// Компонент для защищенных маршрутов, доступных только авторизованным пользователям
const ProtectedRoute = ({ children }) => {
    const { isAuthenticated, loading } = useAuth();

    if (loading) {
        return <LoadingScreen />;
    }

    return isAuthenticated ? children : <Navigate to="/" />;
};

function App() {
    const { tg, isReady } = useTelegram();
    const { loading } = useAuth();

    // Настройка обработчика кнопки "Назад" в Telegram
    useEffect(() => {
        if (isReady && tg) {
            // Скрываем кнопку "Назад" при загрузке приложения
            tg.BackButton.hide();

            // Функция обработки события нажатия кнопки "Назад"
            const handleBackButton = () => {
                window.history.back();
            };

            // Устанавливаем обработчик события
            tg.BackButton.onClick(handleBackButton);

            // Очистка обработчика при размонтировании компонента
            return () => {
                tg.BackButton.offClick(handleBackButton);
            };
        }
    }, [isReady, tg]);

    // Показываем загрузку, пока проверяем аутентификацию
    if (loading) {
        return <LoadingScreen />;
    }

    return (
        <AppContainer>
            <Routes>
                {/* Публичные маршруты */}
                <Route path="/" element={<HomePage />} />
                <Route path="/game/:id" element={<GamePage />} />

                {/* Защищенные маршруты */}
                <Route
                    path="/join-game/:id"
                    element={
                        <ProtectedRoute>
                            <JoinGamePage />
                        </ProtectedRoute>
                    }
                />
                <Route
                    path="/lobby/:id"
                    element={
                        <ProtectedRoute>
                            <LobbyPage />
                        </ProtectedRoute>
                    }
                />
                <Route
                    path="/create-game"
                    element={
                        <ProtectedRoute>
                            <CreateGamePage />
                        </ProtectedRoute>
                    }
                />
                <Route
                    path="/profile"
                    element={
                        <ProtectedRoute>
                            <ProfilePage />
                        </ProtectedRoute>
                    }
                />

                {/* Маршрут по умолчанию - редирект на главную */}
                <Route path="*" element={<Navigate to="/" />} />
            </Routes>
        </AppContainer>
    );
}

export default App; 