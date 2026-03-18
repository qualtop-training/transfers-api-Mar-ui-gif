package services

import (
	"context"
	"fmt"
	"strings"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/logging"
	"transfers-api/internal/models"
)

//go:generate mockery --name TransfersRepository --structname TransfersRepositoryMock --filename transfers_repository_mock.go --output mocks --outpkg mocks

type TransfersRepository interface {
	Create(ctx context.Context, transfer models.Transfer) (string, error)
	GetByID(ctx context.Context, id string) (models.Transfer, error)
	GetByUserID(ctx context.Context, userID string) (models.Transfer, error)
	Update(ctx context.Context, transfer models.Transfer) error
	Delete(ctx context.Context, id string) error
}

type TransfersService struct {
	businessCfg     config.BusinessConfig
	transfersRepo   TransfersRepository
	transfersCCache TransfersRepository
}

func NewTransfersService(businessCfg config.BusinessConfig, transfersRepo TransfersRepository, transfersCCache TransfersRepository) *TransfersService {
	return &TransfersService{
		businessCfg:     businessCfg,
		transfersRepo:   transfersRepo,
		transfersCCache: transfersCCache,
	}
}

func (s *TransfersService) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	if strings.TrimSpace(transfer.SenderID) == "" {
		return "", fmt.Errorf("sender_id is required: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.ReceiverID) == "" {
		return "", fmt.Errorf("receiver_id is required: %w", known_errors.ErrBadRequest)
	}
	if transfer.Currency == enums.CurrencyUnknown {
		return "", fmt.Errorf("invalid currency %s: %w", transfer.Currency.String(), known_errors.ErrBadRequest)
	}
	if transfer.Amount <= 0 {
		return "", fmt.Errorf("amount should be greater than 0: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.State) == "" { // TODO: replace with enums.ParseState
		return "", fmt.Errorf("state is required: %w", known_errors.ErrBadRequest)
	}
	id, err := s.transfersRepo.Create(ctx, transfer)
	if err != nil {
		return "", fmt.Errorf("error creating transfer in repository: %w", err)
	}
	logging.Logger.Infof("Transfer created in DB with ID: %s", id)
	// also create in cache
	transfer.ID = id
	if _, err := s.transfersCCache.Create(ctx, transfer); err != nil {
		logging.Logger.Warnf("error creating transfer in ccache: %w", err)
	}
	logging.Logger.Infof("Transfer created in ccache with ID: %s", id)
	return id, nil
}

func (s *TransfersService) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	// first try to get from cache
	transfer, err := s.transfersCCache.GetByID(ctx, id)
	if err == nil {
		logging.Logger.Infof("Transfer retrieved from ccache with ID: %s", id)
		return transfer, nil
	}

	// if not found in cache, get from repository
	transfer, err = s.transfersRepo.GetByID(ctx, id)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error getting transfer %s from repository: %w", id, err)
	}
	logging.Logger.Infof("Transfer retrieved from DB with ID: %s", id)

	// also create in cache
	if _, err := s.transfersCCache.Create(ctx, transfer); err != nil {
		logging.Logger.Warnf("error creating transfer in ccache: %w", err)
	}
	logging.Logger.Infof("Transfer created in ccache with ID: %s", id)
	return transfer, nil
}

func (s *TransfersService) GetByUserID(ctx context.Context, userID string) (models.Transfer, error) {
	// first try to get from cache
	transfer, err := s.transfersCCache.GetByUserID(ctx, userID)
	if err == nil {
		logging.Logger.Infof("Transfer retrieved from ccache with ID: %s", userID)
		return transfer, nil
	}

	// if not found in cache, get from repository
	transfer, err = s.transfersRepo.GetByUserID(ctx, userID)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error getting transfer for user %s from repository: %w", userID, err)
	}
	logging.Logger.Infof("Transfer retrieved from DB with ID: %s", userID)

	// also create in cache
	if _, err := s.transfersCCache.Create(ctx, transfer); err != nil {
		logging.Logger.Warnf("error creating transfer in ccache: %w", err)
	}
	logging.Logger.Infof("Transfer created in ccache with ID: %s", userID)
	return transfer, nil
}

func (s *TransfersService) Update(ctx context.Context, transfer models.Transfer) error {
	if strings.TrimSpace(transfer.ID) == "" {
		return fmt.Errorf("ID is required: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.SenderID) == "" &&
		strings.TrimSpace(transfer.ReceiverID) == "" &&
		transfer.Currency == enums.CurrencyUnknown &&
		transfer.Amount <= 0 &&
		strings.TrimSpace(transfer.State) == "" {
		return fmt.Errorf("error updating transfer %s: no fields to update: %w", transfer.ID, known_errors.ErrBadRequest)
	}
	if err := s.transfersRepo.Update(ctx, transfer); err != nil {
		return fmt.Errorf("error updating transfer %s in repository: %w", transfer.ID, err)
	}
	// also update in cache
	if err := s.transfersCCache.Update(ctx, transfer); err != nil {
		logging.Logger.Warnf("error updating transfer in ccache: %w", err)
	}
	return nil
}

func (s *TransfersService) Delete(ctx context.Context, id string) error {
	if err := s.transfersRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting transfer %s from repository: %w", id, err)
	}
	// also delete from cache
	if err := s.transfersCCache.Delete(ctx, id); err != nil {
		logging.Logger.Warnf("error deleting transfer from ccache: %w", err)
	}
	return nil
}
