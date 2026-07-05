package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"

	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/catalog-service/internal/config"
	"github.com/jackietana/ticket-platform/catalog-service/internal/domain"
	"github.com/minio/minio-go/v7"
)

const (
	BUCKET_NAME        = "posters"
	CREATE_EVENT_QUERY = `INSERT INTO events (id, category_id, title, description, poster_url, price_cents,
	total_slots, event_date, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	GET_EVENT_QUERY = `SELECT id, category_id, title, description, poster_url, price_cents, total_slots, event_date,
	created_at FROM events WHERE id=$1`
	LIST_EVENTS_QUERY = `SELECT id, category_id, title, description, poster_url, price_cents, total_slots, event_date,
	created_at FROM events WHERE category_id=$1`
)

type CatalogRepository struct {
	extEndpoint string
	db          *sql.DB
	fs          *minio.Client
}

func NewRepository(ctx context.Context, psql *sql.DB, minio *minio.Client, cfg config.MinioConfig) *CatalogRepository {
	return &CatalogRepository{
		extEndpoint: cfg.ExternalEndpoint,
		db:          psql,
		fs:          minio,
	}
}

func (r *CatalogRepository) InitStorage(ctx context.Context) error {
	err := r.fs.MakeBucket(ctx, BUCKET_NAME, minio.MakeBucketOptions{})
	if err != nil {
		exist, errBucketExist := r.fs.BucketExists(ctx, BUCKET_NAME)
		if errBucketExist == nil && exist {
			log.Printf("bucket %s already exists", BUCKET_NAME)
			return nil
		} else {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	log.Printf("successfully created %s", BUCKET_NAME)
	return nil
}

func (r *CatalogRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	_, err := r.db.ExecContext(ctx, CREATE_EVENT_QUERY, event.ID, *event.CategoryID, event.Title, event.Description,
		event.PosterUrl, event.PriceCents, event.TotalSlots, event.EventDate, event.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *CatalogRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	var event domain.Event
	err := r.db.QueryRowContext(ctx, GET_EVENT_QUERY, id).Scan(&event.ID, &event.CategoryID, &event.Title, &event.Description,
		&event.PosterUrl, &event.PriceCents, &event.TotalSlots, &event.EventDate, &event.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *CatalogRepository) ListEvents(ctx context.Context, categoryID *int64) ([]domain.Event, error) {
	if categoryID == nil {
		return nil, errors.New("category ID is nil")
	}

	events := make([]domain.Event, 0)
	rows, err := r.db.QueryContext(ctx, LIST_EVENTS_QUERY, *categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event domain.Event

		err := rows.Scan(&event.ID, &event.CategoryID, &event.Title, &event.Description, &event.PosterUrl,
			&event.PriceCents, &event.TotalSlots, &event.EventDate, &event.CreatedAt)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *CatalogRepository) UploadPoster(ctx context.Context, file io.Reader, size int64, contentType string) (string, error) {
	exts, err := mime.ExtensionsByType(contentType)
	ext := ".jpg"
	if err == nil && len(exts) > 0 {
		ext = exts[0]
	}

	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	_, err = r.fs.PutObject(ctx, BUCKET_NAME, objectName, file, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to put object to minio: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s/%s", r.extEndpoint, BUCKET_NAME, objectName)

	return fileURL, nil
}
