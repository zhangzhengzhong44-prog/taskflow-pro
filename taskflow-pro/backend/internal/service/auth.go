package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"taskflow-pro/backend/internal/config"
	"taskflow-pro/backend/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("邮箱或密码错误")
	ErrEmailExists        = errors.New("邮箱已被注册")
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users UserStore
	cache Cache
	cfg   config.JWTConfig
}

func NewAuthService(users UserStore, cache Cache, cfg config.JWTConfig) *AuthService {
	return &AuthService{users: users, cache: cache, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.AuthResponse, error) {
	if _, err := s.users.FindByEmail(req.Email); err == nil {
		return model.AuthResponse{}, ErrEmailExists
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return model.AuthResponse{}, err
	}

	user := model.User{Username: req.Username, Email: req.Email, PasswordHash: hash}
	if err := s.users.Create(&user); err != nil {
		return model.AuthResponse{}, err
	}

	token, err := s.IssueToken(user.ID)
	if err != nil {
		return model.AuthResponse{}, err
	}
	_ = s.cacheUser(ctx, user)

	return model.AuthResponse{Token: token, User: model.ToUserResponse(user)}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		return model.AuthResponse{}, ErrInvalidCredentials
	}
	if !CheckPassword(req.Password, user.PasswordHash) {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.IssueToken(user.ID)
	if err != nil {
		return model.AuthResponse{}, err
	}
	_ = s.cacheUser(ctx, *user)

	return model.AuthResponse{Token: token, User: model.ToUserResponse(*user)}, nil
}

func (s *AuthService) Me(ctx context.Context, userID uint) (model.UserResponse, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil {
			var user model.UserResponse
			if json.Unmarshal([]byte(cached), &user) == nil {
				return user, nil
			}
		}
	}

	user, err := s.users.FindByID(userID)
	if err != nil {
		return model.UserResponse{}, err
	}
	_ = s.cacheUser(ctx, *user)
	return model.ToUserResponse(*user), nil
}

func (s *AuthService) IssueToken(userID uint) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.TTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "taskflow-pro",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
}

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效 token")
	}
	return claims, nil
}

func (s *AuthService) cacheUser(ctx context.Context, user model.User) error {
	if s.cache == nil {
		return nil
	}
	data, err := json.Marshal(model.ToUserResponse(user))
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, fmt.Sprintf("user:%d", user.ID), data, time.Hour)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
