package services

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/dtos"
	"brainz/developersapi/internal/entity"
	"brainz/developersapi/internal/security"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func init() {
	// Security used by UserService (Argon2) requires config; use test defaults.
	security.LoadCfg(&config.AuthConfig{
		Pepper:        "test-pepper",
		Argon2Memory:  64 * 1024,
		Argon2Time:    1,
		Argon2Threads: 2,
	})
}

type mockUserRepo struct {
	byEmail map[string]*entity.DeveloperAccount
	byID    map[uuid.UUID]*entity.DeveloperAccount
	create  func(ctx context.Context, u *entity.DeveloperAccount) (*entity.DeveloperAccount, error)
	update  func(ctx context.Context, u *entity.DeveloperAccount) (*entity.DeveloperAccount, error)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.DeveloperAccount, error) {
	if m.byEmail != nil {
		if u, ok := m.byEmail[email]; ok {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepo) GetById(ctx context.Context, id uuid.UUID) (*entity.DeveloperAccount, error) {
	if m.byID != nil {
		if u, ok := m.byID[id]; ok {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepo) Create(ctx context.Context, u *entity.DeveloperAccount) (*entity.DeveloperAccount, error) {
	if m.create != nil {
		return m.create(ctx, u)
	}
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	return u, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *entity.DeveloperAccount) (*entity.DeveloperAccount, error) {
	if m.update != nil {
		return m.update(ctx, u)
	}
	return u, nil
}

type mockEmailSender struct {
	send func(ctx context.Context, devID uuid.UUID) (*entity.EmailConfirmationToken, error)
}

func (m *mockEmailSender) SendConfirmationEmail(ctx context.Context, devID uuid.UUID) (*entity.EmailConfirmationToken, error) {
	if m.send != nil {
		return m.send(ctx, devID)
	}
	return &entity.EmailConfirmationToken{
		ID:          uuid.New(),
		DeveloperID: devID,
		Token:       "test-token",
		NumbericCode: "123456",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
		CreatedAt:   time.Now(),
	}, nil
}

func TestUserService_Authenticate(t *testing.T) {
	ctx := context.Background()
	password := "password123456"
	salt := security.GetRandomSalt()
	hash := security.GetHashArgon2(password, salt)
	confirmed := time.Now()

	tests := []struct {
		name      string
		email     string
		password  string
		setupRepo func() *mockUserRepo
		wantErr   error
	}{
		{
			name:     "success",
			email:    "dev@test.com",
			password: password,
			setupRepo: func() *mockUserRepo {
				return &mockUserRepo{
					byEmail: map[string]*entity.DeveloperAccount{
						"dev@test.com": {
							ID:               uuid.New(),
							Email:            "dev@test.com",
							PasswordHash:     hash,
							Salt:             salt,
							EmailConfirmedAt: &confirmed,
						},
					},
				}
			},
			wantErr: nil,
		},
		{
			name:     "wrong_password",
			email:    "dev@test.com",
			password: "wrongpassword123",
			setupRepo: func() *mockUserRepo {
				return &mockUserRepo{
					byEmail: map[string]*entity.DeveloperAccount{
						"dev@test.com": {
							ID:               uuid.New(),
							Email:            "dev@test.com",
							PasswordHash:     hash,
							Salt:             salt,
							EmailConfirmedAt: &confirmed,
						},
					},
				}
			},
			wantErr: entity.ErrWrongCredentials,
		},
		{
			name:     "user_banned",
			email:    "banned@test.com",
			password: password,
			setupRepo: func() *mockUserRepo {
				banned := time.Now()
				return &mockUserRepo{
					byEmail: map[string]*entity.DeveloperAccount{
						"banned@test.com": {
							ID:               uuid.New(),
							Email:            "banned@test.com",
							PasswordHash:     hash,
							Salt:             salt,
							EmailConfirmedAt: &confirmed,
							BannedAt:         &banned,
						},
					},
				}
			},
			wantErr: entity.ErrUserBanned,
		},
		{
			name:     "email_not_confirmed",
			email:    "unconfirmed@test.com",
			password: password,
			setupRepo: func() *mockUserRepo {
				return &mockUserRepo{
					byEmail: map[string]*entity.DeveloperAccount{
						"unconfirmed@test.com": {
							ID:           uuid.New(),
							Email:        "unconfirmed@test.com",
							PasswordHash: hash,
							Salt:         salt,
							EmailConfirmedAt: nil,
						},
					},
				}
			},
			wantErr: entity.ErrEmailNotConfirmed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := NewUserService(repo, &mockEmailSender{})
			acc, err := svc.Authenticate(ctx, tt.email, tt.password)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.Email != tt.email {
				t.Errorf("account email = %s, want %s", acc.Email, tt.email)
			}
		})
	}
}

func TestUserService_RegistrateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepo{}
		svc := NewUserService(repo, &mockEmailSender{})
		dto := &dtos.RegisterDto{Email: "new@test.com", Password: "securepassword123"}
		ect, err := svc.RegistrateUser(ctx, dto)
		if err != nil {
			t.Fatalf("RegistrateUser: %v", err)
		}
		if ect == nil || ect.DeveloperID == uuid.Nil {
			t.Error("expected non-nil token with developer id")
		}
	})

	t.Run("email_already_exists", func(t *testing.T) {
		repo := &mockUserRepo{
			create: func(ctx context.Context, u *entity.DeveloperAccount) (*entity.DeveloperAccount, error) {
				return nil, entity.ErrEmailAlreadyExists
			},
		}
		svc := NewUserService(repo, &mockEmailSender{})
		dto := &dtos.RegisterDto{Email: "exists@test.com", Password: "securepassword123"}
		_, err := svc.RegistrateUser(ctx, dto)
		if err != entity.ErrEmailAlreadyExists {
			t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})
}

func TestUserService_ConfirmEmail(t *testing.T) {
	ctx := context.Background()
	devID := uuid.New()

	t.Run("success", func(t *testing.T) {
		updated := false
		repo := &mockUserRepo{
			byID: map[uuid.UUID]*entity.DeveloperAccount{
				devID: {
					ID:        devID,
					Email:     "u@test.com",
					EmailConfirmedAt: nil,
				},
			},
			update: func(ctx context.Context, u *entity.DeveloperAccount) (*entity.DeveloperAccount, error) {
				updated = true
				if u.EmailConfirmedAt == nil {
					t.Error("EmailConfirmedAt should be set")
				}
				return u, nil
			},
		}
		svc := NewUserService(repo, &mockEmailSender{})
		err := svc.ConfirmEmail(ctx, devID)
		if err != nil {
			t.Fatalf("ConfirmEmail: %v", err)
		}
		if !updated {
			t.Error("Update was not called")
		}
	})

	t.Run("user_not_found", func(t *testing.T) {
		repo := &mockUserRepo{byID: map[uuid.UUID]*entity.DeveloperAccount{}}
		svc := NewUserService(repo, &mockEmailSender{})
		err := svc.ConfirmEmail(ctx, uuid.New())
		if err == nil {
			t.Fatal("expected error when user not found")
		}
	})
}
