package repositories

import (
	"context"
	"fmt"
	"time"

	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"

	"github.com/karlseguin/ccache/v3"
)

type TransfersCCacheRepo struct {
	cache *ccache.Cache[models.Transfer]
	ttl   time.Duration
}

func NewTransfersCCacheRepository(cfg config.CCache) *TransfersCCacheRepo {
	cacheCfg := ccache.Configure[models.Transfer]().
		MaxSize(cfg.MaxSize).
		GetsPerPromote(cfg.GetsPerPromote).
		PercentToPrune(cfg.PercentToPrune)

	return &TransfersCCacheRepo{
		cache: ccache.New(cacheCfg),
		ttl:   time.Duration(cfg.TTLSeconds) * time.Second,
	}
}

func (r *TransfersCCacheRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	_ = ctx
	if transfer.ID == "" {
		return "", fmt.Errorf("transfer ID required for cache create")
	}

	if transfer.Currency == enums.CurrencyUnknown {
		return "", fmt.Errorf("valid currency required for cache create")
	}

	r.cache.Set(transfer.ID, cloneTransfer(transfer), r.ttl)
	return transfer.ID, nil
}

func (r *TransfersCCacheRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	_ = ctx

	item := r.cache.Get(id)
	if item == nil || item.Expired() {
		if item != nil && item.Expired() {
			r.cache.Delete(id)
		}
		return models.Transfer{}, fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
	}

	return cloneTransfer(item.Value()), nil
}

func (r *TransfersCCacheRepo) GetTransfersByUserID(ctx context.Context, userID string) ([]models.Transfer, error) {
	_ = ctx

	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	var transfers []models.Transfer

	r.cache.ForEachFunc(func(key string, item *ccache.Item[models.Transfer]) bool {
		if item == nil {
			return true
		}

		if item.Expired() {
			r.cache.Delete(key)
			return true
		}

		transfer := item.Value()
		if transfer.SenderID == userID || transfer.ReceiverID == userID {
			transfers = append(transfers, cloneTransfer(transfer))
		}

		return true
	})

	if len(transfers) == 0 {
		return nil, fmt.Errorf("transfers not found for userID %s: %w", userID, known_errors.ErrNotFound)
	}

	return transfers, nil
}

func (r *TransfersCCacheRepo) Update(ctx context.Context, transfer models.Transfer) error {
	_ = ctx

	item := r.cache.Get(transfer.ID)
	if item == nil || item.Expired() {
		if item != nil && item.Expired() {
			r.cache.Delete(transfer.ID)
		}
		return fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
	}

	cached := item.Value()
	if transfer.SenderID != "" {
		cached.SenderID = transfer.SenderID
	}
	if transfer.ReceiverID != "" {
		cached.ReceiverID = transfer.ReceiverID
	}
	if transfer.Currency != enums.CurrencyUnknown {
		cached.Currency = transfer.Currency
	}
	if transfer.Amount != 0 {
		cached.Amount = transfer.Amount
	}
	if transfer.State != "" {
		cached.State = transfer.State
	}

	r.cache.Set(transfer.ID, cloneTransfer(cached), r.ttl)
	return nil
}

func (r *TransfersCCacheRepo) Delete(ctx context.Context, id string) error {
	_ = ctx

	item := r.cache.GetWithoutPromote(id)
	if item == nil || item.Expired() {
		if item != nil && item.Expired() {
			r.cache.Delete(id)
		}
		return fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
	}

	r.cache.Delete(id)
	return nil
}

func cloneTransfer(transfer models.Transfer) models.Transfer {
	return models.Transfer{
		ID:         transfer.ID,
		SenderID:   transfer.SenderID,
		ReceiverID: transfer.ReceiverID,
		Currency:   transfer.Currency,
		Amount:     transfer.Amount,
		State:      transfer.State,
	}
}
