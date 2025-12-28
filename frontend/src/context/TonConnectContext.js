import React, { createContext, useContext, useCallback, useEffect, useState } from 'react';
import { TonConnectUIProvider, useTonConnectUI, useTonWallet, useTonAddress } from '@tonconnect/ui-react';
import { userAPI } from '../api';

// Контекст для работы с TON Connect
const TonConnectContext = createContext(null);

// Внутренний провайдер с логикой TON Connect
const TonConnectInnerProvider = ({ children }) => {
    const [tonConnectUI] = useTonConnectUI();
    const wallet = useTonWallet();
    const address = useTonAddress();
    const [isWalletSynced, setIsWalletSynced] = useState(false);

    // Синхронизация адреса кошелька с бэкендом
    useEffect(() => {
        const syncWallet = async () => {
            if (address && !isWalletSynced) {
                try {
                    await userAPI.connectWallet(address);
                    setIsWalletSynced(true);
                    console.log('Кошелек синхронизирован с бэкендом:', address);
                } catch (error) {
                    console.error('Ошибка синхронизации кошелька:', error);
                }
            }
        };

        syncWallet();
    }, [address, isWalletSynced]);

    // Сброс флага синхронизации при отключении кошелька
    useEffect(() => {
        if (!wallet) {
            setIsWalletSynced(false);
        }
    }, [wallet]);

    // Подключение кошелька
    const connectWallet = useCallback(async () => {
        try {
            await tonConnectUI.openModal();
        } catch (error) {
            console.error('Ошибка подключения кошелька:', error);
            throw error;
        }
    }, [tonConnectUI]);

    // Отключение кошелька
    const disconnectWallet = useCallback(async () => {
        try {
            await tonConnectUI.disconnect();
            setIsWalletSynced(false);
        } catch (error) {
            console.error('Ошибка отключения кошелька:', error);
            throw error;
        }
    }, [tonConnectUI]);

    // Отправка TON транзакции
    const sendTransaction = useCallback(async (toAddress, amount, comment = '') => {
        if (!wallet) {
            throw new Error('Кошелек не подключен');
        }

        try {
            // Конвертируем TON в nanotons (1 TON = 10^9 nanotons)
            const amountInNanotons = Math.floor(amount * 1000000000).toString();

            const transaction = {
                validUntil: Math.floor(Date.now() / 1000) + 600, // 10 минут
                messages: [
                    {
                        address: toAddress,
                        amount: amountInNanotons,
                        payload: comment ? encodeComment(comment) : undefined,
                    }
                ]
            };

            const result = await tonConnectUI.sendTransaction(transaction);
            console.log('Транзакция отправлена:', result);
            return result;
        } catch (error) {
            console.error('Ошибка отправки транзакции:', error);
            throw error;
        }
    }, [wallet, tonConnectUI]);

    // Формирование deep link для оплаты
    const getPaymentDeepLink = useCallback((toAddress, amount, comment = '') => {
        const amountInNanotons = Math.floor(amount * 1000000000);
        let url = `ton://transfer/${toAddress}?amount=${amountInNanotons}`;
        if (comment) {
            url += `&text=${encodeURIComponent(comment)}`;
        }
        return url;
    }, []);

    const value = {
        wallet,
        address,
        isConnected: !!wallet,
        isWalletSynced,
        connectWallet,
        disconnectWallet,
        sendTransaction,
        getPaymentDeepLink,
        tonConnectUI,
    };

    return (
        <TonConnectContext.Provider value={value}>
            {children}
        </TonConnectContext.Provider>
    );
};

// Функция кодирования комментария в base64 payload
const encodeComment = (comment) => {
    // Простой текстовый комментарий (opcode 0)
    const encoder = new TextEncoder();
    const commentBytes = encoder.encode(comment);
    
    // Формируем ячейку с opcode 0 (текстовый комментарий)
    // Для простоты возвращаем base64 строку
    const buffer = new ArrayBuffer(4 + commentBytes.length);
    const view = new DataView(buffer);
    view.setUint32(0, 0, false); // opcode 0 для текстового комментария
    new Uint8Array(buffer, 4).set(commentBytes);
    
    return btoa(String.fromCharCode(...new Uint8Array(buffer)));
};

// Главный провайдер TON Connect
export const TonConnectProvider = ({ children }) => {
    const manifestUrl = window.location.origin + '/tonconnect-manifest.json';

    return (
        <TonConnectUIProvider manifestUrl={manifestUrl}>
            <TonConnectInnerProvider>
                {children}
            </TonConnectInnerProvider>
        </TonConnectUIProvider>
    );
};

// Хук для использования TON Connect
export const useTonConnect = () => {
    const context = useContext(TonConnectContext);
    if (!context) {
        throw new Error('useTonConnect must be used within a TonConnectProvider');
    }
    return context;
};

export default TonConnectContext;
