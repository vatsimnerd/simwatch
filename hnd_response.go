package simwatch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type (
	HTTPError struct {
		Error string `json:"error"`
	}

	PaginatedResponse[T any] struct {
		Count      int `json:"count"`
		TotalPages int `json:"total_pages"`
		Page       int `json:"page"`
		Data       []T `json:"data"`
	}
)

const (
	defaultPage  = 1
	defaultLimit = 20
)

func sendError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	he := HTTPError{Error: message}
	sendJSON(w, he)
}

func paginated[T any](data []T, page, limit int) *PaginatedResponse[T] {
	count := len(data)
	totalPages := count / limit

	if page > totalPages {
		page = totalPages
	} else if page < 1 {
		page = 1
	}

	start := (page - 1) * limit
	end := page * limit

	if end > len(data) {
		end = len(data)
	}

	return &PaginatedResponse[T]{
		Count:      count,
		Page:       page,
		TotalPages: totalPages,
		Data:       data[start:end],
	}
}

func getPagination(r *http.Request) (page, limit int) {
	values := r.URL.Query()

	pStr := values.Get("page")
	page = defaultPage
	if pStr != "" {
		p, err := strconv.ParseInt(pStr, 10, 64)
		if err == nil {
			page = int(p)
		}
	}

	lStr := values.Get("limit")
	limit = defaultLimit
	if lStr != "" {
		l, err := strconv.ParseInt(lStr, 10, 64)
		if err == nil {
			limit = int(l)
		}
	}

	return
}

func sendPaginated[T any](w http.ResponseWriter, r *http.Request, data []T) {
	page, limit := getPagination(r)
	pData := paginated(data, page, limit)
	sendJSON(w, pData)
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	l := log.WithField("func", "sendJSON")

	raw, err := json.Marshal(data)
	if err != nil {
		l.WithError(err).Error("error marshaling paginated data")
		sendError(w, 500, fmt.Sprintf("error marshaling paginated data: %v", err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(raw)
	if err != nil {
		log.WithError(err).Error("error writing paginated data")
	}
}
