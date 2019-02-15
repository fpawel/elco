PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS work
(
  work_id    INTEGER   NOT NULL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL UNIQUE DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  name       TEXT      NOT NULL
);

CREATE TABLE IF NOT EXISTS entry
(
  entry_id   INTEGER   NOT NULL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL UNIQUE DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  message    TEXT      NOT NULL,
  level      TEXT      NOT NULL,
  work_id    INTEGER   NOT NULL,
  FOREIGN KEY (work_id) REFERENCES work (work_id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS work_info AS
SELECT *,
       EXISTS(
           SELECT level IN ('panic', 'error', 'fatal') FROM entry WHERE entry.work_id = work.work_id) AS error_occurred,
       CAST(STRFTIME('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER)                              AS year,
       CAST(STRFTIME('%m', DATETIME(created_at, '+3 hours')) AS INTEGER)                              AS month,
       CAST(STRFTIME('%d', DATETIME(created_at, '+3 hours')) AS INTEGER)                              AS day
FROM work;