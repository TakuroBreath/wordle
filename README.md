Entities:

- Game
  - ID
  - Title (optional)
  - Word
  - Length
  - Creator
  - Difficulty
  - Max Tries
  - Reward Multiplier
  - Currency (TON/USDT)
  - Prize Pool
  - Min Bet
  - Max Bet
  - Created At
  - Status (active, inactive)

- User
  - Telegram ID
  - Username
  - Wallet
  - Balance
  - Wins
  - Losses

- Lobby
  - ID
  - Game
  - User
  - Max Tries
  - Tries
  - Bet
  - Potential Reward
  - Created At
  - Status (active, inactive)

- History
  - ID
  - User
  - Game
  - Lobby
  - Status (creator win, player win)
  - Created At
