package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"subscription-service/internal/log"
	"subscription-service/internal/model"
)

func NewPostgres(dsn string, logger *log.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(16)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	logger.Info("db connected")
	return db, nil
}

// ApplyMigrations — простой раннер .sql миграций
func ApplyMigrations(db *sql.DB, dir string, logger *log.Logger) error {
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT now())`)
	applied := map[string]bool{}
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err == nil {
		for rows.Next() {
			var v string
			_ = rows.Scan(&v)
			applied[v] = true
		}
		rows.Close()
	}

	var files []string
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		ver := filepath.Base(f)
		if applied[ver] {
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply %s: %w", ver, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations(version) VALUES($1)`, ver); err != nil {
			return err
		}
		logger.Info("migration applied", "file", ver)
	}
	return nil
}

// --- Repository ---

type SubscriptionsRepo struct{ db *sql.DB }

func NewSubscriptionsRepo(db *sql.DB) *SubscriptionsRepo { return &SubscriptionsRepo{db: db} }

func (r *SubscriptionsRepo) Create(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	q := `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
	      VALUES ($1,$2,$3,$4,$5,$6)
	      RETURNING created_at, updated_at`
	err := r.db.QueryRowContext(ctx, q,
		s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).Scan(&s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *SubscriptionsRepo) GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	var s model.Subscription
	q := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	      FROM subscriptions WHERE id=$1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return s, fmt.Errorf("subscription not found")
	}
	return s, err
}

func (r *SubscriptionsRepo) Update(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	q := `UPDATE subscriptions SET service_name=$2, price=$3, start_date=$4, end_date=$5, updated_at=now()
	      WHERE id=$1 RETURNING updated_at`
	err := r.db.QueryRowContext(ctx, q, s.ID, s.ServiceName, s.Price, s.StartDate, s.EndDate).
		Scan(&s.UpdatedAt)
	return s, err
}

func (r *SubscriptionsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

func (r *SubscriptionsRepo) List(ctx context.Context, q model.ListQuery) ([]model.Subscription, error) {
	sb := strings.Builder{}
	sb.WriteString(`SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	                FROM subscriptions WHERE 1=1`)
	var args []any
	if q.UserID != "" {
		sb.WriteString(fmt.Sprintf(` AND user_id = $%d`, len(args)+1))
		args = append(args, q.UserID)
	}
	if q.ServiceName != "" {
		sb.WriteString(fmt.Sprintf(` AND service_name = $%d`, len(args)+1))
		args = append(args, q.ServiceName)
	}
	sb.WriteString(` ORDER BY created_at DESC`)
	sb.WriteString(fmt.Sprintf(` LIMIT %d OFFSET %d`, q.Limit, q.Offset))

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

// Total: суммарная стоимость подписок по месяцам в интервале [From..To] (YYYY-MM)
func (r *SubscriptionsRepo) Total(ctx context.Context, q model.TotalQuery) (int64, error) {
	sqlQ := `
WITH bounds AS (
  SELECT date_trunc('month', $1::date) AS from_m,
         date_trunc('month', $2::date) AS to_m
),
filtered AS (
  SELECT s.*
  FROM subscriptions s, bounds b
  WHERE s.user_id = $3
    AND ( $4 = '' OR s.service_name = $4)
    AND date_trunc('month', s.start_date) <= b.to_m
    AND date_trunc('month', COALESCE(s.end_date, b.to_m)) >= b.from_m
),
months AS (
  SELECT f.id, f.price, gs::date AS month
  FROM filtered f, bounds b,
       generate_series(
         GREATEST(date_trunc('month', f.start_date), b.from_m),
         LEAST(date_trunc('month', COALESCE(f.end_date, b.to_m)), b.to_m),
         interval '1 month'
       ) AS gs
)
SELECT COALESCE(SUM(price),0) FROM months;
`
	var total int64
	err := r.db.QueryRowContext(ctx, sqlQ, q.From, q.To, q.UserID, q.ServiceName).Scan(&total)
	return total, err
}
