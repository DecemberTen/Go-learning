package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProductHandlerRejectsInvalidJSON 用于验证商品写接口拒绝未知字段和多个 JSON；参数 t 管理测试状态；返回值为空。
func TestProductHandlerRejectsInvalidJSON(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "POST rejects unknown field",
			method: http.MethodPost,
			path:   "/products",
			body:   `{"name":"苹果","price_cents":100,"stock":10,"status":"active","extra":true}`,
		},
		{
			name:   "POST rejects multiple JSON values",
			method: http.MethodPost,
			path:   "/products",
			body:   `{"name":"苹果","price_cents":100,"stock":10,"status":"active"}{}`,
		},
		{
			name:   "PUT rejects unknown field",
			method: http.MethodPut,
			path:   "/products/1",
			body:   `{"name":"苹果","price_cents":100,"stock":10,"status":"active","extra":true}`,
		},
		{
			name:   "PUT rejects multiple JSON values",
			method: http.MethodPut,
			path:   "/products/1",
			body:   `{"name":"苹果","price_cents":100,"stock":10,"status":"active"}{}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newProductStore()
			handler := newProductHandler(store)
			mux := http.NewServeMux()
			handler.RegisterRoutes(mux)

			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("期望状态码 %d，实际为 %d", http.StatusBadRequest, response.Code)
			}
		})
	}
}
