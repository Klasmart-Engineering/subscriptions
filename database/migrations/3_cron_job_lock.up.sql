CREATE TABLE cron_job_lock (
    name VARCHAR(255) NOT NULL,
    locked_by VARCHAR(255) NOT NULL,
    locked_until TIMESTAMP NOT NULL,
    PRIMARY KEY (name)
);

INSERT INTO cron_job_lock VALUES ('access-log-compaction', 'na', now());
