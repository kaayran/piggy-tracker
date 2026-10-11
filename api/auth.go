package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	loginMaxAge   = 24 * time.Hour
	sessionCookie = "session"
	sessionTTL    = 30 * 24 * time.Hour
)

var errUnauthorized = errors.New("invalid login payload")

type tgUser struct {
	ID        int64
	FirstName string
	Username  string
}

type ctxKey struct{}

func userID(ctx context.Context) int64 {
	return ctx.Value(ctxKey{}).(int64)
}

// verifyLogin checks the HMAC that Telegram puts on a Login Widget payload and returns the
// user it carries. Everything about a request's identity comes from here and nowhere else.
func verifyLogin(values url.Values, botToken string, now time.Time) (tgUser, error) {
	pairs := make([]string, 0, len(values))
	var gotHash string
	for key, vs := range values {
		if len(vs) != 1 {
			return tgUser{}, errUnauthorized
		}
		if key == "hash" {
			gotHash = vs[0]
		} else {
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
	if age > loginMaxAge || age < -time.Minute {
		return tgUser{}, errUnauthorized
	}

	secret := sha256.Sum256([]byte(botToken))
	want := hmacSHA256(secret[:], strings.Join(pairs, "\n"))
	got, err := hex.DecodeString(gotHash)
	if err != nil || !hmac.Equal(want, got) {
		return tgUser{}, errUnauthorized
	}

	id, err := strconv.ParseInt(values.Get("id"), 10, 64)
	if err != nil || id <= 0 || values.Get("first_name") == "" {
		return tgUser{}, errUnauthorized
	}
	return tgUser{ID: id, FirstName: values.Get("first_name"), Username: values.Get("username")}, nil
}

func hmacSHA256(key []byte, msg string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return m.Sum(nil)
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func (s *server) handleSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		writeErr(w, http.StatusForbidden, "cross_site", "sign-in must come from this site")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	if err := r.ParseForm(); err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized", "invalid login payload")
		return
	}
	u, err := verifyLogin(r.PostForm, s.botToken, time.Now())
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized", "invalid login payload")
		return
	}

	token, err := s.signIn(r.Context(), u, r.Header.Get("Accept-Language"))
	if err != nil {
		serverErr(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) signIn(ctx context.Context, u tgUser, acceptLanguage string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		insert into users (id, first_name, username, base_currency)
		values ($1, $2, nullif($3, ''), $4)
		on conflict (id) do update
			set first_name = excluded.first_name,
			    username = excluded.username`,
		u.ID, u.FirstName, u.Username, guessCurrency(primaryLanguage(acceptLanguage)),
	); err != nil {
		return "", err
	}
	if err := ensurePersonalContext(ctx, tx, u.ID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `delete from sessions where expires_at < now()`); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		insert into sessions (token_hash, user_id, expires_at) values ($1, $2, $3)`,
		hashToken(token), u.ID, time.Now().Add(sessionTTL),
	); err != nil {
		return "", err
	}
	return token, tx.Commit(ctx)
}

func (s *server) handleSignOut(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		if _, err := s.db.Exec(r.Context(), `delete from sessions where token_hash = $1`, hashToken(c.Value)); err != nil {
			serverErr(w, err)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "unauthorized", "not signed in")
			return
		}
		var id int64
		err = s.db.QueryRow(r.Context(), `
			select user_id from sessions where token_hash = $1 and expires_at > now()`,
			hashToken(c.Value)).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusUnauthorized, "unauthorized", "not signed in")
			return
		}
		if err != nil {
			serverErr(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func primaryLanguage(acceptLanguage string) string {
	lang, _, _ := strings.Cut(acceptLanguage, ",")
	lang, _, _ = strings.Cut(lang, ";")
	return strings.TrimSpace(lang)
}
