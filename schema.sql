create type torrent_file as (
  "path" text,
  "length" bigint
);

create table torrents (
  "info_hash" char(40) not null primary key,
  "name" text not null,
  "comment" text,
  "size" bigint not null,
  "first_seen" timestamp not null,
  "files" torrent_file[] not null,
  "trackers" text[] not null,
  "locations" text[] not null,
  "last_scrape" timestamp
);

create index name_idx on torrents using gin (name gin_trgm_ops);

create table old_ids (
  "old_id" bigint not null primary key,
  "info_hash" char(40) not null
);

create table scrapes (
  "info_hash" char(40) not null references "torrents" ("info_hash"),
  "tracker" text not null,
  "time" timestamp not null,
  "success" boolean not null,
  "downloaded" bigint,
  "complete" bigint,
  "incomplete" bigint
);

create index "info_hash_idx" on "scrapes" ("info_hash");
