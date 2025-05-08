import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { gameAPI } from '../api';
import { useAuth } from '../context/AuthContext';

const Container = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: 16px;
`;

const Header = styled.div`
  margin-bottom: 24px;
`;

const BackButton = styled.button`
  background: none;
  border: none;
  color: #0077cc;
  font-size: 16px;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  margin-bottom: 16px;
  
  &:hover {
    text-decoration: underline;
  }
`;

const Title = styled.h1`
  font-size: 24px;
  margin: 0 0 8px;
  color: #333;
`;

const Subtitle = styled.p`
  font-size: 16px;
  color: #666;
  margin: 8px 0;
`;

const Form = styled.form`
  background-color: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
`;

const FormGroup = styled.div`
  margin-bottom: 20px;
`;

const Label = styled.label`
  display: block;
  font-weight: bold;
  margin-bottom: 8px;
  font-size: 14px;
  color: #333;
`;

const Input = styled.input`
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  
  &:focus {
    outline: none;
    border-color: #0077cc;
  }
`;

const TextArea = styled.textarea`
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  resize: vertical;
  min-height: 100px;
  
  &:focus {
    outline: none;
    border-color: #0077cc;
  }
`;

const Select = styled.select`
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  background-color: #fff;
  
  &:focus {
    outline: none;
    border-color: #0077cc;
  }
`;

const RadioGroup = styled.div`
  display: flex;
  gap: 16px;
  margin-top: 8px;
`;

const RadioLabel = styled.label`
  display: flex;
  align-items: center;
  cursor: pointer;
`;

const RadioInput = styled.input`
  margin-right: 8px;
`;

const RangeContainer = styled.div`
  margin-top: 8px;
`;

const RangeInput = styled.input`
  width: 100%;
  margin: 8px 0;
`;

const RangeLabels = styled.div`
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
`;

const SubmitButton = styled.button`
  background-color: #0077cc;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 14px 20px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  width: 100%;
  margin-top: 16px;
  
  &:hover {
    background-color: #0066b3;
  }
  
  &:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
  }
`;

const ErrorMessage = styled.div`
  background-color: #ffebee;
  color: #c62828;
  padding: 12px;
  border-radius: 6px;
  margin: 16px 0;
  font-size: 14px;
`;

const HelperText = styled.p`
  font-size: 12px;
  color: #666;
  margin-top: 4px;
`;

const CreateGamePage = () => {
    const navigate = useNavigate();
    const { user } = useAuth();

    const [formData, setFormData] = useState({
        title: '',
        description: '',
        word: '',
        difficulty: 'medium',
        max_tries: 6,
        min_bet: 1,
        max_bet: 10,
        reward_multiplier: 2,
        currency: 'TON',
        initial_deposit: 0
    });

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    // Обработка изменения полей формы
    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    // Обработка отправки формы
    const handleSubmit = async (e) => {
        e.preventDefault();

        // Валидация формы
        if (!formData.title.trim()) {
            setError('Пожалуйста, укажите название игры');
            return;
        }

        if (!formData.word.trim()) {
            setError('Пожалуйста, укажите слово для угадывания');
            return;
        }

        // Проверка длины слова (от 3 до 8 букв)
        if (formData.word.length < 3 || formData.word.length > 8) {
            setError('Длина слова должна быть от 3 до 8 букв');
            return;
        }

        // Проверка на кириллицу
        if (!/^[а-яё]+$/i.test(formData.word)) {
            setError('Слово должно содержать только русские буквы');
            return;
        }

        try {
            setLoading(true);
            setError(null);

            // Подготавливаем данные для отправки
            const gameData = {
                ...formData,
                word: formData.word.toLowerCase(),
                length: formData.word.length,
                min_bet: parseFloat(formData.min_bet),
                max_bet: parseFloat(formData.max_bet),
                reward_multiplier: parseFloat(formData.reward_multiplier),
                initial_deposit: parseFloat(formData.initial_deposit)
            };

            // Создаем игру
            const response = await gameAPI.createGame(gameData);

            // Если успешно, переходим на страницу созданной игры
            navigate(`/games/${response.data.id}`);
        } catch (err) {
            console.error('Error creating game:', err);
            setError(err.response?.data?.error || 'Не удалось создать игру');
        } finally {
            setLoading(false);
        }
    };

    // Возврат на главную страницу
    const handleBack = () => {
        navigate('/');
    };

    return (
        <Container>
            <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>

            <Header>
                <Title>Создание новой игры</Title>
                <Subtitle>Задайте параметры игры и загадайте слово</Subtitle>
            </Header>

            {error && <ErrorMessage>{error}</ErrorMessage>}

            <Form onSubmit={handleSubmit}>
                <FormGroup>
                    <Label htmlFor="title">Название игры</Label>
                    <Input
                        type="text"
                        id="title"
                        name="title"
                        value={formData.title}
                        onChange={handleChange}
                        placeholder="Введите название игры"
                        required
                    />
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="description">Описание (необязательно)</Label>
                    <TextArea
                        id="description"
                        name="description"
                        value={formData.description}
                        onChange={handleChange}
                        placeholder="Введите описание игры"
                    />
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="word">Загаданное слово</Label>
                    <Input
                        type="text"
                        id="word"
                        name="word"
                        value={formData.word}
                        onChange={handleChange}
                        placeholder="Введите слово для угадывания"
                        required
                    />
                    <HelperText>Только русские буквы, от 3 до 8 букв</HelperText>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="difficulty">Сложность</Label>
                    <Select
                        id="difficulty"
                        name="difficulty"
                        value={formData.difficulty}
                        onChange={handleChange}
                    >
                        <option value="easy">Легкая</option>
                        <option value="medium">Средняя</option>
                        <option value="hard">Сложная</option>
                    </Select>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="max_tries">Количество попыток</Label>
                    <Input
                        type="number"
                        id="max_tries"
                        name="max_tries"
                        min="3"
                        max="10"
                        value={formData.max_tries}
                        onChange={handleChange}
                    />
                </FormGroup>

                <FormGroup>
                    <Label>Валюта</Label>
                    <RadioGroup>
                        <RadioLabel>
                            <RadioInput
                                type="radio"
                                name="currency"
                                value="TON"
                                checked={formData.currency === 'TON'}
                                onChange={handleChange}
                            />
                            TON
                        </RadioLabel>
                        <RadioLabel>
                            <RadioInput
                                type="radio"
                                name="currency"
                                value="USDT"
                                checked={formData.currency === 'USDT'}
                                onChange={handleChange}
                            />
                            USDT
                        </RadioLabel>
                    </RadioGroup>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="min_bet">Минимальная ставка ({formData.currency})</Label>
                    <Input
                        type="number"
                        id="min_bet"
                        name="min_bet"
                        min="0.1"
                        step="0.1"
                        value={formData.min_bet}
                        onChange={handleChange}
                    />
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="max_bet">Максимальная ставка ({formData.currency})</Label>
                    <Input
                        type="number"
                        id="max_bet"
                        name="max_bet"
                        min={formData.min_bet}
                        step="0.1"
                        value={formData.max_bet}
                        onChange={handleChange}
                    />
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="reward_multiplier">Множитель награды</Label>
                    <RangeContainer>
                        <RangeInput
                            type="range"
                            id="reward_multiplier"
                            name="reward_multiplier"
                            min="1.1"
                            max="5"
                            step="0.1"
                            value={formData.reward_multiplier}
                            onChange={handleChange}
                        />
                        <RangeLabels>
                            <span>1.1x</span>
                            <span>{formData.reward_multiplier}x</span>
                            <span>5x</span>
                        </RangeLabels>
                    </RangeContainer>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="initial_deposit">Начальный депозит в призовой фонд ({formData.currency})</Label>
                    <Input
                        type="number"
                        id="initial_deposit"
                        name="initial_deposit"
                        min="0"
                        step="0.1"
                        value={formData.initial_deposit}
                        onChange={handleChange}
                    />
                    <HelperText>
                        Чем больше начальный депозит, тем привлекательнее будет игра для участников
                    </HelperText>
                </FormGroup>

                <SubmitButton type="submit" disabled={loading}>
                    {loading ? 'Создание...' : 'Создать игру'}
                </SubmitButton>
            </Form>
        </Container>
    );
};

export default CreateGamePage; 