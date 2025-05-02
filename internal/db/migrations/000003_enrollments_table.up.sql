CREATE TABLE enrollments (
                             id SERIAL PRIMARY KEY,
                             user_id INTEGER NOT NULL,
                             course_id INTEGER NOT NULL,
                             created_at TIMESTAMP DEFAULT now(),
                             updated_at TIMESTAMP DEFAULT now(),
                             deleted_at TIMESTAMP
);
