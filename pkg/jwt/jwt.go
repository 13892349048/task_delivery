package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenInvalid     = errors.New("token无效")
	ErrTokenMalformed   = errors.New("token格式错误")
	ErrTokenNotValidYet = errors.New("token尚未生效")
)

// Claims JWT声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	secretKey     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey string, tokenExpiry, refreshExpiry time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenExpiry:   tokenExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

// GenerateToken 生成访问令牌
func (j *JWTManager) GenerateToken(userID uint, username, email, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{"taskmanage"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTManager) GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Issuer:    j.issuer,
		Subject:   fmt.Sprintf("%d", userID),
		Audience:  []string{"taskmanage-refresh"},
		ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken 验证令牌
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ValidateRefreshToken 验证刷新令牌
func (j *JWTManager) ValidateRefreshToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return 0, ErrTokenMalformed
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return 0, ErrTokenNotValidYet
		}
		return 0, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// 检查audience
		expectedAudience := "taskmanage-refresh"
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == expectedAudience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return 0, ErrTokenInvalid
		}
		
		// 解析用户ID
		var userID uint
		if _, err := fmt.Sscanf(claims.Subject, "%d", &userID); err != nil {
			return 0, ErrTokenInvalid
		}
		
		return userID, nil
	}

	return 0, ErrTokenInvalid
}

// RefreshToken 刷新令牌
func (j *JWTManager) RefreshToken(refreshTokenString string, username, email, role string) (string, string, error) {
	userID, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// 生成新的访问令牌和刷新令牌
	accessToken, err := j.GenerateToken(userID, username, email, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := j.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ExtractUserID 从令牌中提取用户ID
func (j *JWTManager) ExtractUserID(tokenString string) (uint, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetTokenExpiry 获取令牌过期时间
func (j *JWTManager) GetTokenExpiry() time.Duration {
	return j.tokenExpiry
}

// GetRefreshExpiry 获取刷新令牌过期时间
func (j *JWTManager) GetRefreshExpiry() time.Duration {
	return j.refreshExpiry
}
