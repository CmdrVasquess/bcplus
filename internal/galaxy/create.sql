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

CREATE TABLE sysloc (
  id INTEGER PRIMARY KEY
, sys INTEGER NOT NULL REFERENCES sys(addr) 
, body INTEGER NULL
, name TEXT NULL
, center INTEGER NULL REFERENCES sysloc(addr)
, type TEXT NOT NULL
, UNIQUE (sys, body)
);
