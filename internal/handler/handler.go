package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/NailUsmanov/linkchecker11_11_2025/internal/models"
	"github.com/NailUsmanov/linkchecker11_11_2025/internal/service"
	"go.uber.org/zap"
)

// reqNum - нужен для инкрементного увеличения счетчика запросов.
var reqNum atomic.Uint64

// NextReqID - увеличивает счетчик запросов инкрементно.
func NextReqNum() uint64 {
	return reqNum.Add(1)
}

// NewCreateLinks - проверяет и сохраняет переданные в запросе ссылки.
func NewCreateLinks(s Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверим метод
		if r.Method != http.MethodPost {
			http.Error(w, "only POST method avaible", http.StatusMethodNotAllowed)
			return
		}

		// проверим, что запрос отправлен в виде JSON формата
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		// берем данные из запроса
		defer r.Body.Close()

		var req models.RequestSentLinks
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sugar.Errorf("cannot decode requset JSON body: %v", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		if len(req.Links) == 0 {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}

		// Получаем номер запроса
		numReq := NextReqNum()

		// Возвращаем ответ
		resp := models.ResponseSentLinks{
			Num:   int(numReq),
			Links: make(map[string]string, len(req.Links)),
		}

		for _, v := range req.Links {
			stat, err := service.CheckLink(r.Context(), v)
			if err != nil {
				sugar.Errorf("checklink failed: %v", err)
				http.Error(w, "checklink failed", http.StatusInternalServerError)
				return
			}

			statusStr := "not available"
			if stat {
				statusStr = "available"
			}
			resp.Links[v] = statusStr
		}

		// Сохраняем ссылки
		err := s.Save(r.Context(), resp)
		if err != nil {
			sugar.Errorf("save links failed: %v", err)
			http.Error(w, "save links failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			sugar.Errorf("error encoding response: %v", err)
		}
	}
}

// NewGetLinks - выдает пользователю PDF файл по конкретному номеру запроса с уже проверенными ссылками.
func NewGetLinks(s Storage, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяю метод
		if r.Method != http.MethodGet {
			http.Error(w, "only GET method avaible", http.StatusMethodNotAllowed)
			return
		}

		// проверим, что запрос отправлен в виде JSON формата
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		// Парсим JSON из тела в RequestLinksNum
		defer r.Body.Close()

		reqNums := models.RequestLinksNum{}
		if err := json.NewDecoder(r.Body).Decode(&reqNums); err != nil {
			sugar.Errorf("cannot decode request JSON format: %v", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		if len(reqNums.LinksList) == 0 {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}

		// Достаем ссылки из базы данных
		data, err := s.Get(r.Context(), reqNums.LinksList)
		if err != nil {
			sugar.Errorf("get links failed: %v", err)
			http.Error(w, "get links failed", http.StatusInternalServerError)
			return
		}

		// Проверю, что все номера нашлись
		if len(data) != len(reqNums.LinksList) {
			missing := make([]int, 0)
			for _, n := range reqNums.LinksList {
				if _, ok := data[n]; !ok {
					missing = append(missing, n)
				}
			}

			sugar.Warnw("some request numbers not found", "missing", missing)
			http.Error(w, "some request numbers not found", http.StatusNotFound)
			return
		}

		// Собираем PDF
		buf, err := service.CreatePDF(data)
		if err != nil {
			sugar.Errorf("create pdf failed: %v", err)
			http.Error(w, "create pdf failed", http.StatusInternalServerError)
			return
		}

		// Возвращаем ответ
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf)
	}
}
