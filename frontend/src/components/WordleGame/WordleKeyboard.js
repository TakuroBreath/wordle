import React from 'react';
import styled from 'styled-components';

// Контейнер для клавиатуры
const KeyboardContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
  padding: 10px 0;
`;

// Ряд клавиатуры
const KeyboardRow = styled.div`
  display: flex;
  justify-content: center;
  width: 100%;
  margin-bottom: 8px;
`;

// Стили для клавиши
const Key = styled.button`
  display: flex;
  justify-content: center;
  align-items: center;
  min-width: 20px;
  flex: ${props => props.flex || 1};
  height: 48px;
  margin: 0 3px;
  border-radius: 4px;
  border: none;
  background-color: ${props => {
        if (props.status === 'correct') return 'var(--tile-correct-color, #6aaa64)';
        if (props.status === 'present') return 'var(--tile-present-color, #c9b458)';
        if (props.status === 'absent') return 'var(--tile-absent-color, #787c7e)';
        return 'var(--tg-theme-secondary-bg-color, #e1e1e1)';
    }};
  color: ${props => {
        if (props.status) return 'white';
        return 'var(--tg-theme-text-color, #000000)';
    }};
  font-weight: bold;
  font-size: 14px;
  text-transform: uppercase;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
  
  &:active {
    opacity: 0.7;
  }
  
  @media (max-width: 400px) {
    height: 40px;
    font-size: 12px;
  }
`;

// Определение рядов клавиатуры
const firstRow = ['й', 'ц', 'у', 'к', 'е', 'н', 'г', 'ш', 'щ', 'з', 'х', 'ъ'];
const secondRow = ['ф', 'ы', 'в', 'а', 'п', 'р', 'о', 'л', 'д', 'ж', 'э'];
const thirdRow = ['Enter', 'я', 'ч', 'с', 'м', 'и', 'т', 'ь', 'б', 'ю', 'Backspace'];

/**
 * Компонент клавиатуры для Wordle
 * @param {Object} props - Свойства компонента
 * @param {Function} props.onKeyPress - Обработчик нажатия на клавишу
 * @param {Object} props.letterStatuses - Объект со статусами букв ('correct', 'present', 'absent')
 */
const WordleKeyboard = ({ onKeyPress, letterStatuses = {} }) => {
    const handleKeyClick = (key) => {
        onKeyPress(key);
    };

    // Функция создания ряда клавиатуры
    const createKeyboardRow = (keys) => {
        return (
            <KeyboardRow>
                {keys.map((key) => {
                    // Определяем размер специальных клавиш
                    let flex = 1;
                    if (key === 'Enter') flex = 1.5;
                    if (key === 'Backspace') flex = 1.5;

                    // Получаем статус буквы (если есть)
                    const status = letterStatuses[key.toLowerCase()];

                    return (
                        <Key
                            key={key}
                            flex={flex}
                            status={status}
                            onClick={() => handleKeyClick(key)}
                        >
                            {key === 'Backspace' ? '⌫' : key}
                        </Key>
                    );
                })}
            </KeyboardRow>
        );
    };

    // Глобальный обработчик нажатия клавиш на физической клавиатуре
    React.useEffect(() => {
        const handleKeyDown = (e) => {
            const key = e.key.toLowerCase();

            if (key === 'enter') {
                onKeyPress('Enter');
            } else if (key === 'backspace') {
                onKeyPress('Backspace');
            } else if (/^[а-яё]$/.test(key)) {
                onKeyPress(key);
            }
        };

        window.addEventListener('keydown', handleKeyDown);

        return () => {
            window.removeEventListener('keydown', handleKeyDown);
        };
    }, [onKeyPress]);

    return (
        <KeyboardContainer>
            {createKeyboardRow(firstRow)}
            {createKeyboardRow(secondRow)}
            {createKeyboardRow(thirdRow)}
        </KeyboardContainer>
    );
};

export default WordleKeyboard; 