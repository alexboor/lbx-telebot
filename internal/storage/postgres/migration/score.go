package migration

const score = `
create table if not exists score (
    user_id serial primary key,
    score int not null,
	last_update timestamp not null
)
`
