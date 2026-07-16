package goroutine

import (
	"fmt"
	"sync"
)

func StartGoroutine(w *sync.WaitGroup) {
	for i := 0; i < 5; i++ {
		w.Add(1)
		go func(i int) {
			defer w.Done()
			fmt.Println("正在处理任务", i)
			fmt.Println("任务", i, "处理完成")
		}(i)
	}
}
