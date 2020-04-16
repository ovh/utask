-- +migrate Up

-- Add a missing 'ON DELETE CASCADE' on resolution->task FOREIGN KEY
ALTER TABLE resolution DROP CONSTRAINT resolution_id_task_fkey, ADD CONSTRAINT resolution_id_task_fkey FOREIGN KEY (id_task) REFERENCES task(id) ON DELETE CASCADE;

-- +migrate Down

ALTER TABLE resolution DROP CONSTRAINT resolution_id_task_fkey, ADD CONSTRAINT resolution_id_task_fkey FOREIGN KEY (id_task) REFERENCES task(id);
