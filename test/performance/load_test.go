package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// LoadTestConfig 负载测试配置
type LoadTestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	RequestsPerUser int
	TestDuration    time.Duration
}

// TestResult 测试结果
type TestResult struct {
	TotalRequests    int
	SuccessRequests  int
	FailedRequests   int
	AverageResponse  time.Duration
	MinResponse      time.Duration
	MaxResponse      time.Duration
	RequestsPerSec   float64
	ErrorRate        float64
}

// PerformanceTest 性能测试套件
type PerformanceTest struct {
	config    LoadTestConfig
	authToken string
	results   []time.Duration
	errors    []error
	mutex     sync.Mutex
}

func NewPerformanceTest(config LoadTestConfig) *PerformanceTest {
	return &PerformanceTest{
		config:  config,
		results: make([]time.Duration, 0),
		errors:  make([]error, 0),
	}
}

// TestEmployeeListPerformance 员工列表性能测试
func TestEmployeeListPerformance(t *testing.T) {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8081",
		ConcurrentUsers: 50,
		RequestsPerUser: 20,
		TestDuration:    30 * time.Second,
	}
	
	test := NewPerformanceTest(config)
	
	// 获取认证token
	if err := test.authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	
	// 执行负载测试
	result := test.runLoadTest("/api/v1/employees", "GET", nil)
	
	// 验证性能指标
	t.Logf("Performance Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-result.ErrorRate)*100)
	t.Logf("Average Response Time: %v", result.AverageResponse)
	t.Logf("Requests Per Second: %.2f", result.RequestsPerSec)
	
	// 性能断言
	if result.AverageResponse > 200*time.Millisecond {
		t.Errorf("Average response time too high: %v", result.AverageResponse)
	}
	
	if result.ErrorRate > 0.01 { // 1% 错误率阈值
		t.Errorf("Error rate too high: %.2f%%", result.ErrorRate*100)
	}
	
	if result.RequestsPerSec < 100 { // 最低100 RPS
		t.Errorf("Requests per second too low: %.2f", result.RequestsPerSec)
	}
}

// TestEmployeeCreatePerformance 员工创建性能测试
func TestEmployeeCreatePerformance(t *testing.T) {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8081",
		ConcurrentUsers: 20,
		RequestsPerUser: 10,
		TestDuration:    20 * time.Second,
	}
	
	test := NewPerformanceTest(config)
	
	// 获取认证token
	if err := test.authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	
	// 准备测试数据
	createData := map[string]interface{}{
		"user_id":     1,
		"employee_no": "EMP_PERF_TEST",
		"department":  "测试部门",
		"position":    "测试职位",
		"level":       "中级",
		"status":      "active",
		"max_tasks":   5,
	}
	
	// 执行负载测试
	result := test.runLoadTest("/api/v1/employees", "POST", createData)
	
	// 输出结果
	t.Logf("Employee Create Performance Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-result.ErrorRate)*100)
	t.Logf("Average Response Time: %v", result.AverageResponse)
	t.Logf("Requests Per Second: %.2f", result.RequestsPerSec)
}

// TestSkillManagementPerformance 技能管理性能测试
func TestSkillManagementPerformance(t *testing.T) {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8081",
		ConcurrentUsers: 30,
		RequestsPerUser: 15,
		TestDuration:    25 * time.Second,
	}
	
	test := NewPerformanceTest(config)
	
	// 获取认证token
	if err := test.authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	
	// 测试技能列表查询性能
	result := test.runLoadTest("/api/v1/skills", "GET", nil)
	
	t.Logf("Skill List Performance Results:")
	t.Logf("Average Response Time: %v", result.AverageResponse)
	t.Logf("Requests Per Second: %.2f", result.RequestsPerSec)
	t.Logf("Error Rate: %.2f%%", result.ErrorRate*100)
}

// authenticate 获取认证token
func (pt *PerformanceTest) authenticate() error {
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}
	
	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(pt.config.BaseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	data := result["data"].(map[string]interface{})
	pt.authToken = data["access_token"].(string)
	
	return nil
}

// runLoadTest 执行负载测试
func (pt *PerformanceTest) runLoadTest(endpoint, method string, data interface{}) TestResult {
	var wg sync.WaitGroup
	startTime := time.Now()
	
	// 启动并发用户
	for i := 0; i < pt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			pt.simulateUser(userID, endpoint, method, data)
		}(i)
	}
	
	wg.Wait()
	endTime := time.Now()
	
	return pt.calculateResults(startTime, endTime)
}

// simulateUser 模拟用户请求
func (pt *PerformanceTest) simulateUser(userID int, endpoint, method string, data interface{}) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	for i := 0; i < pt.config.RequestsPerUser; i++ {
		start := time.Now()
		
		// 准备请求数据
		var requestData interface{} = data
		if data != nil && method == "POST" {
			// 为每个请求生成唯一数据
			if empData, ok := data.(map[string]interface{}); ok {
				requestData = make(map[string]interface{})
				for k, v := range empData {
					requestData.(map[string]interface{})[k] = v
				}
				requestData.(map[string]interface{})["employee_no"] = fmt.Sprintf("EMP_%d_%d", userID, i)
			}
		}
		
		err := pt.makeRequest(client, endpoint, method, requestData)
		duration := time.Since(start)
		
		pt.mutex.Lock()
		pt.results = append(pt.results, duration)
		if err != nil {
			pt.errors = append(pt.errors, err)
		}
		pt.mutex.Unlock()
		
		// 添加小延迟模拟真实用户行为
		time.Sleep(10 * time.Millisecond)
	}
}

// makeRequest 发送HTTP请求
func (pt *PerformanceTest) makeRequest(client *http.Client, endpoint, method string, data interface{}) error {
	var body *bytes.Buffer
	
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	} else {
		body = bytes.NewBuffer(nil)
	}
	
	req, err := http.NewRequest(method, pt.config.BaseURL+endpoint, body)
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pt.authToken)
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	return nil
}

// calculateResults 计算测试结果
func (pt *PerformanceTest) calculateResults(startTime, endTime time.Time) TestResult {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	
	totalRequests := len(pt.results)
	failedRequests := len(pt.errors)
	successRequests := totalRequests - failedRequests
	
	if totalRequests == 0 {
		return TestResult{}
	}
	
	// 计算响应时间统计
	var totalDuration time.Duration
	minDuration := pt.results[0]
	maxDuration := pt.results[0]
	
	for _, duration := range pt.results {
		totalDuration += duration
		if duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}
	}
	
	avgDuration := totalDuration / time.Duration(totalRequests)
	testDuration := endTime.Sub(startTime)
	rps := float64(totalRequests) / testDuration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests)
	
	return TestResult{
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		AverageResponse: avgDuration,
		MinResponse:     minDuration,
		MaxResponse:     maxDuration,
		RequestsPerSec:  rps,
		ErrorRate:       errorRate,
	}
}

// BenchmarkEmployeeList 基准测试
func BenchmarkEmployeeList(b *testing.B) {
	config := LoadTestConfig{
		BaseURL: "http://localhost:8081",
	}
	
	test := NewPerformanceTest(config)
	if err := test.authenticate(); err != nil {
		b.Fatalf("Failed to authenticate: %v", err)
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			test.makeRequest(client, "/api/v1/employees", "GET", nil)
		}
	})
}
