package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"
)

func calcuate(channel chan<- int) {
	for i := 0; i < 5; i++ {
		channel <- i * i
	}
	close(channel)
}

func sendText(text string, chanel chan<- string) {
	chanel <- text
}

func sell(stock *int, w *sync.WaitGroup, mutex *sync.Mutex) {
	defer w.Done()
	mutex.Lock()
	*stock -= 3
	defer mutex.Unlock()
}

func processOrder() {
	fmt.Println("正在处理订单")

	defer fmt.Println("处理结束")
	defer fmt.Println("扣减库存")
	defer fmt.Println("检查库存")
}

var ErrInsufficientBalance = errors.New("余额不足")

func Withdraw(amount int, stock int) (int, error) {
	if amount <= 0 {
		return stock, errors.New("金额不能小于等于0")
	}
	if amount > stock {
		return stock, fmt.Errorf("余额不足，当前余额为%d，提现金额为%d, %w", stock, amount, ErrInsufficientBalance)
	}
	return stock - amount, nil
}

type ProductID int64

func (p ProductID) String() string {
	return fmt.Sprintf("PRODUCT-%d", p)
}

// type ProductStatus string

// const ProductStatusActive ProductStatus = "active"
// const ProductStatusSoldOut ProductStatus = "sold_out"
// const ProductStatusDisabled ProductStatus = "disabled"

// type Product struct {
// 	id     ProductID
// 	status ProductStatus
// }

func describe(value any) {
	switch t := value.(type) {
	case string:
		fmt.Printf("value: %s,字符数量%d\n", t, utf8.RuneCountInString(t))
	case int:
		fmt.Printf("value: %d,两倍为%d\n", t, t*2)
	case bool:
		fmt.Printf("value: %v,相反值为%v\n", t, !t)
	default:
		fmt.Printf("value: %T\n", t)
	}

}

func average(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	sum := float64(0)
	for _, num := range numbers {
		sum += num
	}
	return sum / float64(len(numbers))
}

func createMultiplier(factor int) func(int) int {
	n := factor
	return func(num int) int {
		return num * n
	}
}

type Address struct {
	City   string
	Street string
}

func (address Address) FullAddress() string {
	return fmt.Sprintf("%s, %s", address.City, address.Street)
}

type Customer struct {
	Name string
	Address
}

func writeNote(filename, content string) error {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("写入%s失败,%w", filename, err)
	}
	return nil
}

func readNote(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("读取%s失败,%w", filename, err)
	}
	return string(content), nil
}

type Product1 struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Tags        []string `json:"tags,omitempty"`
	InternalKey string   `json:"-"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Cl struct {
	Category string `json:"category"`
	Limit    int    `json:"limit"`
}

// type CreateProductRequest struct {
// 	Name  string  `json:"name"`
// 	Price float64 `json:"price"`
// }

// type Product struct {
// 	ID    int     `json:"id"`
// 	Name  string  `json:"name"`
// 	Price float64 `json:"price"`
// }

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Println("tokenMiddleware")
		token := request.Header.Get("Authorization")
		if token != "Bearer abc123" {
			http.Error(response, "Invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func processTask(ctx context.Context) error {

	select {
	case <-time.After(3 * time.Second):
		fmt.Println("任务完成")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// main 用于创建商品存储、注册 CRUD 路由并启动 HTTP 服务；参数为空；返回值为空。
func main() {
	// fmt.Println("hello world")
	// fmt.Println(calculator.Add(1, 2))
	// w := &sync.WaitGroup{}
	// goroutine.StartGoroutine(w)
	// w.Wait()
	// fmt.Println("所有任务处理完成")

	// channel := make(chan int)

	// go calcuate(channel)

	// for result := range channel {
	// 	fmt.Println(result)
	// }

	// orderChannel := make(chan string)
	// paymentChannel := make(chan string)

	// go sendText("订单创建完成", orderChannel)
	// go sendText("支付成功", paymentChannel)

	// for i := 0; i < 2; i++ {
	// 	select {
	// 	case text := <-orderChannel:
	// 		fmt.Println(text)
	// 	case text := <-paymentChannel:
	// 		fmt.Println(text)
	// 	}
	// }

	// var waitgroup sync.WaitGroup
	// var mutex sync.Mutex

	// stock := 100

	// for i := 0; i < 10; i++ {
	// 	waitgroup.Add(1)
	// 	go sell(&stock, &waitgroup, &mutex)
	// }
	// waitgroup.Wait()
	// fmt.Println(stock)

	// text := "你好Go"

	// fmt.Printf("字节数量%d\n", len(text))
	// fmt.Printf("字符数量%d\n", utf8.RuneCountInString(text))

	// for index, char := range text {
	// 	fmt.Printf("索引%d,字符%c\n", index, char)
	// }

	// r := []rune(text)
	// r[0] = '我'

	// text = string(r)
	// fmt.Println(text)

	// processOrder()

	// stock := 10

	// stock, err := Withdraw(20, stock)
	// if err != nil {
	// 	if errors.Is(err, ErrInsufficientBalance) {
	// 		fmt.Println("补充余额")
	// 	}
	// 	fmt.Println(err)
	// }
	// fmt.Println("当前余额:", stock)

	// product := Product{
	// 	id:     ProductID(1),
	// 	status: ProductStatusActive,
	// }

	// if product.status == ProductStatusActive {
	// 	fmt.Println("商品状态为活跃")
	// }

	// var numbers []int
	// numbers = append(numbers, 10, 20, 30)
	// fmt.Println(numbers)

	// var m1 map[string]int
	// if m1 != nil {
	// 	fmt.Println(m1)
	// } else {
	// 	fmt.Println("m为nil")
	// }

	// m2 := make(map[string]int)
	// m2["GO"] = 100
	// fmt.Println(m2)

	// num := new(int)
	// *num = 50
	// fmt.Println(*num)

	// fmt.Println(average(10, 20, 30))

	// numbers := []float64{80, 90, 100}
	// fmt.Println(average(numbers...))

	// double := createMultiplier(2)
	// triple := createMultiplier(3)

	// fmt.Println(double(10)) // 20
	// fmt.Println(triple(10)) // 30

	// customer := &Customer{
	// 	Name: "张三",
	// 	Address: Address{
	// 		City:   "北京",
	// 		Street: "东城区",
	// 	},
	// }
	// fmt.Println(customer.FullAddress())

	// err := writeNote("note.txt", "今天学习了 Go 文件操作")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("写入成功")
	// // 读取文件
	// content, err := readNote("note.txt")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("读取成功:", content)

	// product := Product1{
	// 	ID:          1,
	// 	Name:        "商品1",
	// 	Price:       100.0,
	// 	Tags:        []string{"标签1", "标签2"},
	// 	InternalKey: "内部键",
	// }

	// data, err := json.Marshal(product)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(data))

	// var product2 Product1

	// err = json.Unmarshal(data, &product2)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(product2.Name, product2.Price)

	// data := `{"username":"admin","password":"123456","extra":"字段"}`

	// reader := strings.NewReader(data)
	// decoder := json.NewDecoder(reader)
	// decoder.DisallowUnknownFields()

	// var loginRequest LoginRequest

	// err := decoder.Decode(&loginRequest)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Printf("用户名:%s\n", loginRequest.Username)
	// fmt.Printf("密码:%s\n", loginRequest.Password)

	// mux := http.NewServeMux()

	// mux.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
	// 	fmt.Fprintf(response, "Hello, World!")
	// })
	// mux.HandleFunc("/about", func(response http.ResponseWriter, request *http.Request) {
	// 	fmt.Fprintf(response, "关于我们")
	// })
	// mux.HandleFunc("/users", func(response http.ResponseWriter, request *http.Request) {
	// 	if request.Method != http.MethodGet {
	// 		response.Header().Set("Allow", http.MethodGet)
	// 		http.Error(response, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 		return
	// 	}

	// 	users := []User{
	// 		{
	// 			ID:   1,
	// 			Name: "张三",
	// 		},
	// 		{
	// 			ID:   2,
	// 			Name: "李四",
	// 		},
	// 	}

	// 	response.Header().Set("Content-Type", "application/json;charset=utf-8")
	// 	response.WriteHeader(http.StatusOK)
	// 	err := json.NewEncoder(response).Encode(users)
	// 	if err != nil {
	// 		http.Error(response, "Internal Server Error", http.StatusInternalServerError)
	// 	}
	// })

	// mux.HandleFunc("/products", func(response http.ResponseWriter, request *http.Request) {
	// 	if request.Method != http.MethodGet {
	// 		response.Header().Set("Allow", http.MethodGet)
	// 		http.Error(response, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 		return
	// 	}

	// 	query := request.URL.Query()
	// 	category := query.Get("category")
	// 	limitText := query.Get("limit")
	// 	limit := 20
	// 	if limitText != "" {
	// 		limit, err := strconv.Atoi(limitText)
	// 		if err != nil || limit <= 0 {
	// 			http.Error(response, "Invalid limit", http.StatusBadRequest)
	// 			return
	// 		}
	// 	}

	// 	result := Cl{
	// 		Category: category,
	// 		Limit:    limit,
	// 	}

	// 	response.Header().Set("Content-Type", "application/json;charset=utf-8")
	// 	response.WriteHeader(http.StatusOK)

	// 	err := json.NewEncoder(response).Encode(result)
	// 	if err != nil {
	// 		fmt.Printf("编码失败")
	// 	}
	// })

	// mux.HandleFunc("/products", func(response http.ResponseWriter, request *http.Request) {
	// 	if request.Method != http.MethodPost {
	// 		response.Header().Set("Allow", http.MethodPost)
	// 		http.Error(response, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 		return
	// 	}

	// 	var input CreateProductRequest

	// 	error := json.NewDecoder(request.Body).Decode(&input)
	// 	if error != nil {
	// 		http.Error(response, "Invalid request body", http.StatusBadRequest)
	// 		return
	// 	}
	// 	if input.Name == "" {
	// 		http.Error(response, "Name is required", http.StatusBadRequest)
	// 		return
	// 	}
	// 	if input.Price <= 0 {
	// 		http.Error(response, "Price must be greater than 0", http.StatusBadRequest)
	// 		return
	// 	}

	// 	product := Product{
	// 		ID:    1,
	// 		Name:  input.Name,
	// 		Price: input.Price,
	// 	}

	// 	response.Header().Set("Content-Type", "application/json;charset=utf-8")

	// 	err := json.NewEncoder(response).Encode(product)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// })

	// mux.HandleFunc("GET /users/{id}", func(response http.ResponseWriter, request *http.Request) {
	// 	idText := request.PathValue("id")

	// 	userId, err := strconv.Atoi(idText)
	// 	if err != nil || userId <= 0 {
	// 		http.Error(response, "Invalid user ID", http.StatusBadRequest)
	// 		return
	// 	}

	// 	user := User{
	// 		ID:   userId,
	// 		Name: "张三",
	// 	}

	// 	response.Header().Set("Content-Type", "application/json;charset=utf-8")
	// 	response.WriteHeader(http.StatusOK)
	// 	err = json.NewEncoder(response).Encode(user)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// })

	// mux.HandleFunc("GET /profile", func(response http.ResponseWriter, request *http.Request) {
	// 	fmt.Printf("profile")
	// })

	// fmt.Println("已启动")

	// handler := tokenMiddleware(mux)

	// err := http.ListenAndServe(":8080", handler)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// ctx, cancel := context.WithTimeout(
	// 	context.Background(),
	// 	5*time.Second,
	// )
	// defer cancel()

	// err := processTask(ctx)
	// if err != nil {
	// 	if errors.Is(err, context.DeadlineExceeded) {
	// 		fmt.Println("任务超时")
	// 	} else {
	// 		fmt.Println(err)
	// 	}
	// }

	store := newProductStore()
	handler := newProductHandler(store)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	serverError := make(chan error, 1)

	go func() {
		serverError <- server.ListenAndServe()
	}()

	signalContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("服务器启动失败:", err)
		}
		return
	case <-signalContext.Done():
		fmt.Println("正在关闭服务器")
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	err := server.Shutdown(shutdownContext)
	if err != nil {
		fmt.Println("服务器关闭失败:", err)
		return
	}

	err = <-serverError
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("服务器运行失败:", err)
		return
	}

	fmt.Println("服务器已关闭")
}
