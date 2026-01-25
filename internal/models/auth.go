package models

// Auth provider constants.
const (
	AuthProviderTelegram = "telegram"
	AuthProviderBase     = "base"
)

// BaseAuthMethod constants.
const (
	BaseAuthMethodSIWF   = "siwf"
	BaseAuthMethodWallet = "wallet"
)

// BaseAuthPayload представляет данные Base Mini App авторизации.
type BaseAuthPayload struct {
	Address            string `json:"address"`
	Message            string `json:"message"`
	Signature          string `json:"signature"`
	FID                uint64 `json:"fid,omitempty"`
	Username           string `json:"username,omitempty"`
	DisplayName        string `json:"display_name,omitempty"`
	PfpURL             string `json:"pfp_url,omitempty"`
	CustodyAddress     string `json:"custody_address,omitempty"`
	SmartWalletAddress string `json:"smart_wallet_address,omitempty"`
	AuthMethod         string `json:"auth_method,omitempty"`
}
