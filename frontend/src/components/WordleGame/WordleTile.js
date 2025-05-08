import React from 'react';
import styled, { css, keyframes } from 'styled-components';

// Анимация переворота плитки
const flip = keyframes`
  0% {
    transform: rotateX(0);
  }
  50% {
    transform: rotateX(90deg);
  }
  100% {
    transform: rotateX(0);
  }
`;

// Анимация "прыжка" при вводе буквы
const bounce = keyframes`
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
`;

// Определяем цвета для различных состояний плитки
const getBackgroundColor = (status) => {
    switch (status) {
        case 'correct':
            return 'var(--tile-correct-color, #6aaa64)';
        case 'present':
            return 'var(--tile-present-color, #c9b458)';
        case 'absent':
            return 'var(--tile-absent-color, #787c7e)';
        case 'filled':
            return 'transparent';
        default:
            return 'transparent';
    }
};

// Определяем цвет текста для различных состояний плитки
const getTextColor = (status) => {
    switch (status) {
        case 'correct':
        case 'present':
        case 'absent':
            return 'white';
        default:
            return 'var(--tg-theme-text-color, #000000)';
    }
};

// Определяем цвет границы для различных состояний плитки
const getBorderColor = (status) => {
    switch (status) {
        case 'correct':
        case 'present':
        case 'absent':
            return 'transparent';
        case 'filled':
            return 'var(--tg-theme-text-color, #000000)';
        default:
            return 'var(--tg-theme-hint-color, #cccccc)';
    }
};

// Стили для плитки
const Tile = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  width: 48px;
  height: 48px;
  font-size: 24px;
  font-weight: bold;
  text-transform: uppercase;
  border: 2px solid ${props => getBorderColor(props.status)};
  background-color: ${props => getBackgroundColor(props.status)};
  color: ${props => getTextColor(props.status)};
  transition: border-color 0.15s ease;
  
  ${props => props.status === 'filled' && css`
    animation: ${bounce} 0.1s ease-in-out;
  `}
  
  ${props => (props.status === 'correct' || props.status === 'present' || props.status === 'absent') && css`
    animation: ${flip} 0.5s ease;
  `}
  
  @media (max-width: 600px) {
    width: 40px;
    height: 40px;
    font-size: 20px;
  }
  
  @media (max-width: 350px) {
    width: 32px;
    height: 32px;
    font-size: 16px;
  }
`;

// Инициализация CSS переменных для цветов плиток
const initTileColors = () => {
    if (typeof document !== 'undefined') {
        // Устанавливаем цвета по умолчанию для плиток
        document.documentElement.style.setProperty('--tile-correct-color', '#6aaa64');
        document.documentElement.style.setProperty('--tile-present-color', '#c9b458');
        document.documentElement.style.setProperty('--tile-absent-color', '#787c7e');
    }
};

/**
 * Компонент плитки Wordle
 * @param {Object} props - Свойства компонента
 * @param {string} props.letter - Буква в плитке
 * @param {string} props.status - Статус плитки ('empty', 'filled', 'correct', 'present', 'absent')
 */
const WordleTile = ({ letter, status = 'empty' }) => {
    // Инициализация цветов при первом рендере
    React.useEffect(() => {
        initTileColors();
    }, []);

    return <Tile status={status}>{letter}</Tile>;
};

export default WordleTile; 