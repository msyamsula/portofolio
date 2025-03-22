CREATE TABLE messages (
	id bigserial PRIMARY KEY NOT NULL,
	sender_id bigserial NOT NULL,
	receiver_id bigserial NOT NULL,
	text varchar(10000) NOT NULL,
	create_time timestamp DEFAULT current_timestamp NOT NULL,
	update_time timestamp NOT NULL DEFAULT current_timestamp,
	FOREIGN KEY (sender_id) REFERENCES users(id) on delete cascade,
	FOREIGN KEY (receiver_id) REFERENCES users(id) on delete cascade
);
create index sender_id on messages(sender_id);
create index receiver_id on messages(receiver_id);
create index create_time on messages(create_time);