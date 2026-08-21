-- Dagens kurskod: the answer for each day, fixed the first time it is asked for.
--
-- Before this table the answer was derived from the date on every request. That
-- made it global only while every server saw an identical candidate pool, and
-- the pool comes from live exam counts — so a scrape that pushed a course over
-- the threshold mid-day silently changed that day's answer for everyone who
-- loaded after it. Writing the answer down once removes the whole problem.

create table if not exists public.daily_puzzle (
  puzzle_date  date not null,
  university   text not null,
  course_code  text not null,
  created_at   timestamptz not null default now(),
  -- Composite, not just the date: each university runs its own puzzle, and a
  -- date-only key would let the first one asked for lock out the rest.
  primary key (puzzle_date, university)
);

create index if not exists daily_puzzle_university_date_idx
  on public.daily_puzzle (university, puzzle_date desc);

-- RLS on with no policies at all: PostgREST callers using the anon or an
-- authenticated key get nothing. Only the service role — which bypasses RLS,
-- and which only the Go service holds — can read or write.
--
-- This is the point of the table's security: the anon key is public and shipped
-- to every browser, so anything it can SELECT is effectively published. Today's
-- answer must never be in that set.
alter table public.daily_puzzle enable row level security;
