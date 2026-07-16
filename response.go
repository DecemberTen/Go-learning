package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON 用于返回统一的 JSON 响应；参数 response 是响应写入器、status 是 HTTP 状态码、data 是响应数据；返回值为空。
func writeJSON(response http.ResponseWriter, status int, data any) {
	response.Header().Set("Content-Type", "application/json;charset=utf-8")
	response.WriteHeader(status)

	err := json.NewEncoder(response).Encode(data)
	if err != nil {
		fmt.Println(err)
	}
}

// writeError 用于返回统一的 JSON 错误响应；参数 response 是响应写入器、status 是 HTTP 状态码、message 是错误信息；返回值为空。
func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, errorResponse{Error: message})
}
