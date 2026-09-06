package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const maxNoteLen = 500

var amountPattern = regexp.MustCompile(`^\d{1,16}(\.\d{1,2})?$`)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func writeErr(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Error{Code: code, Message: message})
}

func serverErr(w http.ResponseWriter, err error) {
	slog.Error("request failed", "err", err)
	writeErr(w, http.StatusInternalServerError, "internal", "something went wrong")
}

func (s *server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	me, err := s.loadMe(r.Context(), userFrom(r.Context()))
	if err != nil {
		serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, me)
}

func (s *server) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	from, err := parseDate(r.URL.Query().Get("from"), time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_from", "from must be a YYYY-MM-DD date")
		return
	}
	to, err := parseDate(r.URL.Query().Get("to"), today)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_to", "to must be a YYYY-MM-DD date")
		return
	}
	if to.Before(from) {
		writeErr(w, http.StatusBadRequest, "invalid_range", "to is before from")
		return
	}

	contextID, _, err := s.personalContext(r.Context(), userFrom(r.Context()).ID)
	if err != nil {
		serverErr(w, err)
		return
	}

	transactions, err := s.listTransactions(r.Context(), contextID, from, to)
	if err != nil {
		serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transactions)
}

func parseDate(raw string, fallback time.Time) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	return time.Parse(time.DateOnly, raw)
}

func (s *server) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var in NewTransaction
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_body", "body is not valid JSON")
		return
	}
	if code, message := validateNewTransaction(in); code != "" {
		writeErr(w, http.StatusBadRequest, code, message)
		return
	}

	user := userFrom(r.Context())
	contextID, _, err := s.personalContext(r.Context(), user.ID)
	if err != nil {
		serverErr(w, err)
		return
	}

	currency, kind, err := s.accountAndCategory(r.Context(), contextID, user.ID, in.AccountId, in.CategoryId)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusBadRequest, "unknown_account", "no such account in this context")
		return
	}
	if err != nil {
		serverErr(w, err)
		return
	}
	if kind == nil {
		writeErr(w, http.StatusBadRequest, "unknown_category", "no such category in this context")
		return
	}
	if *kind != string(in.Type) {
		writeErr(w, http.StatusBadRequest, "category_kind_mismatch", "the category does not match the transaction type")
		return
	}

	created, err := s.createTransaction(r.Context(), contextID, user.ID, in, currency)
	if err != nil {
		serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func validateNewTransaction(in NewTransaction) (code, message string) {
	switch in.Type {
	case TransactionTypeExpense, TransactionTypeIncome:
	case TransactionTypeTransfer:
		return "unsupported_type", "transfers are not implemented yet"
	default:
		return "invalid_type", "type must be expense or income"
	}
	if in.CategoryId == nil {
		return "category_required", "expense and income need a category"
	}
	if in.AccountId == uuid.Nil {
		return "account_required", "account_id is required"
	}
	if !amountPattern.MatchString(in.Amount) || strings.Trim(in.Amount, "0.") == "" {
		return "invalid_amount", "amount must be a positive number with at most two decimals"
	}
	if in.OccurredOn.Year() < 2000 || in.OccurredOn.After(time.Now().AddDate(1, 0, 0)) {
		return "invalid_date", "occurred_on is out of range"
	}
	if in.Note != nil && len([]rune(*in.Note)) > maxNoteLen {
		return "note_too_long", "note is longer than 500 characters"
	}
	return "", ""
}

func (s *server) handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "no such transaction")
		return
	}

	user := userFrom(r.Context())
	contextID, ownerID, err := s.personalContext(r.Context(), user.ID)
	if err != nil {
		serverErr(w, err)
		return
	}

	var authorID int64
	err = s.db.QueryRow(r.Context(),
		`select author_id from transactions where id = $1 and context_id = $2`,
		openapi_types.UUID(id), contextID).Scan(&authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not_found", "no such transaction")
		return
	}
	if err != nil {
		serverErr(w, err)
		return
	}
	if authorID != user.ID && ownerID != user.ID {
		writeErr(w, http.StatusForbidden, "forbidden", "only the author or the context owner may delete this")
		return
	}

	if _, err := s.db.Exec(r.Context(), `delete from transactions where id = $1`, openapi_types.UUID(id)); err != nil {
		serverErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
