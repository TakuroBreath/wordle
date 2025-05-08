import React from 'react';
import styled from 'styled-components';

// Основной стиль карточки
const CardContainer = styled.div`
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  border: 1px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.1));
  overflow: hidden;
  width: ${props => props.fullWidth ? '100%' : 'auto'};
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  
  ${props => props.clickable && `
    cursor: pointer;
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }
    &:active {
      transform: translateY(0);
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    }
  `}
`;

// Заголовок карточки
const CardTitle = styled.h3`
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--tg-theme-text-color, #000000);
`;

// Контент карточки
const CardContent = styled.div`
  margin-bottom: ${props => props.noMargin ? '0' : '8px'};
`;

// Нижняя часть карточки (опционально)
const CardFooter = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--tg-theme-hint-color, rgba(0, 0, 0, 0.1));
`;

// Компонент Badge для статуса и меток
const Badge = styled.span`
  display: inline-block;
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 12px;
  margin-right: 8px;
  color: white;
  background-color: ${props => {
        switch (props.variant) {
            case 'success':
                return '#34c759';
            case 'warning':
                return '#ff9500';
            case 'danger':
                return '#ff3b30';
            case 'info':
            default:
                return 'var(--tg-theme-button-color, #50a8eb)';
        }
    }};
`;

const Card = ({
    title,
    children,
    footer,
    badges = [],
    clickable = false,
    fullWidth = false,
    onClick,
    ...props
}) => {
    return (
        <CardContainer
            clickable={clickable}
            fullWidth={fullWidth}
            onClick={clickable && onClick ? onClick : undefined}
            {...props}
        >
            {title && (
                <CardTitle>
                    {title}
                    {badges && badges.length > 0 && (
                        <div style={{ display: 'inline-flex', marginLeft: '8px' }}>
                            {badges.map((badge, index) => (
                                <Badge key={index} variant={badge.variant}>
                                    {badge.text}
                                </Badge>
                            ))}
                        </div>
                    )}
                </CardTitle>
            )}

            <CardContent noMargin={!footer}>
                {children}
            </CardContent>

            {footer && <CardFooter>{footer}</CardFooter>}
        </CardContainer>
    );
};

export default Card; 