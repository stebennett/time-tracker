-- 1. Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_name TEXT NOT NULL
);

-- 2. Add task_id column to time_entries (nullable for now)
ALTER TABLE time_entries ADD COLUMN task_id INTEGER;

-- 3. Insert unique descriptions as tasks
INSERT INTO tasks (task_name)
SELECT DISTINCT description FROM time_entries WHERE description IS NOT NULL;

-- 4. Update time_entries to reference the correct task_id
UPDATE time_entries
SET task_id = (SELECT id FROM tasks WHERE tasks.task_name = time_entries.description)
WHERE description IS NOT NULL;

-- 5. Make task_id NOT NULL (SQLite doesn't support altering column nullability directly, so skip for now)

-- 6. Remove description column (SQLite doesn't support DROP COLUMN directly, so recreate table)
CREATE TABLE time_entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    task_id INTEGER NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

INSERT INTO time_entries_new (id, start_time, end_time, task_id)
SELECT id, start_time, end_time, task_id FROM time_entries;

DROP TABLE time_entries;
ALTER TABLE time_entries_new RENAME TO time_entries; 