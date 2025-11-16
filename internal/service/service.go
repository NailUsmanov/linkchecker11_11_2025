package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
	"github.com/jung-kurt/gofpdf"
)

var client = &http.Client{Timeout: 5 * time.Second}

// CheckLink - возвращает статус ссылки или ошибку.
func CheckLink(ctx context.Context, link string) (bool, error) {
	// Принимаю и нормализую URL
	rawURL := strings.TrimSpace(link)
	if rawURL == "" {
		return false, errors.New("empty url")
	}
	// Если нету ://, значит схему не указывали, добавляю.
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	// Достаю сам УРЛ
	u, err := url.Parse(rawURL)
	if err != nil {
		return false, fmt.Errorf("invalid url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false, fmt.Errorf("unsupported scheme: %s", err)
	}

	if u.Host == "" {
		return false, errors.New("missing host in URL")
	}

	// Формирую запрос через вызов HEAD
	req, err := http.NewRequest(http.MethodHead, u.String(), nil)
	if err != nil {
		return false, err
	}

	// Выполняю HTTP-запрос
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Если метод HEAD не поддерживается, используем GET
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
		// само тело ответа не требуется, отбрасываем его
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		req, err = http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return false, err
		}

		resp, err = client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
	}

	// Ссылка доступна
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	}

	// Если 400-ые и 500-ые коды, значит сайт недоступен
	return false, nil
}

// CreatePDF - создает PDF файл с обработанными ссылками.
func CreatePDF(data map[int]models.ResponseSentLinks) ([]byte, error) {
	// Cоздаю pdf
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Links report", false)
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// для удобства собираем ключи мапы в слайс, далее их сортировка
	reqKeys := make([]int, 0, len(data))
	for k := range data {
		reqKeys = append(reqKeys, k)
	}
	slices.Sort(reqKeys)

	// проходимся циклом по data и обрабатываем данные для формирования PDF
	for _, key := range reqKeys {
		resp := data[key]

		// заголовок для конкретного блока
		header := fmt.Sprintf("Request #%d", resp.Num)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, header)
		pdf.Ln(10)

		// сортировка ссылок для вывода по алфавиту
		linkKeys := make([]string, 0, len(resp.Links))
		for link := range resp.Links {
			linkKeys = append(linkKeys, link)
		}
		slices.Sort(linkKeys)

		// добавление самих ссылок со статусами.
		pdf.SetFont("Arial", "", 11)
		for _, link := range linkKeys {
			status := resp.Links[link]

			line := fmt.Sprintf("%s - %s", link, status)
			pdf.Cell(0, 6, line)
			pdf.Ln(6)
		}

		// Пустая строка между блоками запросов
		pdf.Ln(4)
	}

	// Буфер памяти, куда кладется готовый PDF
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	// Возврат байты PDF.
	return buf.Bytes(), nil
}
