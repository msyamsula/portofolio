create table users (
	id bigserial primary key not null,
	username varchar(100) not null unique,
	online bool not null default false
	create_time timestamp default current_timestamp not null,
	update_time timestamp default current_timestamp not null
);
create index username on users (username);


