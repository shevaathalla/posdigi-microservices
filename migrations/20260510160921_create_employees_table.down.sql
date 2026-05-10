-- Drop indexes first
DROP INDEX IF EXISTS idx_employees_deleted_at;
DROP INDEX IF EXISTS idx_employees_status;
DROP INDEX IF EXISTS idx_employees_manager;
DROP INDEX IF EXISTS idx_employees_department;
DROP INDEX IF EXISTS idx_employees_code;
DROP INDEX IF EXISTS idx_employees_user_id;

-- Drop the employees table
DROP TABLE IF EXISTS employees;
