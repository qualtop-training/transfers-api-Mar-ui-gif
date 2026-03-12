package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"transfers-api/internal/config"
	"transfers-api/internal/logging"
	"transfers-api/internal/models"
)

type TransfersMySQLRepo struct {
	db *sql.DB
}

type transferMySQLDAO struct {
	ID         int64
	SenderID   string
	ReceiverID string
	Currency   string
	Amount     float64
	State      string
}

func NewTransfersMySQLRepository(cfg config.MySQL) *TransfersMySQLRepo {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging.Logger.Fatalf("error opening MySQL: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logging.Logger.Fatalf("error connecting to MySQL: %v", err)
	}

	return &TransfersMySQLRepo{db: db}
}

func (r *TransfersMySQLRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	return "", errors.New("not implemented")
}

func (r *TransfersMySQLRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	return models.Transfer{}, errors.New("not implemented")
}

func (r *TransfersMySQLRepo) Update(ctx context.Context, transfer models.Transfer) error {
	return errors.New("not implemented")
}

func (r *TransfersMySQLRepo) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
