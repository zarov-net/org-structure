-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INTEGER REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT chk_name_length CHECK (char_length(trim(name)) BETWEEN 1 AND 200)
);

CREATE UNIQUE INDEX idx_departments_name_parent ON departments(name, parent_id);
CREATE INDEX idx_departments_parent_id ON departments(parent_id);

CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    department_id INTEGER NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT chk_full_name_length CHECK (char_length(trim(full_name)) BETWEEN 1 AND 200),
    CONSTRAINT chk_position_length CHECK (char_length(trim(position)) BETWEEN 1 AND 200)
);

CREATE INDEX idx_employees_department_id ON employees(department_id);
CREATE INDEX idx_employees_created_at ON employees(created_at);
CREATE INDEX idx_employees_full_name ON employees(full_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd