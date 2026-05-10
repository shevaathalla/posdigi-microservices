CREATE TABLE employees (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(36) NOT NULL,
    employee_code VARCHAR(20) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    department VARCHAR(50),
    position VARCHAR(50),
    salary DECIMAL(10,2),
    hire_date DATE NOT NULL,
    employment_status VARCHAR(20) DEFAULT 'active',
    manager_id VARCHAR(36),
    emergency_contact VARCHAR(100),
    emergency_phone VARCHAR(20),
    address TEXT,
    profile_image VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,

    CONSTRAINT fk_employee_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_employee_manager FOREIGN KEY (manager_id) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_employees_user_id ON employees(user_id);
CREATE UNIQUE INDEX idx_employees_code ON employees(employee_code);
CREATE INDEX idx_employees_department ON employees(department);
CREATE INDEX idx_employees_manager ON employees(manager_id);
CREATE INDEX idx_employees_status ON employees(employment_status);
CREATE INDEX idx_employees_deleted_at ON employees(deleted_at);

-- Add comment for documentation
COMMENT ON TABLE employees IS 'Detailed employee information and employment data';
COMMENT ON COLUMN employees.employee_code IS 'Unique employee identifier (e.g., EMP001)';
COMMENT ON COLUMN employees.user_id IS 'Foreign key reference to users table';
COMMENT ON COLUMN employees.manager_id IS 'Self-referencing foreign key for organizational hierarchy';
COMMENT ON COLUMN employees.employment_status IS 'Values: active, terminated, on_leave, suspended';
