package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const initDataHeader = "X-Telegram-Init-Data"

const initDataMaxAge = 24 * time.Hour

var errUnauthorized = errors.New("invalid init data")

type tgUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type ctxKey struct{}

func userFrom(ctx context.Context) tgUser {
	return ctx.Value(ctxKey{}).(tgUser)
}

// verifyInitData checks the HMAC that Telegram puts on initData and returns the user it
// carries. Everything about a request's identity comes from here and nowhere else.
func verifyInitData(raw, botToken string, now time.Time) (tgUser, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return tgUser{}, errUnauthorized
	}

	pairs := make([]string, 0, len(values))
	var gotHash string
	for key, vs := range values {
		if len(vs) != 1 {
			return tgUser{}, errUnauthorized
		}
		switch key {
		case "hash":
			gotHash = vs[0]
		case "signature":
			// Ed25519 over the same string; it cannot be part of what it signs.
		default:
			pairs = append(pairs, key+"="+vs[0])
		}
	}
	if gotHash == "" {
		return tgUser{}, errUnauthorized
	}
	sort.Strings(pairs)

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return tgUser{}, errUnauthorized
	}
	age := now.Sub(time.Unix(authDate, 0))
	if age > initDataMaxAge || age < -time.Minute {
		return tgUser{}, errUnauthorized
	}

	secret := hmacSHA256([]byte("WebAppData"), botToken)
	want := hmacSHA256(secret, strings.Join(pairs, "\n"))
	got, err := hex.DecodeString(gotHash)
	if err != nil || !hmac.Equal(want, got) {
		return tgUser{}, errUnauthorized
	}

	var u tgUser
	if err := json.Unmarshal([]byte(values.Get("user")), &u); err != nil {
		return tgUser{}, errUnauthorized
	}
	if u.ID <= 0 || u.FirstName == "" {
		return tgUser{}, errUnauthorized
	}
	return u, nil
}

func hmacSHA256(key []byte, msg string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return m.Sum(nil)
}

func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := verifyInitData(r.Header.Get(initDataHeader), s.botToken, time.Now())
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "unauthorized", "invalid or missing init data")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
	})
}
