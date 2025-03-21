create table users (
	id bigserial primary key not null,
	username varchar(100) not null unique,
	create_time timestamp default current_timestamp not null,
	update_time timestamp default current_timestamp not null
);
create index username on users (username);

create table friendship (
	id bigserial primary key not null,
	small_id bigint not null,
	big_id bigint not null,
	create_time timestamp default current_timestamp not null,
	FOREIGN KEY (small_id) REFERENCES users(id) on delete cascade,
	FOREIGN KEY (big_id) REFERENCES users(id) on delete cascade,
	CONSTRAINT unique_friendship UNIQUE (small_id,big_id)
);
create index users_to_users on friendship(small_id, big_id);