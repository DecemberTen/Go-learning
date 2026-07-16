package main

import (
	"sync"
)

type ProductStore struct {
	mutex    sync.RWMutex
	products map[int64]Product
	nextID   int64
}

// newProductStore 用于创建初始化完成的内存商品仓库；参数为空；返回 ProductStore 指针。
func newProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[int64]Product),
		nextID:   1,
	}
}

func (store *ProductStore) Create(input CreateProductRequest) Product {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	product := Product{
		ID:         store.nextID,
		Name:       input.Name,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		Status:     input.Status,
	}
	store.products[store.nextID] = product
	store.nextID++
	return product
}

func (store *ProductStore) List() []Product {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	products := make([]Product, 0, len(store.products))
	for _, product := range store.products {
		products = append(products, product)
	}
	return products
}

func (store *ProductStore) Get(id int64) (Product, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	product, exists := store.products[id]
	return product, exists
}

func (store *ProductStore) Update(id int64, input UpdateProductRequest) (Product, bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	products, exits := store.products[id]
	if !exits {
		return Product{}, false
	}
	products.Name = input.Name
	products.PriceCents = input.PriceCents
	products.Stock = input.Stock
	products.Status = input.Status
	store.products[id] = products
	return products, true
}

func (store *ProductStore) Delete(id int64) bool {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	_, exists := store.products[id]
	if !exists {
		return false
	}
	delete(store.products, id)
	return true
}
