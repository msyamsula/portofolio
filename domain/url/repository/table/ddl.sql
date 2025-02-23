create table url (
	id bigserial primary key not null,
	short_url varchar(50) not null unique,
	long_url varchar(5000) not null,
	create_time date default current_timestamp not null
);

create index short_url_index on url (short_url);