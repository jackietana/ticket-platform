//go:build integration

package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jackietana/ticket-platform/auth-service/internal/config"
	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
	"github.com/jackietana/ticket-platform/auth-service/internal/dto"
	"github.com/jackietana/ticket-platform/auth-service/internal/repository"
	"github.com/jackietana/ticket-platform/auth-service/pkg/hash"
	pkgcache "github.com/jackietana/ticket-platform/pkg/cache"
	pkgpsql "github.com/jackietana/ticket-platform/pkg/database"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestIntegration_HappyPath(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewTestConfig()

	// add postgres and redis containers
	psqlContainer, err := postgres.Run(ctx,
		"postgres:alpine",
		postgres.WithDatabase(cfg.Postgres.Name),
		postgres.WithUsername(cfg.Postgres.User),
		postgres.WithPassword(cfg.Postgres.Pass),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to create psql container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(psqlContainer); err != nil {
			t.Errorf("failed to terminate psql container: %v", err)
		}
	})

	psqlPort, err := psqlContainer.MappedPort(ctx, cfg.Postgres.Port)
	if err != nil {
		t.Errorf("failed to get psql port: %v", err)
	}

	cfg.Postgres.Port = psqlPort.Port()

	redisContainer, err := tcredis.Run(ctx, "redis:alpine")
	if err != nil {
		t.Fatalf("failed to create redis container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Errorf("failed to terminate redis container: %v", err)
		}
	})

	// init dependencies
	psqlDB, err := pkgpsql.NewPostgresConnection(&cfg.Postgres)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	redisDB := pkgcache.NewRedisConnection(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port), cfg.Redis.Pass)

	authHash := hash.NewSHA1Hasher(cfg.Salt)
	authRepo := repository.NewRepository(psqlDB)
	authCache := repository.NewCache(redisDB)
	authService := NewAuthService(authHash, authRepo, authCache)

	if err := pkgpsql.RunUpMigrations(&cfg.Postgres, "auth"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Run("Happy path", func(t *testing.T) {
		// STEP 1: registration (SignUp)
		// create and save user into db
		user := dto.UserRequest{Email: "test_user1@mail.com", Password: "password_123"}

		id, err := authService.SignUp(context.Background(), user)
		if err != nil {
			t.Errorf("faield to save user into db: %v", err)
		}

		// get saved user from db
		db, err := sql.Open("postgres", cfg.Postgres.GetDatabaseConnString())
		if err != nil {
			t.Errorf("failed to connect to db: %v", err)
		}

		var queryUser domain.User
		err = db.QueryRow("SELECT id, email, password, created_at FROM users WHERE email=$1", user.Email).
			Scan(&queryUser.ID, &queryUser.Email, &queryUser.Password, &queryUser.CreatedAt)
		if err != nil {
			t.Errorf("failed to execute query: %v", err)
		}

		// varidy user from db
		if id != queryUser.ID {
			t.Errorf("invalid user ID, want: %s, got: %s", id, queryUser.ID)
		}

		if user.Email != queryUser.Email {
			t.Errorf("invalid user email, want: %s, got: %s", user.Email, queryUser.Email)
		}

		if user.Password == queryUser.Password {
			t.Error("invalid user password, it should be hashed")
		}

		// STEP 2: Login (SignIn)
		// get token by credentials
		clientIP := "192.168.1.1"
		userAgent := "Mozilla/5.0"
		token, err := authService.SignIn(ctx, user, clientIP, userAgent)
		if err != nil {
			t.Errorf("failed to get token by credentials: %v", err)
		}

		// validate presence of token in cache
		session, err := authService.cache.GetSessionContext(ctx, token)
		if err != nil {
			t.Errorf("failed to get session from cache: %v", err)
		}

		// validate session from cache
		if clientIP != session.ClientIP {
			t.Errorf("invalid client IP, want: %s, got: %s", clientIP, session.ClientIP)
		}

		if userAgent != session.UserAgent {
			t.Errorf("invalid user agent, want: %s, got: %s", userAgent, session.UserAgent)
		}

		// STEP 3: Session validation
		// get and validate obtained user ID
		userID, err := authService.ValidateSession(ctx, token, clientIP, userAgent)
		if err != nil {
			t.Errorf("failed to validate session: %v", err)
		}

		if session.UserID != userID {
			t.Errorf("invalid user ID, want: %s, got: %s", session.UserID, userID)
		}

		// STEP 4: Validate sessions limitation
		// imitating multiple logins
		for i := range repository.MAX_SESSIONS {
			exampleUA := fmt.Sprintf("example_%d", i)
			_, err := authService.SignIn(ctx, user, clientIP, exampleUA)
			if err != nil {
				t.Errorf("failed to login: %v", err)
			}
		}

		// validate number of active sessions
		sessionsKey := fmt.Sprintf(repository.SESSIONS_PLACEHOLDER, session.UserID)
		sessionsCount, err := redisDB.ZCard(ctx, sessionsKey).Result()

		if repository.MAX_SESSIONS != sessionsCount {
			t.Errorf("invalid sessions count, want: %d, got: %d", repository.MAX_SESSIONS, sessionsCount)
		}

		// validate absence of invalid session
		_, err = redisDB.ZScore(ctx, sessionsKey, token).Result()
		if err != redis.Nil {
			t.Errorf("extra session should not present in cache")
		}
	})
}
