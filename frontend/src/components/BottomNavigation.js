import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';

const NavigationContainer = styled.nav`
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  display: flex;
  justify-content: space-around;
  padding: 12px 0;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
  z-index: 1000;
`;

const NavItem = styled.button`
  display: flex;
  flex-direction: column;
  align-items: center;
  text-decoration: none;
  color: ${props => props.$active ? '#2196F3' : '#666'};
  font-size: 12px;
  padding: 4px 0;
  width: 33.33%;
  background: none;
  border: none;
  cursor: pointer;
  
  svg {
    margin-bottom: 4px;
    width: 24px;
    height: 24px;
    stroke-width: 2;
  }

  &:hover {
    color: #2196F3;
  }
`;

const BottomNavigation = () => {
    const location = useLocation();
    const navigate = useNavigate();

    const handleNavigation = (path) => {
        navigate(path);
    };

    return (
        <NavigationContainer>
            <NavItem
                $active={location.pathname === '/my-games'}
                onClick={() => handleNavigation('/my-games')}
            >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                    <polyline points="9 22 9 12 15 12 15 22" />
                </svg>
                Мои игры
            </NavItem>
            <NavItem
                $active={location.pathname === '/all-games'}
                onClick={() => handleNavigation('/all-games')}
            >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path d="M21 12a9 9 0 0 1-9 9m9-9a9 9 0 0 0-9-9m9 9H3m9 9a9 9 0 0 1-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 0 1 9-9" />
                </svg>
                Все игры
            </NavItem>
            <NavItem
                $active={location.pathname === '/profile'}
                onClick={() => handleNavigation('/profile')}
            >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                    <circle cx="12" cy="7" r="4" />
                </svg>
                Профиль
            </NavItem>
        </NavigationContainer>
    );
};

export default BottomNavigation; 