package rest

import "fmt"

// ExampleHandler_GetMetricByPath показывает получение метрики по типу и имени.
func ExampleHandler_GetMetricByPath() {
	fmt.Println("GET /value/gauge/Alloc")
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")
	fmt.Println()
	fmt.Println("123.45")

	// Output:
	// GET /value/gauge/Alloc
	//
	// HTTP/1.1 200 OK
	//
	// 123.45
}

// ExampleHandler_GetMetric показывает получение метрики через JSON-запрос.
func ExampleHandler_GetMetric() {
	fmt.Println("POST /value/")
	fmt.Println("Content-Type: application/json")
	fmt.Println()
	fmt.Println(`{"id":"Alloc","type":"gauge"}`)
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")
	fmt.Println()
	fmt.Println(`{"id":"Alloc","type":"gauge","value":123.45}`)

	// Output:
	// POST /value/
	// Content-Type: application/json
	//
	// {"id":"Alloc","type":"gauge"}
	//
	// HTTP/1.1 200 OK
	//
	// {"id":"Alloc","type":"gauge","value":123.45}
}

// ExampleHandler_GetMetrics показывает получение HTML-страницы со всеми метриками.
func ExampleHandler_GetMetrics() {
	fmt.Println("GET /")
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")
	fmt.Println("Content-Type: text/html; charset=utf-8")
	fmt.Println()
	fmt.Println("<html>...</html>")

	// Output:
	// GET /
	//
	// HTTP/1.1 200 OK
	// Content-Type: text/html; charset=utf-8
	//
	// <html>...</html>
}

// ExampleHandler_UpdateMetricByPath показывает обновление метрики через URL.
func ExampleHandler_UpdateMetricByPath() {
	fmt.Println("POST /update/gauge/Alloc/123.45")
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")

	// Output:
	// POST /update/gauge/Alloc/123.45
	//
	// HTTP/1.1 200 OK
}

// ExampleHandler_UpdateMetric показывает обновление метрики через JSON.
func ExampleHandler_UpdateMetric() {
	fmt.Println("POST /update/")
	fmt.Println("Content-Type: application/json")
	fmt.Println()
	fmt.Println(`{"id":"Alloc","type":"gauge","value":123.45}`)
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")

	// Output:
	// POST /update/
	// Content-Type: application/json
	//
	// {"id":"Alloc","type":"gauge","value":123.45}
	//
	// HTTP/1.1 200 OK
}

// ExampleHandler_UpdateMetrics показывает пакетное обновление метрик.
func ExampleHandler_UpdateMetrics() {
	fmt.Println("POST /updates")
	fmt.Println("Content-Type: application/json")
	fmt.Println()
	fmt.Println(`[{"id":"Alloc","type":"gauge","value":123.45}, {"id":"PollCount","type":"counter","delta":5}]`)
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")

	// Output:
	// POST /updates
	// Content-Type: application/json
	//
	// [{"id":"Alloc","type":"gauge","value":123.45}, {"id":"PollCount","type":"counter","delta":5}]
	//
	// HTTP/1.1 200 OK
}

// ExampleHandler_Ping показывает проверку доступности сервиса.
func ExampleHandler_Ping() {
	fmt.Println("GET /ping")
	fmt.Println()
	fmt.Println("HTTP/1.1 200 OK")

	// Output:
	// GET /ping
	//
	// HTTP/1.1 200 OK
}
