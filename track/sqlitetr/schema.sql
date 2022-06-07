PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  track_code VARCHAR,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS track_points (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  track_id INTEGER,
  latitude FLOAT,
  longitude FLOAT,
  altitude INTEGER,
  heading INTEGER,
  groundspeed INTEGER,
  ts DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS track_track_code_idx ON tracks (track_code);
CREATE INDEX IF NOT EXISTS track_points_ts_idx ON track_points (ts);
