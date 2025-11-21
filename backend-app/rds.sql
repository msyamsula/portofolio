create table friendship (
	id bigserial primary key not null,
	small_id bigint not null,
	big_id bigint not null,
	create_time timestamp default current_timestamp not null,
	CONSTRAINT unique_friendship UNIQUE (small_id,big_id)
);
create index users_to_users on friendship(small_id, big_id);
create index small on friendship(small_id);
create index big on friendship(big_id);

---
create table url (
	id bigserial primary key not null,
	short_url varchar(50) not null unique,
	long_url varchar(5000) not null,
	create_time date default current_timestamp not null
);

create index short_url_index on url (short_url);
create index long_url_index on url (long_url);

--
create table users (
	id bigserial primary key not null,
	username varchar(100) not null unique,
	online bool not null default false,
	create_time timestamp default current_timestamp not null,
	update_time timestamp default current_timestamp not null
);
create index username on users (username);


select * FROM users;
select * from friendship;
select * from url;


