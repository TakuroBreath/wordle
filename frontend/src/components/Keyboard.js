import React from 'react';
import styled from 'styled-components';

// Стили для клавиатуры
const KeyboardContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin: 20px 0;
  width: 100%;
  max-width: 500px;
`;

const KeyboardRow = styled.div`
  display: flex;
  justify-content: center;
  gap: 6px;
  width: 100%;
`;

const Key = styled.button`
  min-width: 30px;
  height: 58px;
  border-radius: 4px;
  border: none;
  background-color: ${props => {
        if (props.state === 2) return '#6aaa64'; // Правильная буква и позиция
        if (props.state === 1) return '#c9b458'; // Буква есть, но не на той позиции
        if (props.state === 0) return '#787c7e'; // Буквы нет в слове
        return '#d3d6da'; // Обычная клавиша
    }};
  color: ${props => (props.state !== undefined ? '#ffffff' : '#000000')};
  font-weight: bold;
  font-size: 14px;
  cursor: pointer;
  text-transform: uppercase;
  flex: ${props => (props.wide ? '1.5' : '1')};
  
  &:hover {
    opacity: 0.8;
  }
  
  &:active {
    opacity: 0.6;
  }
  
  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  
  @media (max-width: 480px) {
    height: 50px;
    min-width: 20px;
    font-size: 12px;
  }
`;

// Компонент клавиатуры
const Keyboard = ({ onKeyPress, onEnter, onDelete, letterStates, disabled }) => {
    // Ряды клавиатуры (русская раскладка)
    const rows = [
        ['й', 'ц', 'у', 'к', 'е', 'н', 'г', 'ш', 'щ', 'з', 'х', 'ъ'],
        ['ф', 'ы', 'в', 'а', 'п', 'р', 'о', 'л', 'д', 'ж', 'э'],
        ['я', 'ч', 'с', 'м', 'и', 'т', 'ь', 'б', 'ю']
    ];

    const handleKeyClick = (key) => {
        if (disabled) return;
        onKeyPress(key);
    };

    const handleEnterClick = () => {
        if (disabled) return;
        onEnter();
    };

    const handleDeleteClick = () => {
        if (disabled) return;
        onDelete();
    };

    return (
        <KeyboardContainer>
            {rows.map((row, rowIndex) => (
                <KeyboardRow key={`row-${rowIndex}`}>
                    {rowIndex === 2 && (
                        <Key wide onClick={handleEnterClick} disabled={disabled}>
                            Ввод
                        </Key>
                    )}

                    {row.map((key) => (
                        <Key
                            key={key}
                            onClick={() => handleKeyClick(key)}
                            state={letterStates[key]}
                            disabled={disabled}
                        >
                            {key}
                        </Key>
                    ))}

                    {rowIndex === 2 && (
                        <Key wide onClick={handleDeleteClick} disabled={disabled}>
                            ⌫
                        </Key>
                    )}
                </KeyboardRow>
            ))}
        </KeyboardContainer>
    );
};

export default Keyboard; 