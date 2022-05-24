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

CREATE TABLE if not exists subscription_account_product
(
    subscription_id uuid                   NOT NULL,
    product         varchar                NOT NULL,
    type            character varying(255) NOT NULL,
    threshold       int                    NULL,
    action          character varying(255) NOT NULL,
    FOREIGN KEY (type) REFERENCES subscription_type (name),
    FOREIGN KEY (action) REFERENCES subscription_action (name)
);


INSERT INTO subscription_account_product (subscription_id, product, type, action)
VALUES ((SELECT id FROM subscription_account), 'Simple Teacher Module', 'Uncapped', 'API Call'),
       ((SELECT id FROM subscription_account), 'Homework', 'Uncapped', 'API Call');

CREATE TABLE if not exists subscription_account_log
(
    subscription_id uuid,
    action_type     varchar                                            NOT NULL,
    usage           int                                                NOT NULL,
    product_name    varchar                                            NOT NULL,
    interaction_at  timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    valid_usage  boolean NOT NULL DEFAULT TRUE,
    FOREIGN KEY (subscription_id) REFERENCES subscription_account (id)
);
