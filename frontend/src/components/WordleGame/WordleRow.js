import React from 'react';
import styled from 'styled-components';
import WordleTile from './WordleTile';

// Стили для строки Wordle
const Row = styled.div`
  display: flex;
  gap: 5px;
  margin-bottom: 5px;
`;

/**
 * Компонент строки Wordle
 * @param {Object} props - Свойства компонента
 * @param {string} props.word - Слово в строке (текущее или предыдущая попытка)
 * @param {Array} props.result - Результат проверки (массив чисел: 0, 1, 2)
 * @param {number} props.wordLength - Длина загаданного слова
 * @param {boolean} props.isActive - Флаг активности строки (для текущей попытки)
 */
const WordleRow = ({
    word = '',
    result = [],
    wordLength = 5,
    isActive = false
}) => {
    // Преобразуем слово в массив букв
    const letters = word.split('');

    // Создаем массив плиток для отображения
    const tiles = [];

    for (let i = 0; i < wordLength; i++) {
        // Определяем состояние плитки
        let status = 'empty';
        if (letters[i]) {
            if (isActive) {
                // Для текущей попытки все плитки с введенными буквами имеют статус "filled"
                status = 'filled';
            } else if (result[i] !== undefined) {
                // Для предыдущих попыток используем результат проверки
                status = result[i] === 0 ? 'absent' :
                    result[i] === 1 ? 'present' :
                        'correct';
            }
        }

        // Добавляем плитку в массив
        tiles.push(
            <WordleTile
                key={i}
                letter={letters[i] || ''}
                status={status}
            />
        );
    }

    return <Row>{tiles}</Row>;
};

export default WordleRow; 