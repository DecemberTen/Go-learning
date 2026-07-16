package main

import (
	"sort"
	"sync"
	"time"
)

type ProductStore struct {
	mutex    sync.RWMutex
	products map[int64]Product
	nextID   int64
}

// newProductStore 用于创建初始化完成的商品存储；参数为空；返回 ProductStore 指针。
func newProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[int64]Product),
		nextID:   1,
	}
}

// List 用于查询全部商品；参数为空；返回按 ID 升序排列的商品值切片副本。
func (store *ProductStore) List() []Product {
	store.mutex.RLock()

	products := make([]Product, 0, len(store.products))
	for _, product := range store.products {
		products = append(products, product)
	}
	store.mutex.RUnlock()

	sort.Slice(products, func(first, second int) bool {
		return products[first].ID < products[second].ID
	})

	return products
}

// Get 用于根据商品 ID 查询商品；参数 id 是商品编号；返回商品和是否存在。
func (store *ProductStore) Get(id int64) (Product, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	product, exists := store.products[id]
	return product, exists
}

// Create 用于创建并保存商品；参数 input 是创建请求；返回创建后的完整商品。
func (store *ProductStore) Create(input CreateProductRequest) Product {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	now := time.Now().UTC()
	product := Product{
		ID:         store.nextID,
		Name:       input.Name,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		Status:     input.Status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	store.products[product.ID] = product
	store.nextID++

	return product
}

// Update 用于更新指定商品；参数 id 是商品编号，input 是更新请求；返回更新后的商品和是否存在。
func (store *ProductStore) Update(id int64, input UpdateProductRequest) (Product, bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	product, exists := store.products[id]
	if !exists {
		return Product{}, false
	}

	product.Name = input.Name
	product.PriceCents = input.PriceCents
	product.Stock = input.Stock
	product.Status = input.Status
	product.UpdatedAt = time.Now().UTC()
	store.products[id] = product

	return product, true
}

// Delete 用于删除指定商品；参数 id 是商品编号；返回商品是否原本存在。
func (store *ProductStore) Delete(id int64) bool {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.products[id]; !exists {
		return false
	}

	delete(store.products, id)
	return true
}
