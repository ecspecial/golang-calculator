package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "calculatorapi/utility/models"
    "encoding/json"
)

func TestPingServers(t *testing.T) {
    // Создание мок-сервера, который отвечает, как если бы он был настоящим сервером расчетов
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/ping" {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "status": "running",
                "maxGoroutines": 10,
                "currentGoroutines": 5,
            })
        }
    }))
    defer server.Close()

    // Замена среза серверов на URL мок-сервера
    servers = []string{server.URL}

    // Вызов функции, подлежащей тестированию
    statuses := pingServers()

    // Проверка, правильно ли отрапортированы статусы
    if len(statuses) != 1 || !statuses[0].Running {
        t.Errorf("Expected the server to be running but got %v", statuses[0])
    }
}

func TestSubmitCalculations(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"id", "userId", "operation", "add_duration", "subtract_duration", "multiply_duration", "divide_duration"}).
        AddRow(1, 1, "2+2", 10, 10, 10, 10)
    mock.ExpectQuery("^SELECT (.+) FROM calculations").WillReturnRows(rows)

    // Настройка HTTP-сервера для обработки запросов
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/calculate" {
            var req models.CalculationRequest
            json.NewDecoder(r.Body).Decode(&req)
            if req.Operation != "2+2" {
                t.Errorf("Expected operation '2+2', got '%s'", req.Operation)
            }
            w.WriteHeader(http.StatusOK)
        }
    }))
    defer server.Close()

    // Замена среза серверов на URL мок-сервера
    servers = []string{server.URL}

    // Вызов функции, подлежащей тестированию
    submitCalculations(db)

    // Убедиться, что все ожидания были выполнены
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("There were unfulfilled expectations: %s", err)
    }
}