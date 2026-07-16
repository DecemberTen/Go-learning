package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type ProductHandler struct {
	store *ProductStore
}

// newProductHandler 用于创建商品 HTTP Handler；参数 store 是商品存储；返回 ProductHandler 指针。
func newProductHandler(store *ProductStore) *ProductHandler {
	return &ProductHandler{store: store}
}

// RegisterRoutes 用于注册商品 CRUD 路由；参数 mux 是 HTTP 路由器；返回值为空。
func (handler *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /products", handler.listProducts)
	mux.HandleFunc("GET /products/{id}", handler.getProduct)
	mux.HandleFunc("POST /products", handler.createProduct)
	mux.HandleFunc("PUT /products/{id}", handler.updateProduct)
	mux.HandleFunc("DELETE /products/{id}", handler.deleteProduct)
}

// listProducts 用于查询商品列表；参数 response 和 request 表示 HTTP 响应与请求；返回值为空。
func (handler *ProductHandler) listProducts(response http.ResponseWriter, request *http.Request) {
	products := handler.store.List()
	writeJSON(response, http.StatusOK, products)
}

// getProduct 用于查询指定商品；参数 response 和 request 表示 HTTP 响应与请求；返回值为空。
func (handler *ProductHandler) getProduct(response http.ResponseWriter, request *http.Request) {
	idText := request.PathValue("id")

	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		writeError(response, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, exists := handler.store.Get(int64(id))
	if !exists {
		writeError(response, http.StatusNotFound, "Product not found")
		return
	}

	writeJSON(response, http.StatusOK, product)
}

// createProduct 用于创建商品；参数 response 和 request 表示 HTTP 响应与请求；返回值为空。
func (handler *ProductHandler) createProduct(response http.ResponseWriter, request *http.Request) {
	input := CreateProductRequest{}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&input)
	if err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = validateProductInput(input.Name, input.PriceCents, input.Stock, input.Status)
	if err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}

	createdProduct := handler.store.Create(input)
	writeJSON(response, http.StatusCreated, createdProduct)
}

// updateProduct 用于完整更新指定商品；参数 response 和 request 表示 HTTP 响应与请求；返回值为空。
func (handler *ProductHandler) updateProduct(response http.ResponseWriter, request *http.Request) {
	idText := request.PathValue("id")

	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		writeError(response, http.StatusBadRequest, "Invalid product ID")
		return
	}

	input := UpdateProductRequest{}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&input)
	if err != nil {
		writeError(response, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(response, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = validateProductInput(input.Name, input.PriceCents, input.Stock, input.Status)
	if err != nil {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}

	product, exists := handler.store.Update(int64(id), input)
	if !exists {
		writeError(response, http.StatusNotFound, "Product not found")
		return
	}

	writeJSON(response, http.StatusOK, product)
}

// deleteProduct 用于删除指定商品；参数 response 和 request 表示 HTTP 响应与请求；返回值为空。
func (handler *ProductHandler) deleteProduct(response http.ResponseWriter, request *http.Request) {
	idText := request.PathValue("id")

	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		writeError(response, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if !handler.store.Delete(int64(id)) {
		writeError(response, http.StatusNotFound, "Product not found")
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
