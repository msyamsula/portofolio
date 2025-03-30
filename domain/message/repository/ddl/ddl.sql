CREATE TABLE messages (
	id bigserial PRIMARY KEY NOT NULL,
	sender_id bigserial NOT NULL,
	receiver_id bigserial NOT NULL,
	text varchar(10000) NOT NULL,
	create_time timestamp DEFAULT current_timestamp NOT NULL,
	update_time timestamp NOT NULL DEFAULT current_timestamp,
);
create index sender_id on messages(sender_id);
create index receiver_id on messages(receiver_id);
create index create_time on messages(create_time);

alter table messages add column is_read boolean not null default false;

create table unread (
	id bigserial primary key not null,
	sender_id bigint not null,
	receiver_id bigint not null,
	unread bigint not null default 0,
	create_time timestamp default current_timestamp not null,
	update_time timestamp default current_timestamp not null,
	CONSTRAINT unique_unread UNIQUE (sender_id, receiver_id)
);

create index unread_sender_id on unread(sender_id);
create index unread_receiver_id on unread(receiver_id);