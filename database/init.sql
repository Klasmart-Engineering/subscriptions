CREATE TABLE if not exists subscription_type
(
    id   SERIAL,
    name character varying(255) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);

INSERT INTO subscription_type (name)
VALUES ('Capped'),
       ('Uncapped');


CREATE TABLE if not exists subscription_action
(
    name        character varying(255) NOT NULL UNIQUE,
    description character varying(255) NOT NULL UNIQUE,
    unit        character varying(255) NOT NULL UNIQUE,
    PRIMARY KEY (name)
);


INSERT INTO subscription_action (name, description, unit)
VALUES ('API Call', 'User interaction with public API Gateway', 'HTTP Requests'),
       ('UserLogin', 'User Login Action', 'Account Active');


CREATE TABLE if not exists subscription_state
(
    id   SERIAL,
    name character varying(255) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);

INSERT INTO subscription_state (name)
VALUES ('Active'),
       ('Inactive'),
       ('Disabled'); -- TODO Disabled or Deleted?


CREATE TABLE if not exists subscription_account
(
    id                    uuid DEFAULT gen_random_uuid() UNIQUE,
    account_id            UUID NOT NULL,
    last_processed        timestamp with time zone,
    run_frequency_minutes int    NOT NULL,
    state                 serial NOT NULL,
    FOREIGN KEY (state) REFERENCES subscription_state (id),
    PRIMARY KEY (id, account_id)
);

INSERT INTO subscription_account (id, account_id, run_frequency_minutes, state)
VALUES (gen_random_uuid(), gen_random_uuid(), 5, 1);
