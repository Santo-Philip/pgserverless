package service

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nexbic/platform/config"
	authdto "github.com/nexbic/platform/internal/identity/auth/dto"
	authmodels "github.com/nexbic/platform/internal/identity/auth/models"
	authrepo "github.com/nexbic/platform/internal/identity/auth/repository"
	"github.com/nexbic/platform/pkg/password"
)

type AuthService struct {
	userRepo      *authrepo.UserRepository
	secRepo       *authrepo.SecurityRepo
	tokenRepo     *authrepo.RefreshTokenRepo
	jwtCfg        config.JWTConfig
	superAdminCfg config.SuperAdminConfig
}

func NewAuthService(userRepo *authrepo.UserRepository, secRepo *authrepo.SecurityRepo, tokenRepo *authrepo.RefreshTokenRepo, jwtCfg config.JWTConfig, superAdminCfg config.SuperAdminConfig) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		secRepo:       secRepo,
		tokenRepo:     tokenRepo,
		jwtCfg:        jwtCfg,
		superAdminCfg: superAdminCfg,
	}
}

func (s *AuthService) SeedSuperAdmin(ctx context.Context) {
	if s.superAdminCfg.Email == "" || s.superAdminCfg.Password == "" {
		return
	}
	existing, err := s.userRepo.GetByEmail(ctx, s.superAdminCfg.Email)
	if err != nil {
		slog.Error("seed superadmin: check existing", "error", err)
		return
	}
	if existing != nil {
		slog.Info("superadmin already exists", "email", s.superAdminCfg.Email)
		return
	}
	hash, err := password.Hash(s.superAdminCfg.Password)
	if err != nil {
		slog.Error("seed superadmin: hash password", "error", err)
		return
	}
	user := &authmodels.User{
		Email:         s.superAdminCfg.Email,
		PasswordHash:  hash,
		Name:          "Super Administrator",
		Role:          authmodels.RoleSuperAdmin,
		IsActive:      true,
		EmailVerified: true,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		slog.Error("seed superadmin: create user", "error", err)
		return
	}
	slog.Info("superadmin seeded", "email", s.superAdminCfg.Email)
}

func (s *AuthService) Login(ctx context.Context, req *authdto.LoginRequest, ipAddress, userAgent string) (*authdto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
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
	s.secRepo.LogEvent(ctx, user.ID, "login_success", "low", ipAddress, userAgent)
	s.secRepo.UpsertDevice(ctx, user.ID, userAgent, "web", ipAddress, "")
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
	if err != nil || user == nil {
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

func (s *AuthService) CreateUser(ctx context.Context, req *authdto.CreateUserRequest) (*authmodels.User, error) {
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
	user := &authmodels.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         req.Role,
		IsActive:     true,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, id uuid.UUID, req *authdto.UpdateUserRequest) (*authmodels.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, userID uuid.UUID, req *authdto.UpdatePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}
	valid, err := password.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil || !valid {
		return fmt.Errorf("current password is incorrect")
	}
	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.userRepo.UpdatePassword(ctx, userID, hash)
}

func (s *AuthService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, req *authdto.UpdateUserPasswordRequest) error {
	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.userRepo.UpdatePassword(ctx, userID, hash)
}

func (s *AuthService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *AuthService) ListDevices(ctx context.Context, userID uuid.UUID) ([]authmodels.Device, error) {
	return s.secRepo.ListDevices(ctx, userID)
}

func (s *AuthService) DeleteDevice(ctx context.Context, deviceID, userID uuid.UUID) error {
	return s.secRepo.DeleteDevice(ctx, deviceID, userID)
}

func (s *AuthService) ListSecurityEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]authmodels.SecurityEvent, int, error) {
	return s.secRepo.ListEvents(ctx, userID, limit, offset)
}

func (s *AuthService) CreateAPIKey(ctx context.Context, userID uuid.UUID, req *authdto.CreateAPIKeyRequest) (*authdto.CreateAPIKeyResponse, error) {
	rawKey, keyHash, keyPrefix, err := password.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	key := &authmodels.APIKey{
		UserID:    userID,
		Name:      req.Name,
		Prefix:    keyPrefix,
		Hash:      keyHash,
		Status:    "active",
		ExpiresAt: req.ExpiresAt,
	}
	if err := s.secRepo.CreateAPIKey(ctx, key); err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	return &authdto.CreateAPIKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Key:       rawKey,
		Prefix:    key.Prefix,
		Status:    key.Status,
		ExpiresAt: key.ExpiresAt,
		CreatedAt: key.CreatedAt,
	}, nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context, userID uuid.UUID) ([]authdto.APIKeyResponse, error) {
	keys, err := s.secRepo.ListAPIKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]authdto.APIKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = authdto.APIKeyResponse{
			ID:        k.ID,
			Name:      k.Name,
			Prefix:    k.Prefix,
			Status:    k.Status,
			LastUsed:  k.LastUsedAt,
			ExpiresAt: k.ExpiresAt,
			CreatedAt: k.CreatedAt,
		}
	}
	return resp, nil
}

func (s *AuthService) RevokeAPIKey(ctx context.Context, keyID, userID uuid.UUID) error {
	return s.secRepo.RevokeAPIKey(ctx, keyID, userID)
}

func (s *AuthService) GetOAuthURL(provider string) string {
	switch provider {
	case "google":
		return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email+profile",
			s.jwtCfg.OAuthGoogleClientID, s.jwtCfg.OAuthRedirectURL)
	case "github":
		return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email",
			s.jwtCfg.OAuthGitHubClientID, s.jwtCfg.OAuthRedirectURL)
	default:
		return ""
	}
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider, code string) (*authdto.AuthResponse, error) {
	return nil, fmt.Errorf("oauth %s callback not implemented in v1", provider)
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
