package migration

const (
	event = `
create table if not exists event
(
    name        text constraint event_pk primary key,
    created_at  timestamp default now(),
    finished_at timestamp not null,
    author      numeric   not null,
    result      numeric   not null,
    status      text      default 'open',
    winners     numeric[] not null
)`

	eventParticipant = `
create table if not exists event_participant
(
    event_name text,
    user_id    numeric not null,
    bet        numeric not null,
    updated_at timestamp default now(),
    constraint event_participant_pk primary key (event_name, user_id)
)`
)
