package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

// TestCafeCount проверяет работу сервера при различных значениях параметра count.
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	// Таблица тестов с разными значениями count и ожидаемым числом кафе в ответе
	requests := []struct {
		count int // значение параметра count в запросе
		want  int // ожидаемое количество кафе в ответе
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, len(cafeList["moscow"])},
	}

	for _, tests := range requests {
		// Формируем URL запроса
		url := "/cafe?city=moscow&count=" + strconv.Itoa(tests.count)
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		// Отправляем запрос
		handler.ServeHTTP(rec, req)

		// Проверка ответа
		require.Equal(t, http.StatusOK, rec.Code)

		// Удаляем пробелы и перенос строки
		body := strings.TrimSpace(rec.Body.String())

		// Обрабатываем случай, когда тело пустое
		if body == "" {
			assert.Equal(t, 0, tests.want)
			continue
		}

		// Разделяем строку по запятой
		cafes := strings.Split(body, ",")

		// Проверка количества элементов
		assert.Len(t, cafes, tests.want)
	}
}

// TestCafeSearch проверяет корректность поиска кафе по подстроке search.
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	// Таблица тестов с разными строками поиска и ожидаемым числом совпадений
	requests := []struct {
		search    string // подстрока, которую ищем
		wantCount int    // ожидаемое количество кафе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tests := range requests {
		// Формируем URL запроса
		url := "/cafe?city=moscow&search=" + tests.search
		req := httptest.NewRequest("GET", url, nil)
		rec := httptest.NewRecorder()

		// Отправляем запрос
		handler.ServeHTTP(rec, req)

		// Проверка кода ответа
		require.Equal(t, http.StatusOK, rec.Code)

		// Чистим тело ответа от лишних символов
		body := strings.TrimSpace(rec.Body.String())

		// Обработка случая с пустым ответом
		if body == "" {
			assert.Equal(t, 0, tests.wantCount)
			continue
		}

		// Разделяем строку на названия кафе
		cafes := strings.Split(body, ",")

		// Проверка количества найденных кафе
		assert.Len(t, cafes, tests.wantCount)

		// Проверка, что каждое кафе содержит строку поиска
		searchLower := strings.ToLower(tests.search)
		for _, cafe := range cafes {
			assert.Contains(t, strings.ToLower(cafe), searchLower)
		}
	}
}
