package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackietana/ticket-platform/auth-service/internal/dto"
	mock_service "github.com/jackietana/ticket-platform/auth-service/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestService_SignUp(t *testing.T) {
	t.Parallel()
	type mockBehavior func(hash *mock_service.MockHasher, repo *mock_service.MockRepository,
		user dto.UserRequest)

	testTable := []struct {
		name      string
		inputUser dto.UserRequest
		mockFn    mockBehavior
		want      string
		wantErr   bool
	}{
		{
			name:      "OK",
			inputUser: dto.UserRequest{Email: "test@mail.com", Password: "test_password"},
			mockFn: func(hash *mock_service.MockHasher, repo *mock_service.MockRepository, user dto.UserRequest) {
				hash.EXPECT().Hash(user.Password).Return("hashed_password", nil)
				repo.EXPECT().CreateUser(gomock.Any(), user.Email, "hashed_password").Return("1", nil)
			},
			want: "1",
		},
		{
			name:      "Hasher failure",
			inputUser: dto.UserRequest{Email: "test@mail.com", Password: "test_password"},
			mockFn: func(hash *mock_service.MockHasher, repo *mock_service.MockRepository, user dto.UserRequest) {
				hash.EXPECT().Hash(user.Password).Return("", errors.New("hasher failure"))
			},
			wantErr: true,
			want:    "",
		},
		{
			name:      "Repository failure",
			inputUser: dto.UserRequest{Email: "test@mail.com", Password: "test_password"},
			mockFn: func(hash *mock_service.MockHasher, repo *mock_service.MockRepository, user dto.UserRequest) {
				hash.EXPECT().Hash(user.Password).Return("hashed_password", nil)
				repo.EXPECT().CreateUser(gomock.Any(), user.Email, "hashed_password").
					Return("", errors.New("repository failure"))
			},
			wantErr: true,
			want:    "",
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			controller := gomock.NewController(t)
			hash := mock_service.NewMockHasher(controller)
			repo := mock_service.NewMockRepository(controller)
			test.mockFn(hash, repo, test.inputUser)
			service := NewAuthService(hash, repo, nil)

			res, err := service.SignUp(context.Background(), test.inputUser)

			if test.want != res {
				t.Errorf("invalid result, expected: %s, got: %s", test.want, res)
			}

			if test.wantErr {
				if err == nil {
					t.Errorf("invalid error, expected: error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("invalid error, expected: nil, got: %v", err)
				}
			}
		})
	}
}
