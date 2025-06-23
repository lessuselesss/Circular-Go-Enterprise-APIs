package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com")

	if client.baseURL != "https://api.example.com" {
		t.Errorf("NewClient baseURL: expected %q, got %q", "https://api.example.com", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("NewClient should initialize httpClient")
	}

	if client.timeout == 0 {
		t.Error("NewClient should set a default timeout")
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client := NewClient("https://api.example.com")

	client.SetTimeout(5 * time.Second)

	if client.timeout != 5*time.Second {
		t.Errorf("SetTimeout: expected %v, got %v", 5*time.Second, client.timeout)
	}
}

func TestClient_POST_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	payload := map[string]string{"test": "data"}

	response, err := client.POST(context.Background(), "/test", payload)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}

	var result map[string]string
	err = json.Unmarshal(response, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got %q", result["status"])
	}
}

func TestClient_POST_InvalidJSON(t *testing.T) {
	client := NewClient("https://api.example.com")

	// Test with invalid JSON payload (channel cannot be marshaled)
	payload := make(chan int)

	_, err := client.POST(context.Background(), "/test", payload)
	if err == nil {
		t.Error("POST should return error for invalid JSON payload")
	}
}

func TestClient_POST_ServerError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	payload := map[string]string{"test": "data"}

	_, err := client.POST(context.Background(), "/test", payload)
	if err == nil {
		t.Error("POST should return error for server error response")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Error should mention status code 500, got: %v", err)
	}
}

func TestClient_POST_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetTimeout(50 * time.Millisecond) // Short timeout

	payload := map[string]string{"test": "data"}

	_, err := client.POST(context.Background(), "/test", payload)
	if err == nil {
		t.Error("POST should return timeout error")
	}
}

func TestClient_GET_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "test"})
	}))
	defer server.Close()

	client := NewClient(server.URL)

	response, err := client.GET(context.Background(), "/test")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}

	var result map[string]string
	err = json.Unmarshal(response, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["data"] != "test" {
		t.Errorf("Expected data 'test', got %q", result["data"])
	}
}

func TestClient_GET_NotFound(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GET(context.Background(), "/nonexistent")
	if err == nil {
		t.Error("GET should return error for 404 response")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Error should mention status code 404, got: %v", err)
	}
}

func TestClient_WithRetry(t *testing.T) {
	callCount := 0
	// Create a test server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetRetryAttempts(3)
	client.SetRetryDelay(10 * time.Millisecond)

	payload := map[string]string{"test": "data"}

	response, err := client.POST(context.Background(), "/test", payload)
	if err != nil {
		t.Fatalf("POST with retry should succeed: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls (2 failures + 1 success), got %d", callCount)
	}

	var result map[string]string
	err = json.Unmarshal(response, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got %q", result["status"])
	}
}

func TestClient_RetryExhausted(t *testing.T) {
	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetRetryAttempts(2)
	client.SetRetryDelay(10 * time.Millisecond)

	payload := map[string]string{"test": "data"}

	_, err := client.POST(context.Background(), "/test", payload)
	if err == nil {
		t.Error("POST should fail after retry attempts exhausted")
	}
}
