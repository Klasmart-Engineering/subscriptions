CREATE TABLE if not exists subscription_type
(
    id         SERIAL PRIMARY KEY,
    name       character varying(255)                             NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO subscription_type (id, name)
VALUES (1, 'Capped'),
       (2, 'Uncapped');


CREATE TABLE if not exists subscription_action
(
    name        character varying(255) NOT NULL UNIQUE PRIMARY KEY,
    description character varying(255) NOT NULL UNIQUE,
    unit        character varying(255) NOT NULL UNIQUE
);


INSERT INTO subscription_action (name, description, unit)
VALUES ('API Call', 'User interaction with public API Gateway', 'HTTP Requests');

CREATE TABLE if not exists subscription_account
(
    id                    SERIAL PRIMARY KEY,
    account_holder_id     int NOT NULL,
    last_processed        timestamp with time zone,
    run_frequency_minutes int NOT NULL,
    active                boolean NOT NULL
);


INSERT INTO subscription_account (id, account_holder_id, run_frequency_minutes, active)
VALUES (1, 123, 5, true);

CREATE TABLE if not exists subscription_account_product
(
    subscription_id SERIAL                 NOT NULL,
    product         varchar                NOT NULL,
    name            character varying(255) NOT NULL,
    threshold       int                    NOT NULL,
    action          character varying(255) NOT NULL
);


INSERT INTO subscription_account_product (subscription_id, product, name, threshold, action)
VALUES
       (1, 'Simple Teacher Module', 'Capped', 10, 'API Call'),
       (1, 'Homework', 'Capped', 3,  'API Call');

CREATE TABLE if not exists subscription_account_user_log
(
    GUID                    int,
    subscription_account_id int                                                NOT NULL,
    action_type             varchar                                            NOT NULL,
    usage                   int                                                NOT NULL,
    product                 varchar                                            NOT NULL,
    interaction_at          timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (GUID, action_type, product, interaction_at)
--     CONSTRAINT account_holder_id FOREIGN KEY (subscription_account_id) REFERENCES subscription_account (account_holder_id)
);
