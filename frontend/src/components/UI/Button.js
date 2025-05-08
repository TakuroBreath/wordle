import React from 'react';
import styled from 'styled-components';

// Основной стиль кнопки
const ButtonBase = styled.button`
  padding: 12px 20px;
  border-radius: 8px;
  border: none;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s ease;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 120px;
  outline: none;
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  &:active {
    opacity: 0.8;
  }
`;

// Основная кнопка (primary) в стиле Telegram
const PrimaryButton = styled(ButtonBase)`
  background-color: var(--tg-theme-button-color, #50a8eb);
  color: var(--tg-theme-button-text-color, #ffffff);
`;

// Вторичная кнопка (secondary) с прозрачным фоном
const SecondaryButton = styled(ButtonBase)`
  background-color: transparent;
  color: var(--tg-theme-button-color, #50a8eb);
  border: 1px solid var(--tg-theme-button-color, #50a8eb);
`;

// Опасная кнопка (danger) для удаления и критических действий
const DangerButton = styled(ButtonBase)`
  background-color: #ff3b30;
  color: #ffffff;
`;

// Кнопка для неактивного/нейтрального состояния
const NeutralButton = styled(ButtonBase)`
  background-color: var(--tg-theme-secondary-bg-color, #f1f1f1);
  color: var(--tg-theme-text-color, #000000);
`;

const Button = ({
    children,
    variant = 'primary',
    type = 'button',
    onClick,
    disabled = false,
    fullWidth = false,
    icon = null,
    ...props
}) => {
    // Выбор компонента кнопки в зависимости от варианта
    const ButtonComponent =
        variant === 'primary' ? PrimaryButton :
            variant === 'secondary' ? SecondaryButton :
                variant === 'danger' ? DangerButton :
                    NeutralButton;

    return (
        <ButtonComponent
            type={type}
            onClick={onClick}
            disabled={disabled}
            style={{ width: fullWidth ? '100%' : 'auto' }}
            {...props}
        >
            {icon && <span style={{ marginRight: '8px' }}>{icon}</span>}
            {children}
        </ButtonComponent>
    );
};

export default Button; 