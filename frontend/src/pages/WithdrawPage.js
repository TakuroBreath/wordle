import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { userAPI } from '../api';
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

const SuccessMessage = styled.div`
  background-color: #e8f5e9;
  color: #2e7d32;
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

const BalanceInfo = styled.div`
  background-color: #f5f5f5;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  font-size: 14px;
`;

const WithdrawPage = () => {
    const navigate = useNavigate();
    const { user, refreshUserData } = useAuth();

    const [formData, setFormData] = useState({
        amount: '',
        currency: 'TON'
    });

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(null);

    // Обработка изменения полей формы
    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    // Обработка отправки формы
    const handleSubmit = async (e) => {
        e.preventDefault();

        // Валидация формы
        if (!formData.amount || formData.amount <= 0) {
            setError('Пожалуйста, укажите корректную сумму');
            return;
        }

        const balance = formData.currency === 'TON' ? user.balance_ton : user.balance_usdt;
        if (parseFloat(formData.amount) > balance) {
            setError(`Недостаточно средств. Ваш баланс: ${balance} ${formData.currency}`);
            return;
        }

        try {
            setLoading(true);
            setError(null);
            setSuccess(null);

            // Отправляем запрос на вывод
            await userAPI.requestWithdraw(
                parseFloat(formData.amount),
                formData.currency
            );

            // Показываем информацию об успешном выводе
            setSuccess(`Запрос на вывод ${formData.amount} ${formData.currency} успешно создан. Средства будут отправлены в ближайшее время.`);

            // Очищаем поле суммы
            setFormData(prev => ({ ...prev, amount: '' }));

            // Обновляем данные пользователя
            refreshUserData();
        } catch (err) {
            console.error('Error requesting withdraw:', err);
            setError(err.response?.data?.error || 'Не удалось создать запрос на вывод средств');
        } finally {
            setLoading(false);
        }
    };

    // Возврат на главную страницу
    const handleBack = () => {
        navigate('/');
    };

    // Получаем текущий баланс пользователя
    const currentBalance = formData.currency === 'TON'
        ? parseFloat(user?.balance_ton || 0).toFixed(2)
        : parseFloat(user?.balance_usdt || 0).toFixed(2);

    return (
        <Container>
            <BackButton onClick={handleBack}>← Вернуться на главную</BackButton>

            <Header>
                <Title>Вывод средств</Title>
                <Subtitle>Выведите свои выигрыши</Subtitle>
            </Header>

            {error && <ErrorMessage>{error}</ErrorMessage>}
            {success && <SuccessMessage>{success}</SuccessMessage>}

            <Form onSubmit={handleSubmit}>
                <BalanceInfo>
                    Ваш текущий баланс: {currentBalance} {formData.currency}
                </BalanceInfo>

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
                    <Label htmlFor="amount">Сумма ({formData.currency})</Label>
                    <Input
                        type="number"
                        id="amount"
                        name="amount"
                        min="1"
                        step="0.1"
                        max={currentBalance}
                        value={formData.amount}
                        onChange={handleChange}
                    />
                    <HelperText>Минимальная сумма вывода: 1 {formData.currency}</HelperText>
                </FormGroup>

                <SubmitButton type="submit" disabled={loading}>
                    {loading ? 'Обработка...' : 'Вывести средства'}
                </SubmitButton>
            </Form>
        </Container>
    );
};

export default WithdrawPage; 