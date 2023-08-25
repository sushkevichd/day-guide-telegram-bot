package weather

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type Fetcher interface {
	FetchLatestByLocation(context.Context, domain.Location) (*domain.Weather, error)
}

type TableFormatter interface {
	Format(weather domain.Weather) string
}

type broadcasterService struct {
	fetcher           Fetcher
	locations         []domain.Location
	tableFormatter    TableFormatter
	outCh             chan<- string
	broadcastInterval time.Duration
}

func NewBroadcasterService(
	fetcher Fetcher,
	locations []domain.Location,
	tableFormatter TableFormatter,
	outCh chan<- string,
) (*broadcasterService, error) {
	return &broadcasterService{
		fetcher:           fetcher,
		locations:         locations,
		tableFormatter:    tableFormatter,
		outCh:             outCh,
		broadcastInterval: 3 * time.Hour,
	}, nil
}

func (_ *broadcasterService) Name() string { return "weather broadcaster" }

func (b *broadcasterService) Run(ctx context.Context) error {
	slog.Info("starting weather broadcaster service", "interval", b.broadcastInterval.String())
	defer slog.Info("stopped weather broadcaster service")

	for {
		if err := b.broadcast(ctx); err != nil {
			slog.Error("weather broadcaster pass failed", logger.Err(err))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(b.broadcastInterval):
			continue
		}
	}
}

func (b *broadcasterService) broadcast(ctx context.Context) error {
	slog.Info("starting weather broadcaster pass")
	startAt := time.Now()

	for _, loc := range b.locations {
		weather, err := b.fetcher.FetchLatestByLocation(ctx, loc)
		if err != nil {
			return fmt.Errorf("fetching weather for location %s: %w", loc, err)
		}

		tableStr := b.tableFormatter.Format(*weather)
		b.outCh <- tableStr
	}

	slog.Info("completed weather broadcaster pass", "elapsed_time", time.Now().Sub(startAt).String())
	return nil
}
