package service

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nexbic/platform/config"
	authdto "github.com/nexbic/platform/internal/auth/dto"
	authmodels "github.com/nexbic/platform/internal/auth/models"
	authrepo "github.com/nexbic/platform/internal/auth/repository"
	"github.com/nexbic/platform/pkg/password"
	"github.com/nexbic/platform/pkg/totp"
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
		Email:        s.superAdminCfg.Email,
		PasswordHash: hash,
		Name:         "Super Administrator",
		Role:         authmodels.RoleSuperAdmin,
		IsActive:     true,
		EmailVerified: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		slog.Error("seed superadmin: create user", "error", err)
		return
	}

	slog.Info("superadmin seeded", "email", s.superAdminCfg.Email)
}

// ── Register ────────────────────────────────────────────

func (s *AuthService) Register(ctx context.Context, req *authdto.RegisterRequest) (*authmodels.User, error) {
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
		Role:         authmodels.RoleDeveloper,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// ── Login (with optional TOTP) ──────────────────────────

func (s *AuthService) Login(ctx context.Context, req *authdto.LoginRequest, ipAddress, userAgent string) (*authdto.AuthResponse, error) {
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

	if user.TOTPEnabled && req.TOTPCode == "" {
		return nil, fmt.Errorf("totp_code_required")
	}

	if user.TOTPEnabled {
		if !totp.ValidateCode(user.TOTPSecret, req.TOTPCode, time.Now(), 1) {
			// Check recovery codes
			if !s.checkRecoveryCode(ctx, user, req.TOTPCode) {
				return nil, fmt.Errorf("invalid totp code")
			}
		}
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("update login: %w", err)
	}

	s.secRepo.LogEvent(ctx, user.ID, "login_success", "low", ipAddress, userAgent)
	s.secRepo.UpsertDevice(ctx, user.ID, userAgent, "web", ipAddress, "")

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) checkRecoveryCode(ctx context.Context, user *authmodels.User, code string) bool {
	for i, c := range user.RecoveryCodes {
		if c == code {
			updated := append(user.RecoveryCodes[:i], user.RecoveryCodes[i+1:]...)
			s.userRepo.SetRecoveryCodes(ctx, user.ID, updated)
			return true
		}
	}
	return false
}

// ── TOTP ────────────────────────────────────────────────

func (s *AuthService) EnableTOTP(ctx context.Context, userID uuid.UUID) (*authdto.EnableTOTPResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.TOTPEnabled {
		return nil, fmt.Errorf("totp already enabled")
	}

	secret, err := totp.GenerateSecret()
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}

	if err := s.userRepo.UpdateTOTP(ctx, userID, secret); err != nil {
		return nil, fmt.Errorf("save totp secret: %w", err)
	}

	qrURL := totp.GetQRCodeURL(secret, "Nexbic Platform", user.Email)

	return &authdto.EnableTOTPResponse{
		Secret:    secret,
		QRCodeURL: qrURL,
	}, nil
}

func (s *AuthService) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if !user.TOTPEnabled {
		return nil, fmt.Errorf("totp not enabled")
	}

	if !totp.ValidateCode(user.TOTPSecret, code, time.Now(), 1) {
		return nil, fmt.Errorf("invalid code")
	}

	backupCodes := totp.GenerateBackupCodes(8)
	if err := s.userRepo.SetRecoveryCodes(ctx, userID, backupCodes); err != nil {
		return nil, fmt.Errorf("save recovery codes: %w", err)
	}

	return backupCodes, nil
}

func (s *AuthService) DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}
	if !user.TOTPEnabled {
		return fmt.Errorf("totp not enabled")
	}

	if !totp.ValidateCode(user.TOTPSecret, code, time.Now(), 1) {
		if !s.checkRecoveryCode(ctx, user, code) {
			return fmt.Errorf("invalid code")
		}
	}

	return s.userRepo.DisableTOTP(ctx, userID)
}

// ── Email Verification ──────────────────────────────────

func (s *AuthService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return "", fmt.Errorf("user not found")
	}
	if user.EmailVerified {
		return "", fmt.Errorf("email already verified")
	}

	token, err := password.GenerateToken(32)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	tokenHash := password.HashKey(token)
	if err := s.secRepo.CreateEmailVerificationToken(ctx, userID, tokenHash, time.Now().Add(24*time.Hour)); err != nil {
		return "", fmt.Errorf("store token: %w", err)
	}

	// Return token for development — in production, send via email
	return token, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := password.HashKey(token)
	userID, err := s.secRepo.FindEmailVerificationToken(ctx, tokenHash)
	if err != nil || userID == nil {
		return fmt.Errorf("invalid or expired token")
	}

	if err := s.userRepo.UpdateEmailVerified(ctx, *userID); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}

	return s.secRepo.DeleteEmailVerificationTokens(ctx, *userID)
}

// ── Password Reset ──────────────────────────────────────

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Don't reveal if email exists
		return "", nil
	}

	token, err := password.GenerateToken(32)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	tokenHash := password.HashKey(token)
	if err := s.secRepo.CreatePasswordResetToken(ctx, email, tokenHash, time.Now().Add(1*time.Hour)); err != nil {
		return "", fmt.Errorf("store token: %w", err)
	}

	// Return token for development — in production, send via email
	return token, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	tokenHash := password.HashKey(token)
	email, err := s.secRepo.FindPasswordResetToken(ctx, tokenHash)
	if err != nil || email == nil {
		return fmt.Errorf("invalid or expired token")
	}

	user, err := s.userRepo.GetByEmail(ctx, *email)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}

	hash, err := password.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	s.secRepo.DeletePasswordResetTokens(ctx, *email)
	s.tokenRepo.DeleteByUser(ctx, user.ID)
	s.secRepo.LogEvent(ctx, user.ID, "password_reset", "medium", "", "")

	return nil
}

// ── Devices ─────────────────────────────────────────────

func (s *AuthService) ListDevices(ctx context.Context, userID uuid.UUID) ([]authmodels.Device, error) {
	return s.secRepo.ListDevices(ctx, userID)
}

func (s *AuthService) DeleteDevice(ctx context.Context, deviceID, userID uuid.UUID) error {
	return s.secRepo.DeleteDevice(ctx, deviceID, userID)
}

// ── Security Events ─────────────────────────────────────

func (s *AuthService) ListSecurityEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]authmodels.SecurityEvent, int, error) {
	return s.secRepo.ListEvents(ctx, userID, limit, offset)
}

// ── API Keys ────────────────────────────────────────────

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

// ── Existing User Management ────────────────────────────

func (s *AuthService) LoginOriginal(ctx context.Context, req *authdto.LoginRequest) (*authdto.AuthResponse, error) {
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
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if user == nil {
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

// ── Token Generation ────────────────────────────────────

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
