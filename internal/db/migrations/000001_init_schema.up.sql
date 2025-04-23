CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255),
                       email VARCHAR(255),
                       password VARCHAR(255),
                       role VARCHAR(50),
                       created_at TIMESTAMP,
                       updated_at TIMESTAMP,
                       deleted_at TIMESTAMP
);

CREATE TABLE courses (
                         id SERIAL PRIMARY KEY,
                         title VARCHAR(255),
                         description TEXT,
                         teacher_id INTEGER REFERENCES users(id),
                         created_at TIMESTAMP,
                         updated_at TIMESTAMP,
                         deleted_at TIMESTAMP
);

CREATE TABLE lessons (
                         id SERIAL PRIMARY KEY,
                         title VARCHAR(255),
                         content TEXT,
                         course_id INTEGER REFERENCES courses(id),
                         created_at TIMESTAMP,
                         updated_at TIMESTAMP,
                         deleted_at TIMESTAMP
);
