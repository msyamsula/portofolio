create table users (
	id bigserial primary key not null,
	username varchar(100) not null unique,
	create_time timestamp default current_timestamp not null,
	update_time timestamp default current_timestamp not null
);
create index username on users (username);