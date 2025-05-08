import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { useTelegram } from '../contexts/TelegramContext';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import Button from '../components/UI/Button';
import Card from '../components/UI/Card';
import LoadingScreen from '../components/LoadingScreen';

// Контейнер страницы
const PageContainer = styled.div`
  padding: 16px;
  max-width: 600px;
  margin: 0 auto;
  width: 100%;
`;

// Заголовок страницы
const PageTitle = styled.h1`
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 16px 0;
  color: var(--tg-theme-text-color, #000000);
`;

// Форма
const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: 16px;
`;

// Группа формы
const FormGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`;

// Метка для поля
const Label = styled.label`
  font-size: 14px;
  font-weight: 500;
  color: var(--tg-theme-text-color, #000000);
`;

// Поле ввода
const Input = styled.input`
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.2));
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
  font-size: 16px;
  
  &:focus {
    outline: none;
    border-color: var(--tg-theme-button-color, #50a8eb);
  }
`;

// Выбор
const Select = styled.select`
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.2));
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
  font-size: 16px;
  
  &:focus {
    outline: none;
    border-color: var(--tg-theme-button-color, #50a8eb);
  }
`;

// Подсказка для поля
const FieldHint = styled.p`
  font-size: 12px;
  color: var(--tg-theme-hint-color, #999999);
  margin: 0;
`;

// Группа кнопок
const ButtonGroup = styled.div`
  display: flex;
  gap: 12px;
  margin-top: 8px;
`;

const CreateGamePage = () => {
    const navigate = useNavigate();
    const { tg, isReady } = useTelegram();
    const { isAuthenticated } = useAuth();

    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        word: '',
        title: '',
        description: '',
        difficulty: 'medium',
        max_tries: 6,
        min_bet: 1,
        max_bet: 10,
        reward_multiplier: 2,
        currency: 'TON'
    });

    // Проверка аутентификации
    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/');
        }
    }, [isAuthenticated, navigate]);

    // Настройка кнопок Telegram
    useEffect(() => {
        if (isReady && tg) {
            tg.BackButton.show();
            tg.BackButton.onClick(() => navigate(-1));
        }

        return () => {
            if (isReady && tg) {
                tg.BackButton.hide();
            }
        };
    }, [isReady, tg, navigate]);

    // Обработка изменения полей формы
    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({
            ...formData,
            [name]: value
        });
    };

    // Создание игры
    const handleSubmit = async (e) => {
        e.preventDefault();

        try {
            setLoading(true);

            // Определение длины слова
            const gameData = {
                ...formData,
                length: formData.word.length
            };

            const response = await api.game.create(gameData);

            // Перейти на страницу созданной игры
            navigate(`/game/${response.data.id}`);
        } catch (err) {
            console.error('Ошибка при создании игры:', err);
            alert('Не удалось создать игру. Пожалуйста, попробуйте еще раз.');
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return <LoadingScreen text="Создание игры..." />;
    }

    return (
        <PageContainer>
            <PageTitle>Создание игры</PageTitle>

            <Card title="Настройки игры">
                <Form onSubmit={handleSubmit}>
                    <FormGroup>
                        <Label htmlFor="word">Загаданное слово</Label>
                        <Input
                            type="text"
                            id="word"
                            name="word"
                            value={formData.word}
                            onChange={handleChange}
                            required
                            pattern="[а-яА-ЯёЁ]+"
                            placeholder="Введите слово на русском языке"
                        />
                        <FieldHint>Введите слово, которое игроки будут угадывать (только русские буквы)</FieldHint>
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="title">Название игры</Label>
                        <Input
                            type="text"
                            id="title"
                            name="title"
                            value={formData.title}
                            onChange={handleChange}
                            required
                            placeholder="Введите название игры"
                        />
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="description">Описание</Label>
                        <Input
                            type="text"
                            id="description"
                            name="description"
                            value={formData.description}
                            onChange={handleChange}
                            placeholder="Краткое описание игры"
                        />
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="difficulty">Сложность</Label>
                        <Select
                            id="difficulty"
                            name="difficulty"
                            value={formData.difficulty}
                            onChange={handleChange}
                            required
                        >
                            <option value="easy">Легкая</option>
                            <option value="medium">Средняя</option>
                            <option value="hard">Сложная</option>
                        </Select>
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="max_tries">Максимальное число попыток</Label>
                        <Input
                            type="number"
                            id="max_tries"
                            name="max_tries"
                            value={formData.max_tries}
                            onChange={handleChange}
                            required
                            min="1"
                            max="10"
                        />
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="currency">Валюта</Label>
                        <Select
                            id="currency"
                            name="currency"
                            value={formData.currency}
                            onChange={handleChange}
                            required
                        >
                            <option value="TON">TON</option>
                            <option value="USDT">USDT</option>
                        </Select>
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="min_bet">Минимальная ставка</Label>
                        <Input
                            type="number"
                            id="min_bet"
                            name="min_bet"
                            value={formData.min_bet}
                            onChange={handleChange}
                            required
                            min="0.1"
                            step="0.1"
                        />
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="max_bet">Максимальная ставка</Label>
                        <Input
                            type="number"
                            id="max_bet"
                            name="max_bet"
                            value={formData.max_bet}
                            onChange={handleChange}
                            required
                            min="0.1"
                            step="0.1"
                        />
                        <FieldHint>Максимальная ставка должна быть больше минимальной</FieldHint>
                    </FormGroup>

                    <FormGroup>
                        <Label htmlFor="reward_multiplier">Множитель награды</Label>
                        <Input
                            type="number"
                            id="reward_multiplier"
                            name="reward_multiplier"
                            value={formData.reward_multiplier}
                            onChange={handleChange}
                            required
                            min="1"
                            step="0.1"
                        />
                        <FieldHint>Множитель определяет, сколько получит игрок в случае победы</FieldHint>
                    </FormGroup>

                    <ButtonGroup>
                        <Button type="button" variant="secondary" onClick={() => navigate('/')}>
                            Отмена
                        </Button>
                        <Button type="submit" variant="primary">
                            Создать игру
                        </Button>
                    </ButtonGroup>
                </Form>
            </Card>
        </PageContainer>
    );
};

export default CreateGamePage; 