package migration

// TODO: add primary key for user_id
const wordCount = `
create table if not exists word_count
(
    user_id bigint,
    chat_id bigint,
    date    date,
    val     int,
    unique  (user_id, chat_id, date)
)`
