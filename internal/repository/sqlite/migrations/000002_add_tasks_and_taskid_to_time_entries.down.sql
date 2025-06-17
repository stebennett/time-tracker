-- 1. Add description column back to time_entries (nullable for now)
ALTER TABLE time_entries ADD COLUMN description TEXT;

-- 2. Copy task name back to description
UPDATE time_entries
SET description = (SELECT task_name FROM tasks WHERE tasks.id = time_entries.task_id);

-- 3. Recreate time_entries without task_id and foreign key
CREATE TABLE time_entries_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    description TEXT
);

INSERT INTO time_entries_old (id, start_time, end_time, description)
SELECT id, start_time, end_time, description FROM time_entries;

DROP TABLE time_entries;
ALTER TABLE time_entries_old RENAME TO time_entries;

-- 4. Drop tasks table
DROP TABLE IF EXISTS tasks; 