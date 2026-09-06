-- +goose Up
create table users (
    id            bigint primary key,
    first_name    text not null,
    username      text,
    language_code text,
    base_currency char(3) not null check (base_currency ~ '^[A-Z]{3}$'),
    created_at    timestamptz not null default now()
);

create table contexts (
    id                uuid primary key default gen_random_uuid(),
    kind              text not null check (kind in ('personal', 'group')),
    name              text not null check (length(name) between 1 and 100),
    owner_id          bigint not null references users (id),
    base_currency     char(3) not null check (base_currency ~ '^[A-Z]{3}$'),
    invite_token      text unique,
    invite_expires_at timestamptz,
    created_at        timestamptz not null default now(),
    check ((invite_token is null) = (invite_expires_at is null))
);

create unique index contexts_one_personal_per_owner
    on contexts (owner_id) where kind = 'personal';

create table context_members (
    context_id uuid not null references contexts (id) on delete cascade,
    user_id    bigint not null references users (id),
    status     text not null check (status in ('active', 'left')),
    joined_at  timestamptz not null default now(),
    primary key (context_id, user_id)
);

create table accounts (
    id              uuid primary key default gen_random_uuid(),
    context_id      uuid not null references contexts (id) on delete cascade,
    owner_id        bigint not null references users (id),
    name            text not null check (length(name) between 1 and 60),
    currency        char(3) not null check (currency ~ '^[A-Z]{3}$'),
    initial_balance numeric(18, 2) not null default 0,
    is_archived     boolean not null default false,
    created_at      timestamptz not null default now()
);

create index accounts_active on accounts (context_id) where not is_archived;

-- ponytail: the "a subcategory cannot have children" rule is not enforced here; the only
-- writer today is the default seed. Add a check when a category-create endpoint appears.
create table categories (
    id          uuid primary key default gen_random_uuid(),
    context_id  uuid not null references contexts (id) on delete cascade,
    parent_id   uuid references categories (id),
    kind        text not null check (kind in ('expense', 'income')),
    name        text not null check (length(name) between 1 and 60),
    icon        text not null,
    is_archived boolean not null default false
);

create index categories_active on categories (context_id) where not is_archived;

create table transactions (
    id            uuid primary key default gen_random_uuid(),
    context_id    uuid not null references contexts (id) on delete cascade,
    author_id     bigint not null references users (id),
    type          text not null check (type in ('expense', 'income', 'transfer')),
    account_id    uuid not null references accounts (id),
    to_account_id uuid references accounts (id),
    category_id   uuid references categories (id),
    amount        numeric(18, 2) not null check (amount > 0),
    currency      char(3) not null check (currency ~ '^[A-Z]{3}$'),
    rate          numeric(18, 8) not null check (rate > 0),
    occurred_on   date not null,
    note          text not null default '' check (length(note) <= 500),
    created_at    timestamptz not null default now(),
    check ((type = 'transfer') = (to_account_id is not null)),
    check ((type = 'transfer') = (category_id is null))
);

create index transactions_context_date on transactions (context_id, occurred_on desc, created_at desc);

-- +goose Down
drop table transactions;
drop table categories;
drop table accounts;
drop table context_members;
drop table contexts;
drop table users;
