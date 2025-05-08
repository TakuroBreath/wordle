import React from 'react';
import styled, { keyframes } from 'styled-components';

// Анимация вращения для спиннера
const spin = keyframes`
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
`;

// Контейнер для экрана загрузки
const LoadingContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  width: 100%;
  background-color: var(--tg-theme-bg-color, #ffffff);
`;

// Компонент спиннера
const Spinner = styled.div`
  width: 50px;
  height: 50px;
  border: 5px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.2));
  border-top: 5px solid var(--tg-theme-button-color, #50a8eb);
  border-radius: 50%;
  animation: ${spin} 1s linear infinite;
  margin-bottom: 20px;
`;

// Текст загрузки
const LoadingText = styled.p`
  font-size: 16px;
  color: var(--tg-theme-text-color, #000000);
  margin: 0;
  padding: 0;
`;

const LoadingScreen = ({ text = 'Загрузка...' }) => {
    return (
        <LoadingContainer>
            <Spinner />
            <LoadingText>{text}</LoadingText>
        </LoadingContainer>
    );
};

export default LoadingScreen; 