CREATE DATABASE test;

use test;

CREATE TABLE dummy_table (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    name TEXT
);

INSERT into dummy_table (name) values ('test')