CREATE TABLE IF NOT EXISTS bookings
(
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL,
    boxName TEXT NOT NULL,
    startsAt TEXT NOT NULL,
    expiresAt INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_expiresAt ON bookings (expiresAt);