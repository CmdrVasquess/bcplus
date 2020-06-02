--*- mode: sql; sql-product: sqlite; -*-
CREATE TABLE info (
  key TEXT PRIMARY KEY
, value TEXT NOT NULL
);

INSERT INTO info (key, value) VALUES ('version', '0.1.0');

CREATE TABLE sys (
  addr INTEGER PRIMARY KEY
, name TEXT NOT NULL
, x    REAL
, y    REAL
, z    REAL
);

CREATE UNIQUE INDEX IF NOT EXISTS upsysname
ON sys (upper(name));

CREATE INDEX IF NOT EXISTS sysx ON sys(x);
CREATE INDEX IF NOT EXISTS sysx ON sys(y);
CREATE INDEX IF NOT EXISTS sysx ON sys(z);

CREATE TABLE cmdr (
  id INTEGER PRIMARY KEY
, name TEXT NOT NULL UNIQUE
);

CREATE TABLE visit (
  cmdr INTEGER NOT NULL REFERENCES cmdr(id)
, sys  INTEGER NOT NULL REFERENCES sys(addr)
, t    TEXT NOT NULL
);

CREATE TABLE sysloc (
  sys INTEGER REFERENCES sys(addr) 
, id INTEGER -- id >= 0 => BodyID from journal
, name TEXT NULL
, center INTEGER NULL -- id within same sys
, type TEXT NOT NULL
, PRIMARY KEY (sys, id)
);
