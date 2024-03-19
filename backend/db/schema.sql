-- filename | status | (date ?) | path

CREATE TABLE IF NOT EXISTS pipeline(
    id INTEGER PRIMARY KEY,
    filename TEXT NOT NULL,
    path TEXT NOT NULL,
    status TEXT NOT NULL
)