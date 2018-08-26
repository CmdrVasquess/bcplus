CREATE TABLE meta (
  key TEXT PRIMARY KEY
, val TEXT
);

INSERT INTO meta (key, val) VALUES
  ('version', '1')
;

CREATE TABLE system (
  id INTEGER PRIMARY KEY
, name TEXT NOT NULL UNIQUE
, x REAL
, y REAL
, z REAL
);

CREATE INDEX sysXcoo ON system(x);

CREATE INDEX sysYcoo ON system(y);

CREATE INDEX sysZcoo ON system(z);

CREATE TABLE syspart (
  id INTEGER PRIMARY KEY
, sys INTEGER NOT NULL
              REFERENCES system(id)
, typ INTEGER NOT NULL
, name TEXT NOT NULL
, dfc INTEGER
, tto INTEGER
, hgt REAL
, lat REAL
, lon REAL
, UNIQUE (sys, name)
);

CREATE TABLE resource (
  loc INTEGER NOT NULL
              REFERENCES syspart(id)
, name TEXT NOT NULL
, freq REAL
);
