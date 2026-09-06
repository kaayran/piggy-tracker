package main

import (
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testToken = "123456:test-bot-token"

func signed(t *testing.T, fields map[string]string) string {
	t.Helper()
	pairs := make([]string, 0, len(fields))
	for k, v := range fields {
		if k != "signature" {
			pairs = append(pairs, k+"="+v)
		}
	}
	sort.Strings(pairs)
	sum := hmacSHA256(hmacSHA256([]byte("WebAppData"), testToken), strings.Join(pairs, "\n"))

	q := url.Values{}
	for k, v := range fields {
		q.Set(k, v)
	}
	q.Set("hash", hex.EncodeToString(sum))
	return q.Encode()
}

func validFields(now time.Time) map[string]string {
	return map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"query_id":  "AAH123",
		"user":      `{"id":42,"first_name":"Ann","username":"ann","language_code":"ru"}`,
	}
}

func TestVerifyInitData(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)

	tampered := signed(t, validFields(now))
	tampered = strings.Replace(tampered, "Ann", "Bob", 1)

	stale := validFields(now.Add(-25 * time.Hour))

	future := validFields(now.Add(time.Hour))

	noUser := validFields(now)
	delete(noUser, "user")

	withSignature := validFields(now)
	withSignature["signature"] = "ed25519-blob"

	tests := []struct {
		name string
		raw  string
		ok   bool
	}{
		{"valid", signed(t, validFields(now)), true},
		{"signature is not part of the hash", signed(t, withSignature), true},
		{"missing header", "", false},
		{"not a query string", "%zz", false},
		{"tampered field", tampered, false},
		{"wrong hash", strings.Replace(signed(t, validFields(now)), "hash=", "hash=00", 1), false},
		{"no hash", url.Values{"auth_date": {"1800000000"}}.Encode(), false},
		{"stale auth_date", signed(t, stale), false},
		{"auth_date in the future", signed(t, future), false},
		{"missing user", signed(t, noUser), false},
		{"duplicate key", signed(t, validFields(now)) + "&user=%7B%22id%22%3A1%7D", false},
		{"signed by another token", signedWith(t, "999:other", validFields(now)), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := verifyInitData(tc.raw, testToken, now)
			if tc.ok {
				if err != nil {
					t.Fatalf("want ok, got %v", err)
				}
				if u.ID != 42 || u.FirstName != "Ann" || u.LanguageCode != "ru" {
					t.Fatalf("unexpected user %+v", u)
				}
				return
			}
			if err == nil {
				t.Fatalf("want rejection, got user %+v", u)
			}
		})
	}
}

func signedWith(t *testing.T, token string, fields map[string]string) string {
	t.Helper()
	pairs := make([]string, 0, len(fields))
	for k, v := range fields {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	sum := hmacSHA256(hmacSHA256([]byte("WebAppData"), token), strings.Join(pairs, "\n"))

	q := url.Values{}
	for k, v := range fields {
		q.Set(k, v)
	}
	q.Set("hash", hex.EncodeToString(sum))
	return q.Encode()
}
