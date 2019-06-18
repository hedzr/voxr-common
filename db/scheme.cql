



CREATE KEYSPACE IF NOT EXISTS gwkuul
  WITH REPLICATION = {'class': 'SimpleStrategy','replication_factor':1};



use gwkuul;

DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
  name text PRIMARY KEY,
  email text,
  mobile text,
  password text,
  time_created timestamp,
  time_updated timestamp,

  token1 text,
  followers list<text>,

  s_blocked boolean,
  s_forbidden boolean,
  s_deleted boolean,
);

CREATE INDEX ON users (email);
CREATE INDEX ON users (mobile);

INSERT INTO users (name,email,mobile,password,time_created) VALUES ('admin','admin@example.com','13012345678', '', toTimestamp(now()) );



CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  name text,
  email text,
  mobile text,
  password text,
  created_time timestamp,

  token1 text,
  followers list<text>,

  s_blocked boolean,
  s_forbidden boolean,
  s_deleted boolean,
);


INSERT INTO users (id,name,email,mobile,password,created_time) VALUES (uuid(),'admin','admin@example.com','13012345678', '', toTimestamp(now()) );










