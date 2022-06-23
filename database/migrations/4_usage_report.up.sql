CREATE TABLE usage_report (
    id UUID NOT NULL,
    subscription_id UUID NOT NULL,
    year INT NOT NULL,
    month INT NOT NULL,
    processing_athena_query_id VARCHAR(255),
    PRIMARY KEY (id),
    FOREIGN KEY (subscription_id) REFERENCES subscription (id)
);

CREATE TABLE usage_report_product (
    usage_report_id UUID NOT NULL,
    product VARCHAR(255) NOT NULL,
    value INT NOT NULL,
    PRIMARY KEY (usage_report_id, product)
);