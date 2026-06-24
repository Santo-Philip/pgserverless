package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/shared/models"
)

type DomainService struct {
	domainRepo *repository.DomainRepository
}

func NewDomainService(domainRepo *repository.DomainRepository) *DomainService {
	return &DomainService{domainRepo: domainRepo}
}

func (s *DomainService) Create(ctx context.Context, appID uuid.UUID, domainName string) (*models.Domain, error) {
	token := make([]byte, 16)
	rand.Read(token)

	domain := &models.Domain{
		AppID:             appID,
		Domain:            domainName,
		Status:            models.DomainStatusPending,
		Verified:          false,
		VerificationToken: hex.EncodeToString(token),
	}

	if err := s.domainRepo.Create(ctx, domain); err != nil {
		return nil, fmt.Errorf("create domain: %w", err)
	}

	return domain, nil
}

func (s *DomainService) ListByApp(ctx context.Context, appID uuid.UUID) ([]models.Domain, error) {
	return s.domainRepo.ListByApp(ctx, appID)
}

func (s *DomainService) Verify(ctx context.Context, domainID uuid.UUID) error {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return err
	}
	if domain == nil {
		return fmt.Errorf("domain not found")
	}

	return s.domainRepo.Verify(ctx, domainID)
}

func (s *DomainService) Delete(ctx context.Context, domainID uuid.UUID) error {
	return s.domainRepo.Delete(ctx, domainID)
}
