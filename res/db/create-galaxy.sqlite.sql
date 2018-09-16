-- since 1
CREATE TABLE meta (
  key TEXT PRIMARY KEY
, val TEXT
);

-- since 1
INSERT INTO meta (key, val) VALUES
  ('version', '1')
;

-- since 2
UPDATE meta SET val='2' WHERE key='version';

-- since 1
CREATE TABLE system (
  id INTEGER PRIMARY KEY
, name TEXT NOT NULL
, x REAL
, y REAL
, z REAL
, UNIQUE (name COLLATE NOCASE)
);

-- since 1
CREATE INDEX sysXcoo ON system(x);

-- since 1
CREATE INDEX sysYcoo ON system(y);

-- since 1
CREATE INDEX sysZcoo ON system(z);

-- since 1
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

-- since 1
CREATE TABLE resource (
  loc INTEGER NOT NULL
              REFERENCES syspart(id)
, name TEXT NOT NULL
, freq REAL
);

-- since 2
CREATE TABLE tag (
  id INTEGER PRIMARY KEY
, name TEXT NOT NULL
, parent INTEGER REFERENCES tag(id)
, path TEXT NOT NULL
, UNIQUE (parent, name)
);

-- since 2
CREATE TABLE poi (
  id INTEGER PRIMARY KEY
, title TEXT NOT NULL
, desc TEXT
);

-- since 2
CREATE TABLE poitags (
  poi INTEGER REFERENCES poi(id)
, tag INTEGER REFERENCES tag(id)
, PRIMARY KEY (poi, tag)
);
