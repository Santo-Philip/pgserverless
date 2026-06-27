package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nexbic/platform/config"
	authdto "github.com/nexbic/platform/internal/auth/dto"
	authmodels "github.com/nexbic/platform/internal/auth/models"
	authrepo "github.com/nexbic/platform/internal/auth/repository"
	"github.com/nexbic/platform/pkg/password"
)

type AuthService struct {
	userRepo  *authrepo.UserRepository
	tokenRepo *authrepo.RefreshTokenRepo
	jwtCfg    config.JWTConfig
}

func NewAuthService(userRepo *authrepo.UserRepository, tokenRepo *authrepo.RefreshTokenRepo, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtCfg:    jwtCfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *authdto.RegisterRequest) (*authdto.AuthResponse, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	isFirst, _ := s.userRepo.Count(ctx)
	role := "user"
	if isFirst == 0 {
		role = "admin"
	}

	user := &authmodels.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req *authdto.LoginRequest) (*authdto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	valid, err := password.Verify(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("update login: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*authdto.AuthResponse, error) {
	hash := password.HashKey(refreshTokenString)
	userID, err := s.tokenRepo.FindByToken(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if userID == nil {
		return nil, fmt.Errorf("refresh token expired or not found")
	}

	user, err := s.userRepo.GetByID(ctx, *userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) GetUser(ctx context.Context, id uuid.UUID) (*authmodels.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) ListUsers(ctx context.Context, limit, offset int) ([]authmodels.User, int, error) {
	return s.userRepo.List(ctx, limit, offset)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *authmodels.User) (*authdto.AuthResponse, error) {
	now := time.Now()
	accessExpires := now.Add(s.jwtCfg.AccessTTL)

	accessClaims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
		"iat":   now.Unix(),
		"exp":   accessExpires.Unix(),
		"iss":   s.jwtCfg.Issuer,
		"aud":   s.jwtCfg.Audience,
		"type":  "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshExpires := now.Add(s.jwtCfg.RefreshTTL)
	refreshClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"iat":  now.Unix(),
		"exp":  refreshExpires.Unix(),
		"iss":  s.jwtCfg.Issuer,
		"type": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	refreshHash := password.HashKey(refreshTokenString)
	if err := s.tokenRepo.Create(ctx, user.ID, refreshHash, refreshExpires); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &authdto.AuthResponse{
		Token:        accessTokenString,
		RefreshToken: refreshTokenString,
		User:         *user,
		ExpiresAt:    accessExpires,
	}, nil
}
