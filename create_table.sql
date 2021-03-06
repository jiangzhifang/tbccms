CREATE TABLE coursewares (course_code VARCHAR(100) primary key,course_title varchar(200) NOT NULL,created timestamp(0) NOT NULL,active BOOLEAN NOT NULL DEFAULT TRUE);

CREATE INDEX idx_coursewares_created ON coursewares(created);

CREATE TABLE courseware_files (courseware_filename varchar(200) primary key, course_code VARCHAR(100));

CREATE INDEX idx_courseware_files_course_code ON courseware_files(course_code);


CREATE TABLE users (id serial primary key, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, hashed_password CHAR(60) NOT NULL, created timestamp(0) NOT NULL, active BOOLEAN NOT NULL DEFAULT TRUE);

ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);
