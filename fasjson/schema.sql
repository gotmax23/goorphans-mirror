CREATE TABLE IF NOT EXISTS fas_user (
    user_name TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    exp REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS fas_group (
    group_name TEXT PRIMARY KEY,
    exp REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS group_member (
    group_name TEXT,
    user_name TEXT,
    FOREIGN KEY (group_name) REFERENCES fas_group(group_name) ON DELETE CASCADE,
    PRIMARY KEY (group_name, user_name)
);
