package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// currencyByLanguage is a first guess only; the user can change it later.
var currencyByLanguage = map[string]string{
	"ru": "RUB", "uk": "UAH", "be": "BYN", "kk": "KZT", "uz": "UZS", "az": "AZN",
	"hy": "AMD", "ka": "GEL", "tr": "TRY", "pl": "PLN", "cs": "CZK", "sv": "SEK",
	"nb": "NOK", "no": "NOK", "da": "DKK", "en": "USD", "de": "EUR", "fr": "EUR",
	"es": "EUR", "it": "EUR", "nl": "EUR", "pt": "EUR", "fi": "EUR", "el": "EUR",
	"sr": "RSD", "ja": "JPY", "zh": "CNY", "ko": "KRW", "id": "IDR", "th": "THB",
	"hi": "INR", "he": "ILS", "ar": "AED",
}

func guessCurrency(languageCode string) string {
	lang, _, _ := strings.Cut(strings.ToLower(languageCode), "-")
	if c, ok := currencyByLanguage[lang]; ok {
		return c
	}
	return "USD"
}

type defaultCategory struct {
	name     string
	icon     string
	children []defaultCategory
}

var defaultExpenseCategories = []defaultCategory{
	{name: "Food", icon: "🍔", children: []defaultCategory{
		{name: "Groceries", icon: "🛒"}, {name: "Restaurants", icon: "🍽"}, {name: "Delivery", icon: "🛵"},
	}},
	{name: "Transport", icon: "🚌", children: []defaultCategory{
		{name: "Taxi", icon: "🚕"}, {name: "Fuel", icon: "⛽"},
	}},
	{name: "Home", icon: "🏠", children: []defaultCategory{
		{name: "Rent", icon: "🔑"}, {name: "Utilities", icon: "💡"},
	}},
	{name: "Shopping", icon: "🛍"},
	{name: "Health", icon: "💊"},
	{name: "Entertainment", icon: "🎬"},
	{name: "Travel", icon: "✈️"},
	{name: "Gifts", icon: "🎁"},
	{name: "Other", icon: "💸"},
}

var defaultIncomeCategories = []defaultCategory{
	{name: "Salary", icon: "💰"},
	{name: "Bonus", icon: "🧾"},
	{name: "Gift", icon: "🎁"},
	{name: "Other", icon: "💸"},
}

// loadMe creates the user, their personal context, a starting account and the default
// categories on first call, and returns the same picture on every call after that.
func (s *server) loadMe(ctx context.Context, tu tgUser) (Me, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Me{}, err
	}
	defer tx.Rollback(ctx)

	var me Me
	err = tx.QueryRow(ctx, `
		insert into users (id, first_name, username, language_code, base_currency)
		values ($1, $2, nullif($3, ''), nullif($4, ''), $5)
		on conflict (id) do update
			set first_name = excluded.first_name,
			    username = excluded.username,
			    language_code = excluded.language_code
		returning id, first_name, username, language_code, base_currency`,
		tu.ID, tu.FirstName, tu.Username, tu.LanguageCode, guessCurrency(tu.LanguageCode),
	).Scan(&me.User.Id, &me.User.FirstName, &me.User.Username, &me.User.LanguageCode, &me.User.BaseCurrency)
	if err != nil {
		return Me{}, err
	}

	created := true
	err = tx.QueryRow(ctx, `
		insert into contexts (kind, name, owner_id, base_currency)
		values ('personal', $1, $2, $3)
		on conflict (owner_id) where kind = 'personal' do nothing
		returning id, kind, name, owner_id, base_currency`,
		tu.FirstName, tu.ID, me.User.BaseCurrency,
	).Scan(&me.Context.Id, &me.Context.Kind, &me.Context.Name, &me.Context.OwnerId, &me.Context.BaseCurrency)
	if errors.Is(err, pgx.ErrNoRows) {
		created = false
		err = tx.QueryRow(ctx, `
			select id, kind, name, owner_id, base_currency
			from contexts where owner_id = $1 and kind = 'personal'`,
			tu.ID,
		).Scan(&me.Context.Id, &me.Context.Kind, &me.Context.Name, &me.Context.OwnerId, &me.Context.BaseCurrency)
	}
	if err != nil {
		return Me{}, err
	}

	if created {
		if err := seedContext(ctx, tx, me.Context.Id, tu.ID, me.Context.BaseCurrency); err != nil {
			return Me{}, err
		}
	}

	if me.Accounts, err = listAccounts(ctx, tx, me.Context.Id); err != nil {
		return Me{}, err
	}
	if me.Categories, err = listCategories(ctx, tx, me.Context.Id); err != nil {
		return Me{}, err
	}
	return me, tx.Commit(ctx)
}

func seedContext(ctx context.Context, tx pgx.Tx, contextID openapi_types.UUID, userID int64, currency string) error {
	_, err := tx.Exec(ctx, `
		insert into context_members (context_id, user_id, status) values ($1, $2, 'active')`,
		contextID, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		insert into accounts (context_id, owner_id, name, currency) values ($1, $2, 'Cash', $3)`,
		contextID, userID, currency)
	if err != nil {
		return err
	}

	for kind, roots := range map[string][]defaultCategory{
		"expense": defaultExpenseCategories,
		"income":  defaultIncomeCategories,
	} {
		for _, root := range roots {
			var parentID openapi_types.UUID
			err := tx.QueryRow(ctx, `
				insert into categories (context_id, parent_id, kind, name, icon)
				values ($1, null, $2, $3, $4) returning id`,
				contextID, kind, root.name, root.icon,
			).Scan(&parentID)
			if err != nil {
				return err
			}
			for _, child := range root.children {
				_, err := tx.Exec(ctx, `
					insert into categories (context_id, parent_id, kind, name, icon)
					values ($1, $2, $3, $4, $5)`,
					contextID, parentID, kind, child.name, child.icon)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func listAccounts(ctx context.Context, q pgx.Tx, contextID openapi_types.UUID) ([]Account, error) {
	rows, err := q.Query(ctx, `
		select id, name, currency, initial_balance::text
		from accounts where context_id = $1 and not is_archived
		order by created_at`, contextID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.Id, &a.Name, &a.Currency, &a.InitialBalance); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func listCategories(ctx context.Context, q pgx.Tx, contextID openapi_types.UUID) ([]Category, error) {
	rows, err := q.Query(ctx, `
		select id, parent_id, kind, name, icon
		from categories where context_id = $1 and not is_archived
		order by parent_id nulls first, name`, contextID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Id, &c.ParentId, &c.Kind, &c.Name, &c.Icon); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *server) personalContext(ctx context.Context, userID int64) (openapi_types.UUID, int64, error) {
	var id openapi_types.UUID
	var ownerID int64
	err := s.db.QueryRow(ctx, `
		select id, owner_id from contexts where owner_id = $1 and kind = 'personal'`,
		userID).Scan(&id, &ownerID)
	return id, ownerID, err
}

const transactionColumns = `id, author_id, type, account_id, to_account_id, category_id,
	amount::text, currency, rate::text, occurred_on, note, created_at`

func scanTransaction(row pgx.Row) (Transaction, error) {
	var t Transaction
	var occurredOn time.Time
	var note string
	err := row.Scan(&t.Id, &t.AuthorId, &t.Type, &t.AccountId, &t.ToAccountId, &t.CategoryId,
		&t.Amount, &t.Currency, &t.Rate, &occurredOn, &note, &t.CreatedAt)
	t.OccurredOn = openapi_types.Date{Time: occurredOn}
	t.Note = &note
	return t, err
}

func (s *server) listTransactions(ctx context.Context, contextID openapi_types.UUID, from, to time.Time) ([]Transaction, error) {
	rows, err := s.db.Query(ctx, `select `+transactionColumns+`
		from transactions
		where context_id = $1 and occurred_on between $2 and $3
		order by occurred_on desc, created_at desc`, contextID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

// accountAndCategory reports the account's currency and the category's kind, checking in
// the same query that both belong to this context, are not archived, and that the account
// belongs to this user.
func (s *server) accountAndCategory(ctx context.Context, contextID openapi_types.UUID, userID int64, accountID openapi_types.UUID, categoryID *openapi_types.UUID) (currency string, kind *string, err error) {
	err = s.db.QueryRow(ctx, `
		select a.currency,
		       (select c.kind from categories c
		        where c.id = $4 and c.context_id = $1 and not c.is_archived)
		from accounts a
		where a.id = $3 and a.context_id = $1 and a.owner_id = $2 and not a.is_archived`,
		contextID, userID, accountID, categoryID).Scan(&currency, &kind)
	return currency, kind, err
}

func (s *server) createTransaction(ctx context.Context, contextID openapi_types.UUID, userID int64, in NewTransaction, currency string) (Transaction, error) {
	note := ""
	if in.Note != nil {
		note = *in.Note
	}
	return scanTransaction(s.db.QueryRow(ctx, `
		insert into transactions
			(context_id, author_id, type, account_id, category_id, amount, currency, rate, occurred_on, note)
		values ($1, $2, $3, $4, $5, cast($6 as numeric), $7, 1, $8, $9)
		returning `+transactionColumns,
		contextID, userID, in.Type, in.AccountId, in.CategoryId, in.Amount, currency,
		in.OccurredOn.Time, note))
}
