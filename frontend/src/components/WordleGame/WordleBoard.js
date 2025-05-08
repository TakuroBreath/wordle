import React from 'react';
import styled from 'styled-components';
import WordleRow from './WordleRow';

// Стили для игровой доски
const Board = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  gap: 6px;
  margin-bottom: 24px;
`;

/**
 * Компонент игровой доски Wordle
 * @param {Object} props - Свойства компонента
 * @param {Array} props.attempts - Массив предыдущих попыток
 * @param {number} props.maxTries - Максимальное количество попыток
 * @param {number} props.wordLength - Длина загаданного слова
 * @param {string} props.currentAttempt - Текущая попытка (введенные буквы)
 */
const WordleBoard = ({
    attempts = [],
    maxTries = 6,
    wordLength = 5,
    currentAttempt = ''
}) => {
    // Создаем массив строк для отображения
    const rows = [];

    // Добавляем уже сделанные попытки
    for (let i = 0; i < attempts.length; i++) {
        rows.push(
            <WordleRow
                key={`attempt-${i}`}
                word={attempts[i].word}
                result={attempts[i].result}
                wordLength={wordLength}
            />
        );
    }

    // Добавляем текущую попытку, если игра не закончена
    if (attempts.length < maxTries) {
        rows.push(
            <WordleRow
                key="current"
                word={currentAttempt}
                wordLength={wordLength}
                isActive={true}
            />
        );
    }

    // Добавляем пустые строки для будущих попыток
    for (let i = rows.length; i < maxTries; i++) {
        rows.push(
            <WordleRow
                key={`empty-${i}`}
                wordLength={wordLength}
            />
        );
    }

    return <Board>{rows}</Board>;
};

export default WordleBoard; 