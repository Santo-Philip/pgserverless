package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"strings"

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
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

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
	if domain.Verified {
		return nil
	}

	expectedRecord := "nexbic-verify=" + domain.VerificationToken
	found, err := checkTXTRecord(domain.Domain, expectedRecord)
	if err != nil {
		return fmt.Errorf("dns lookup failed: %w", err)
	}
	if !found {
		return fmt.Errorf("domain verification failed: add a TXT record with value %q to your domain's DNS", expectedRecord)
	}

	slog.Info("domain verified via DNS", "domain", domain.Domain, "id", domain.ID)
	return s.domainRepo.Verify(ctx, domainID)
}

func (s *DomainService) Delete(ctx context.Context, domainID uuid.UUID) error {
	return s.domainRepo.Delete(ctx, domainID)
}

func checkTXTRecord(domain, expected string) (bool, error) {
	txtRecords, err := net.LookupTXT("_nexbic-verify." + domain)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return false, nil
		}
		return false, err
	}

	for _, record := range txtRecords {
		if strings.TrimSpace(record) == expected {
			return true, nil
		}
	}
	return false, nil
}
