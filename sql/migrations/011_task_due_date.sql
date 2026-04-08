-- Migration: Add due_date_at column to task table
-- Allows tasks to have an optional deadline

ALTER TABLE "task" ADD COLUMN IF NOT EXISTS due_date_at TIMESTAMP with time zone;
