package main

import "testing"

// TestProductStoreListReturnsProductsByID 用于验证商品列表按 ID 升序返回；参数 t 管理测试状态；返回值为空。
func TestProductStoreListReturnsProductsByID(t *testing.T) {
	store := newProductStore()
	for id := int64(50); id >= 1; id-- {
		store.products[id] = Product{ID: id}
	}

	products := store.List()
	for index := 1; index < len(products); index++ {
		if products[index-1].ID > products[index].ID {
			t.Fatalf(
				"商品列表未按 ID 升序排列：位置 %d 的 ID 是 %d，下一项是 %d",
				index-1,
				products[index-1].ID,
				products[index].ID,
			)
		}
	}
}
