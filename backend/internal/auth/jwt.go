package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// define the structure of the data that we put into the JWT claims
type CustomClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	KycStatus string    `json:"kyc_status"`
	jwt.RegisteredClaims
}

func GenerateToken(secretKey []byte, userID uuid.UUID, kycStatus string) (string, error) {
	// create the payload (claims)
	claims := CustomClaims{
		UserID:    userID,
		KycStatus: kycStatus,
		RegisteredClaims: jwt.RegisteredClaims{
			// token expires in 24 hours
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			// issues at this exact moment
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// The entity that creates this token
			Issuer: "public-policy-forum",
		},
	}

	// create the token using HS256 algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign the token with our secret key
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func VerifyToken(secretKey []byte, tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
