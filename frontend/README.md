# Wordle Game - Telegram Mini App

Игра Wordle с возможностью делать ставки и выигрывать TON.

## Описание

Wordle Game - это Telegram Mini App, которая позволяет пользователям играть в популярную игру Wordle и делать ставки в криптовалюте TON или USDT. Игроки могут создавать свои игры, загадывая слова, или присоединяться к играм других пользователей.

## Функциональность

- Авторизация через Telegram Mini App
- Создание игр с настраиваемыми параметрами
- Присоединение к играм других пользователей
- Угадывание слов с визуальной обратной связью
- Система ставок и выигрышей
- Пополнение и вывод средств через TON и USDT

## Технологии

- React 18
- React Router 6
- Styled Components
- Axios
- Telegram Mini Apps SDK

## Запуск проекта

### Требования

- Node.js 16+
- npm или yarn

### Установка зависимостей

```bash
npm install
```

или

```bash
yarn install
```

### Запуск в режиме разработки

```bash
npm start
```

или

```bash
yarn start
```

### Сборка для продакшена

```bash
npm run build
```

или

```bash
yarn build
```

## Структура проекта

```
src/
  ├── api/           # API клиент для взаимодействия с бэкендом
  ├── components/    # Переиспользуемые компоненты
  ├── context/       # React контексты (авторизация и др.)
  ├── hooks/         # Кастомные React хуки
  ├── pages/         # Компоненты страниц
  ├── utils/         # Вспомогательные функции
  ├── App.js         # Основной компонент приложения
  └── index.js       # Точка входа
```

## Интеграция с Telegram Mini Apps

Приложение использует Telegram Mini Apps SDK для интеграции с Telegram. Для авторизации используется механизм подписи данных Telegram, который проверяется на бэкенде.

## Взаимодействие с бэкендом

Взаимодействие с бэкендом осуществляется через REST API с использованием библиотеки Axios. API клиент находится в директории `src/api/`. 