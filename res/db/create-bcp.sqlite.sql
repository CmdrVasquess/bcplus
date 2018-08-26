CREATE TABLE l10nsubj (
  id INT64 PRIMARY KEY
, dom TEXT NOT NULL
, val TEXT NOT NULL
, unique (dom, val)
);

CREATE TABLE l10nterm (
  subj INT64 REFERENCES l10nsubj(id)
, lang TEXT
, term TEXT NOT NULL
, PRIMARY KEY (subj, lang)
);
