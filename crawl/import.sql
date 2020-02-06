-- This SQL program shows how to import CSV files into MySQL tables.
-- 
-- To run a MySQL instnace locally, you can use SQLFlow's Docker image:
--    docker run --rm -it -v $PWD:/work sqlflow/sqlflow bash
-- 
-- In the Docker container, type the following command to start MySQL server:
--    service mysql start
-- 
-- MySQL server can import only files in `/var/lib/mysql-files`, so we
-- need to move the files there before running this program.
-- 
-- To run this program from the command line:
--     mysql -u root -p < import.sql
CREATE DATABASE IF NOT EXISTS sqlflow;

USE sqlflow;

DROP TABLE IF EXISTS pulls;

CREATE TABLE IF NOT EXISTS pulls (
       number INT NOT NULL,
       author TEXT,
       title TEXT,
       comment TEXT,
       PRIMARY KEY (number)
);

LOAD DATA INFILE "/var/lib/mysql-files/sqlflow/pulls.csv"
  INTO TABLE pulls
  FIELDS TERMINATED BY ',' ;

DROP TABLE IF EXISTS comments;

CREATE TABLE IF NOT EXISTS comments (
       id INT NOT NULL AUTO_INCREMENT,
       number INT NOT NULL,
       author TEXT,
       comment TEXT,
       PRIMARY KEY (id)
);

LOAD DATA INFILE "/var/lib/mysql-files/sqlflow/comments.csv"
  INTO TABLE comments
  FIELDS TERMINATED BY ','
  (number, author, comment);
