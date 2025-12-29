import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { gameAPI } from '../api';
import { useTonConnect } from '../context/TonConnectContext';

const Container = styled.div`
  max-width: 600px;
  margin: 0 auto;
  padding: 16px;
  padding-bottom: 100px;
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
  box-sizing: border-box;
  
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
  min-height: 80px;
  box-sizing: border-box;
  
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
  box-sizing: border-box;
  
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
  padding: 12px 20px;
  border: 2px solid ${props => props.checked ? '#0088CC' : '#ddd'};
  border-radius: 8px;
  background: ${props => props.checked ? '#e6f4ff' : 'white'};
  transition: all 0.2s;
  flex: 1;
  justify-content: center;
  font-weight: ${props => props.checked ? '600' : '400'};
`;

const RadioInput = styled.input`
  display: none;
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
  background: linear-gradient(135deg, #0088CC 0%, #00AAFF 100%);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 16px 20px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  width: 100%;
  margin-top: 16px;
  transition: transform 0.2s, box-shadow 0.2s;
  
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 136, 204, 0.3);
  }
  
  &:disabled {
    background: #cccccc;
    cursor: not-allowed;
    transform: none;
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

const DepositInfo = styled.div`
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
  border-radius: 12px;
  padding: 20px;
  margin: 24px 0;
  border: 1px solid #e0e0e0;
`;

const DepositTitle = styled.h3`
  margin: 0 0 12px;
  color: #333;
  font-size: 16px;
`;

const DepositAmount = styled.div`
  font-size: 28px;
  font-weight: 700;
  color: #0088CC;
  margin: 8px 0;
`;

const DepositDescription = styled.p`
  font-size: 13px;
  color: #666;
  margin: 8px 0 0;
  line-height: 1.5;
`;

const PaymentModal = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
`;

const PaymentCard = styled.div`
  background: white;
  border-radius: 16px;
  padding: 24px;
  max-width: 400px;
  width: 100%;
  text-align: center;
`;

const PaymentTitle = styled.h2`
  margin: 0 0 16px;
  color: #333;
`;

const PaymentAmount = styled.div`
  font-size: 32px;
  font-weight: 700;
  color: #0088CC;
  margin: 16px 0;
`;

const PaymentAddress = styled.div`
  background: #f5f5f5;
  border-radius: 8px;
  padding: 12px;
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  margin: 16px 0;
`;

const PaymentComment = styled.div`
  background: #fff3e0;
  border: 1px solid #ffcc80;
  border-radius: 8px;
  padding: 12px;
  margin: 16px 0;
  
  strong {
    display: block;
    color: #e65100;
    margin-bottom: 4px;
  }
  
  code {
    font-family: monospace;
    font-size: 16px;
    color: #333;
  }
`;

const PaymentButton = styled.button`
  background: linear-gradient(135deg, #0088CC 0%, #00AAFF 100%);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 14px 24px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  width: 100%;
  margin-top: 16px;
`;

const CancelButton = styled.button`
  background: #e0e0e0;
  color: #333;
  border: none;
  border-radius: 8px;
  padding: 12px 24px;
  font-size: 14px;
  cursor: pointer;
  width: 100%;
  margin-top: 12px;
`;

const CreateGamePage = () => {
    const navigate = useNavigate();
    const { isConnected, sendTransaction, getPaymentDeepLink } = useTonConnect();

    const [formData, setFormData] = useState({
        title: '',
        description: '',
        word: '',
        difficulty: 'medium',
        max_tries: 6,
        time_limit: 5, // минуты
        min_bet: 0.1,
        max_bet: 1,
        reward_multiplier: 2,
        currency: 'TON',
    });

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [paymentInfo, setPaymentInfo] = useState(null);
    const [createdGame, setCreatedGame] = useState(null);

    // Расчет требуемого депозита
    const requiredDeposit = useMemo(() => {
        return parseFloat(formData.max_bet) * parseFloat(formData.reward_multiplier);
    }, [formData.max_bet, formData.reward_multiplier]);

    // Обработка изменения полей формы
    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    // Валидация формы
    const validateForm = () => {
        if (!formData.title.trim()) {
            setError('Пожалуйста, укажите название игры');
            return false;
        }

        if (!formData.word.trim()) {
            setError('Пожалуйста, укажите слово для угадывания');
            return false;
        }

        if (formData.word.length < 3 || formData.word.length > 8) {
            setError('Длина слова должна быть от 3 до 8 букв');
            return false;
        }

        if (!/^[а-яё]+$/i.test(formData.word)) {
            setError('Слово должно содержать только русские буквы');
            return false;
        }

        if (parseFloat(formData.min_bet) <= 0) {
            setError('Минимальная ставка должна быть больше 0');
            return false;
        }

        if (parseFloat(formData.max_bet) < parseFloat(formData.min_bet)) {
            setError('Максимальная ставка не может быть меньше минимальной');
            return false;
        }

        return true;
    };

    // Обработка отправки формы
    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!validateForm()) return;

        try {
            setLoading(true);
            setError(null);

            // Подготавливаем данные для отправки
            const gameData = {
                title: formData.title,
                description: formData.description,
                word: formData.word.toLowerCase(),
                length: formData.word.length,
                difficulty: formData.difficulty,
                max_tries: parseInt(formData.max_tries),
                time_limit: parseInt(formData.time_limit),
                min_bet: parseFloat(formData.min_bet),
                max_bet: parseFloat(formData.max_bet),
                reward_multiplier: parseFloat(formData.reward_multiplier),
                currency: formData.currency,
            };

            // Создаем игру
            const response = await gameAPI.createGame(gameData);
            const game = response.data;

            setCreatedGame(game);

            // Если есть информация о платеже, показываем модал оплаты
            if (game.payment_info) {
                setPaymentInfo(game.payment_info);
            } else {
                // Если депозит не требуется, переходим на страницу игры
                navigate(`/games/${game.id}`);
            }
        } catch (err) {
            console.error('Error creating game:', err);
            setError(err.response?.data?.error || 'Не удалось создать игру');
        } finally {
            setLoading(false);
        }
    };

    // Оплата через TON Connect
    const handleTonConnectPayment = async () => {
        if (!paymentInfo || !isConnected) return;

        try {
            setLoading(true);
            await sendTransaction(
                paymentInfo.address,
                paymentInfo.amount,
                paymentInfo.comment
            );
            
            // После успешной отправки транзакции переходим на страницу игры
            navigate(`/games/${createdGame.id}`, { 
                state: { paymentPending: true } 
            });
        } catch (err) {
            console.error('Payment error:', err);
            setError('Ошибка при оплате. Попробуйте еще раз.');
        } finally {
            setLoading(false);
        }
    };

    // Оплата через deep link
    const handleDeepLinkPayment = () => {
        if (!paymentInfo) return;
        
        const deepLink = getPaymentDeepLink(
            paymentInfo.address,
            paymentInfo.amount,
            paymentInfo.comment
        );
        
        window.open(deepLink, '_blank');
        
        // Переходим на страницу игры
        navigate(`/games/${createdGame.id}`, { 
            state: { paymentPending: true } 
        });
    };

    // Закрытие модала оплаты
    const handleCancelPayment = () => {
        setPaymentInfo(null);
        setCreatedGame(null);
    };

    // Возврат на главную страницу
    const handleBack = () => {
        navigate('/my-games');
    };

    return (
        <Container>
            <BackButton onClick={handleBack}>← Мои игры</BackButton>

            <Header>
                <Title>Создание игры</Title>
                <Subtitle>Загадайте слово и установите условия</Subtitle>
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
                        placeholder="Например: Угадай слово"
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
                        placeholder="Подсказка или описание игры"
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
                        placeholder="Введите слово"
                        required
                    />
                    <HelperText>Только русские буквы, от 3 до 8 букв</HelperText>
                </FormGroup>

                <FormGroup>
                    <Label>Валюта</Label>
                    <RadioGroup>
                        <RadioLabel checked={formData.currency === 'TON'}>
                            <RadioInput
                                type="radio"
                                name="currency"
                                value="TON"
                                checked={formData.currency === 'TON'}
                                onChange={handleChange}
                            />
                            💎 TON
                        </RadioLabel>
                        <RadioLabel checked={formData.currency === 'USDT'}>
                            <RadioInput
                                type="radio"
                                name="currency"
                                value="USDT"
                                checked={formData.currency === 'USDT'}
                                onChange={handleChange}
                            />
                            💵 USDT
                        </RadioLabel>
                    </RadioGroup>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="max_tries">Количество попыток</Label>
                    <Select
                        id="max_tries"
                        name="max_tries"
                        value={formData.max_tries}
                        onChange={handleChange}
                    >
                        {[3, 4, 5, 6, 7, 8, 9, 10].map(n => (
                            <option key={n} value={n}>{n} попыток</option>
                        ))}
                    </Select>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="time_limit">Время на игру</Label>
                    <Select
                        id="time_limit"
                        name="time_limit"
                        value={formData.time_limit}
                        onChange={handleChange}
                    >
                        {[1, 2, 3, 5, 10, 15, 30].map(n => (
                            <option key={n} value={n}>{n} мин</option>
                        ))}
                    </Select>
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="min_bet">Минимальная ставка ({formData.currency})</Label>
                    <Input
                        type="number"
                        id="min_bet"
                        name="min_bet"
                        min="0.01"
                        step="0.01"
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
                        step="0.01"
                        value={formData.max_bet}
                        onChange={handleChange}
                    />
                </FormGroup>

                <FormGroup>
                    <Label htmlFor="reward_multiplier">Множитель награды: {formData.reward_multiplier}x</Label>
                    <RangeContainer>
                        <RangeInput
                            type="range"
                            id="reward_multiplier"
                            name="reward_multiplier"
                            min="1.5"
                            max="5"
                            step="0.1"
                            value={formData.reward_multiplier}
                            onChange={handleChange}
                        />
                        <RangeLabels>
                            <span>1.5x</span>
                            <span>5x</span>
                        </RangeLabels>
                    </RangeContainer>
                    <HelperText>
                        При выигрыше игрок получит ставку × {formData.reward_multiplier}
                    </HelperText>
                </FormGroup>

                <DepositInfo>
                    <DepositTitle>💰 Требуемый депозит</DepositTitle>
                    <DepositAmount>
                        {requiredDeposit.toFixed(4)} {formData.currency}
                    </DepositAmount>
                    <DepositDescription>
                        Депозит = максимальная ставка × множитель награды.
                        Это гарантирует выплату победителю.
                        После создания игры вам нужно будет оплатить депозит.
                    </DepositDescription>
                </DepositInfo>

                <SubmitButton type="submit" disabled={loading}>
                    {loading ? 'Создание...' : `Создать игру (${requiredDeposit.toFixed(4)} ${formData.currency})`}
                </SubmitButton>
            </Form>

            {/* Модал оплаты */}
            {paymentInfo && (
                <PaymentModal onClick={handleCancelPayment}>
                    <PaymentCard onClick={e => e.stopPropagation()}>
                        <PaymentTitle>💎 Оплата депозита</PaymentTitle>
                        
                        <PaymentAmount>
                            {paymentInfo.amount} {paymentInfo.currency}
                        </PaymentAmount>

                        <PaymentAddress>
                            {paymentInfo.address}
                        </PaymentAddress>

                        <PaymentComment>
                            <strong>⚠️ Важно! Укажите комментарий:</strong>
                            <code>{paymentInfo.comment}</code>
                        </PaymentComment>

                        {isConnected ? (
                            <PaymentButton onClick={handleTonConnectPayment} disabled={loading}>
                                {loading ? 'Отправка...' : '🔗 Оплатить через TON Connect'}
                            </PaymentButton>
                        ) : (
                            <PaymentButton onClick={handleDeepLinkPayment}>
                                📱 Открыть кошелек
                            </PaymentButton>
                        )}

                        <CancelButton onClick={handleCancelPayment}>
                            Отмена
                        </CancelButton>
                    </PaymentCard>
                </PaymentModal>
            )}
        </Container>
    );
};

export default CreateGamePage;
