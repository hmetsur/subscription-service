package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"subscription-service/internal/log"
	"subscription-service/internal/model"
	"subscription-service/internal/repo"
)

type Service struct {
	repo   *repo.SubscriptionsRepo
	logger *log.Logger
}

func New(r *repo.SubscriptionsRepo, l *log.Logger) *Service { return &Service{repo: r, logger: l} }

func (s *Service) Create(ctx context.Context, in model.SubscriptionCreate) (model.Subscription, error) {
	if in.ServiceName == "" || in.Price <= 0 || in.UserID == "" || in.StartYM == "" {
		return model.Subscription{}, fmt.Errorf("service_name, price (>0), user_id, start_date required")
	}
	uid, err := uuid.Parse(in.UserID)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("invalid user_id")
	}
	start, err := model.ParseYearMonth(in.StartYM)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("invalid start_date (YYYY-MM)")
	}
	var end *time.Time
	if in.EndYM != "" {
		t, err := model.ParseYearMonth(in.EndYM)
		if err != nil {
			return model.Subscription{}, fmt.Errorf("invalid end_date (YYYY-MM)")
		}
		end = &t
	}
	subs := model.Subscription{
		ID:          uuid.New(),
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      uid,
		StartDate:   start,
		EndDate:     end,
	}
	return s.repo.Create(ctx, subs)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in model.SubscriptionUpdate) (model.Subscription, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Subscription{}, err
	}
	if in.ServiceName != nil {
		cur.ServiceName = *in.ServiceName
	}
	if in.Price != nil {
		if *in.Price <= 0 {
			return model.Subscription{}, fmt.Errorf("price must be > 0")
		}
		cur.Price = *in.Price
	}
	if in.StartYM != nil {
		t, err := model.ParseYearMonth(*in.StartYM)
		if err != nil {
			return model.Subscription{}, fmt.Errorf("invalid start_date")
		}
		cur.StartDate = t
	}
	if in.EndYM != nil {
		if *in.EndYM == "" {
			cur.EndDate = nil
		} else {
			t, err := model.ParseYearMonth(*in.EndYM)
			if err != nil {
				return model.Subscription{}, fmt.Errorf("invalid end_date")
			}
			cur.EndDate = &t
		}
	}
	return s.repo.Update(ctx, cur)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, q model.ListQuery) ([]model.Subscription, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) Total(ctx context.Context, q model.TotalQuery) (int64, error) {
	if q.From.After(q.To) {
		return 0, fmt.Errorf("from must be <= to")
	}
	return s.repo.Total(ctx, q)
}
