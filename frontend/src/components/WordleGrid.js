import React from 'react';
import styled from 'styled-components';

// Стили для сетки
const Grid = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  margin: 20px 0;
`;

const Row = styled.div`
  display: flex;
  gap: 5px;
`;

const Cell = styled.div`
  width: 50px;
  height: 50px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 24px;
  font-weight: bold;
  text-transform: uppercase;
  border: 2px solid #d3d6da;
  background-color: ${props => {
        if (props.state === 2) return '#6aaa64'; // Правильная буква и позиция
        if (props.state === 1) return '#c9b458'; // Буква есть, но не на той позиции
        if (props.state === 0) return '#787c7e'; // Буквы нет в слове
        return '#ffffff'; // Пустая ячейка
    }};
  color: ${props => (props.state !== undefined ? '#ffffff' : '#000000')};
  transition: all 0.2s ease;
  
  @media (max-width: 480px) {
    width: 40px;
    height: 40px;
    font-size: 20px;
  }
`;

// Компонент сетки Wordle
const WordleGrid = ({ attempts, wordLength, maxTries }) => {
    // Создаем пустую сетку с нужным количеством строк и столбцов
    const renderGrid = () => {
        const rows = [];

        // Заполняем сетку существующими попытками
        for (let i = 0; i < maxTries; i++) {
            const attempt = attempts[i] || { word: '', result: [] };
            const cells = [];

            for (let j = 0; j < wordLength; j++) {
                const letter = attempt.word[j] || '';
                const state = attempt.result[j];

                cells.push(
                    <Cell key={`cell-${i}-${j}`} state={state}>
                        {letter}
                    </Cell>
                );
            }

            rows.push(<Row key={`row-${i}`}>{cells}</Row>);
        }

        return rows;
    };

    return <Grid>{renderGrid()}</Grid>;
};

export default WordleGrid; 