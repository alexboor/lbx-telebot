package migration

const count = `
create table if not exists counting
(
    user_id numeric,
    chat_id numeric,
    date    date,
    word    numeric default 0,
    reply   numeric default 0,
    forward numeric default 0,
    media   numeric default 0,
    sticker numeric default 0,
    unique (user_id, chat_id, date)
)`

const countMessages = `
alter table counting
    add if not exists message numeric default 0`
