package services

import (
	"context"
	"testing"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/models"
	"transfers-api/internal/services/mocks"

	"github.com/stretchr/testify/assert"
)

func TestTransfersService_GetByID(t *testing.T) {
	// given
	var (
		ctx                = context.Background()
		cfg                = config.BusinessConfig{TransferMinAmount: 1}
		transfersRepo      = mocks.NewTransfersRepositoryMock(t)
		transfersCCache    = mocks.NewTransfersRepositoryMock(t)
		transfersPublisher = mocks.NewTransfersPublisherMock(t)
	)

	for _, testCase := range []struct {
		name             string
		transferID       string
		expectedTransfer models.Transfer
	}{
		{
			name:       "Transfer successfully retrieved",
			transferID: "Test-1",
			expectedTransfer: models.Transfer{
				ID:         "Test-1",
				SenderID:   "Sender-123",
				ReceiverID: "Receiver-456",
				Amount:     100,
				Currency:   enums.CurrencyUSD,
			},
		},
		{
			name:       "Transfer successfully retrieved",
			transferID: "Test-2",
			expectedTransfer: models.Transfer{
				ID:         "Test-2",
				SenderID:   "Sender-123",
				ReceiverID: "Receiver-456",
				Amount:     100,
				Currency:   enums.CurrencyUSD,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			transfersCCache.On("GetByID", ctx, testCase.transferID).Return(testCase.expectedTransfer, nil)
			svc := NewTransfersService(cfg, transfersRepo, transfersCCache, transfersPublisher)
			// when
			transfer, err := svc.GetByID(ctx, testCase.transferID)
			// then
			assert.Nil(t, err)
			assert.Equal(t, testCase.transferID, transfer.ID)
		})
	}
}

func TestTransfersService_Delete(t *testing.T) {
	// given
	var (
		ctx                = context.Background()
		cfg                = config.BusinessConfig{TransferMinAmount: 1}
		transfersRepo      = mocks.NewTransfersRepositoryMock(t)
		transfersCCache    = mocks.NewTransfersRepositoryMock(t)
		transfersPublisher = mocks.NewTransfersPublisherMock(t)
	)

	for _, testCase := range []struct {
		name        string
		transferID  string
		repoError   error
		expectError bool
	}{
		{
			name:        "Transfer successfully deleted",
			transferID:  "Test-1",
			repoError:   nil,
			expectError: false,
		},
		{
			name:        "Error deleting transfer from repository",
			transferID:  "Test-2",
			repoError:   assert.AnError,
			expectError: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			transfersRepo.On("Delete", ctx, testCase.transferID).Return(testCase.repoError)
			if testCase.repoError == nil {
				transfersCCache.On("Delete", ctx, testCase.transferID).Return(nil)
			}
			svc := NewTransfersService(cfg, transfersRepo, transfersCCache, transfersPublisher)
			// when
			err := svc.Delete(ctx, testCase.transferID)
			// then
			if testCase.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
