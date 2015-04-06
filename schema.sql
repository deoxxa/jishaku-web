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
  "locations" text[] not null
);

create index name_idx on torrents using gin (name gin_trgm_ops);

create table old_ids (
  "old_id" bigint not null primary key,
  "info_hash" char(40) not null
);
