package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/shared/config"
	"github.com/nexbic/platform/shared/models"
	"github.com/nexbic/platform/shared/utils"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
	v := utils.NewValidator()
	v.Required("email", req.Email)
	v.Required("password", req.Password)
	v.Email("email", req.Email)
	v.MinLength("password", req.Password, 8)
	if v.HasErrors() {
		return nil, fmt.Errorf("validation: %s", v.Error())
	}

	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	v := utils.NewValidator()
	v.Required("email", req.Email)
	v.Required("password", req.Password)
	if v.HasErrors() {
		return nil, fmt.Errorf("validation: %s", v.Error())
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if user.Status != models.UserStatusActive {
		return nil, fmt.Errorf("account is not active")
	}

	valid, err := utils.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid email or password")
	}

	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *models.User) (*models.AuthResponse, error) {
	now := time.Now()
	accessExp := now.Add(s.jwtCfg.AccessTTL)
	refreshExp := now.Add(s.jwtCfg.RefreshTTL)

	accessClaims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"name":  user.Name,
		"role":  "authenticated",
		"type":  "access",
		"iss":   s.jwtCfg.Issuer,
		"aud":   s.jwtCfg.Audience,
		"exp":   accessExp.Unix(),
		"iat":   now.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"type": "refresh",
		"iss":  s.jwtCfg.Issuer,
		"exp":  refreshExp.Unix(),
		"iat":  now.Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &models.AuthResponse{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
		User:         *user,
		ExpiresAt:    accessExp,
	}, nil
}

func (s *AuthService) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	return s.userRepo.List(ctx, limit, offset)
}

func (s *AuthService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*models.AuthResponse, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if claims["type"] != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user in token")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.generateAuthResponse(ctx, user)
}
