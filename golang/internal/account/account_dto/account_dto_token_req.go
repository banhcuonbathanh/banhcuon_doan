package account_dto

import "time"

type RefreshToken struct {
    ID           int64     `json:"id"`
    AccountID    int64     `json:"account_id"`
    Token        string    `json:"token"`
    ExpiresAt    time.Time `json:"expires_at"`
    CreatedAt    time.Time `json:"created_at"`
    RevokedAt    *time.Time `json:"revoked_at,omitempty"`
    IsRevoked    bool      `json:"is_revoked"`
}

type GetNewAccessTokenRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

type RGetNewAccessTokenResponse struct {
    AccessToken  string `json:"access_token"`

    ExpiresIn    int64  `json:"expires_in"`
}
