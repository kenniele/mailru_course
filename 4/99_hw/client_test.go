package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFindUsers(t *testing.T) {
	// Поднимаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "token",
		URL:         ts.URL,
	}

	// Список тестов
	tests := []struct {
		name    string
		request SearchRequest
		wantErr bool
		wantLen int
	}{
		{
			name:    "Valid request",
			request: SearchRequest{Query: "John", Limit: 5, Offset: 0, OrderField: "Name", OrderBy: OrderByAsc},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "Invalid order field",
			request: SearchRequest{OrderField: "InvalidField"},
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "Negative limit",
			request: SearchRequest{Limit: -1},
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "Negative offset",
			request: SearchRequest{Offset: -1},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Выполняем запрос через SearchClient
			resp, err := client.FindUsers(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v, wantErr: %v", err, tt.wantErr)
			}
			if err == nil && len(resp.Users) != tt.wantLen {
				t.Errorf("unexpected result length: got %v, want %v", len(resp.Users), tt.wantLen)
			}
		})
	}
}
