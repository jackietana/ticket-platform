package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackietana/ticket-platform/auth-service/internal/dto"
	mock_service "github.com/jackietana/ticket-platform/auth-service/internal/transport/rest/v1/mocks"
	"go.uber.org/mock/gomock"
)

func TestTransport_signUp(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuthService, user dto.UserRequest)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           dto.UserRequest
		mockFn              mockBehavior
		expectedStatusCode  int
		expectedRequestBody any
	}{
		{
			name:      "OK",
			inputBody: `{"email": "test@mail.com","password": "test_password"}`,
			inputUser: dto.UserRequest{Email: "test@mail.com", Password: "test_password"},
			mockFn: func(s *mock_service.MockAuthService, user dto.UserRequest) {
				s.EXPECT().SignUp(gomock.Any(), user).Return("1", nil)
			},
			expectedStatusCode:  http.StatusCreated,
			expectedRequestBody: objToJsonStr(dto.SignUpResponse{ID: "1", Message: "successfully signed up"}),
		},
		{
			name:                "Empty required field",
			inputBody:           `{"email": "test@mail.com"}`,
			inputUser:           dto.UserRequest{Email: "test@mail.com", Password: "test_password"},
			mockFn:              func(s *mock_service.MockAuthService, user dto.UserRequest) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: objToJsonStr(dto.ErrorResponse{Error: "invalid input body"}),
		},
		{
			name:      "Service failure",
			inputBody: `{"email": "test@mail.com","password": "test_password"}`,
			inputUser: dto.UserRequest{Email: "test@mail.com", Password: "test_password"},
			mockFn: func(s *mock_service.MockAuthService, user dto.UserRequest) {
				s.EXPECT().SignUp(gomock.Any(), user).Return("0", errors.New("service failure"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: objToJsonStr(dto.ErrorResponse{Error: "service failure"}),
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			// Init deps
			controller := gomock.NewController(t)
			defer controller.Finish()

			auth := mock_service.NewMockAuthService(controller)
			test.mockFn(auth, test.inputUser)
			handler := NewHandler(auth)

			// Test server
			router := gin.New()
			router.POST("/sign-up", handler.signUp)

			// Test request
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/sign-up", bytes.NewBufferString(test.inputBody))

			// Perform request
			router.ServeHTTP(recorder, request)

			// Assert
			if test.expectedStatusCode != recorder.Code {
				t.Errorf("invalid status code, expected: %d got: %d", test.expectedStatusCode, recorder.Code)
			}

			if test.expectedRequestBody != recorder.Body.String() {
				t.Errorf("invalid request body, expected: %s got: %s", test.expectedRequestBody, recorder.Body.String())
			}
		})
	}
}

func objToJsonStr(object any) string {
	bytesObj, err := json.Marshal(object)
	if err != nil {
		return ""
	}

	return string(bytesObj)
}
