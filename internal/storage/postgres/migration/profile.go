package migration

const profile = `
create table if not exists profile
(
    id         numeric default 0  not null constraint profile_pk primary key,
    first_name text    default '' not null,
    last_name  text    default '' not null,
    user_name  text    default '' not null
)`
