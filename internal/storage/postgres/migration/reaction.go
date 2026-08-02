package migration

// message_author maps a (chat, message) to its author so a later reaction on it can be
// credited to the right person. Kept indefinitely — no retention/cleanup.
const messageAuthor = `
create table if not exists message_author
(
    chat_id      numeric not null,
    message_id   numeric not null,
    user_id      numeric not null,
    message_date date    not null,
    primary key (chat_id, message_id)
)`
