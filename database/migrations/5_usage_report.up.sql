CREATE TABLE usage_report (
    id UUID NOT NULL,
    subscription_id UUID NOT NULL,
    year INT NOT NULL,
    month INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (subscription_id) REFERENCES subscription(id)
);

CREATE TABLE usage_report_instance (
   id UUID NOT NULL,
   usage_report_id UUID NOT NULL,
   requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
   athena_query_id VARCHAR(255) NOT NULL,
   completed_at TIMESTAMP WITH TIME ZONE,
   PRIMARY KEY (id),
   FOREIGN KEY (usage_report_id) REFERENCES usage_report(id)
);

CREATE TABLE usage_report_instance_product (
    usage_report_instance_id UUID NOT NULL,
    product VARCHAR(255) NOT NULL,
    value INT NOT NULL,
    PRIMARY KEY (usage_report_instance_id, product),
    FOREIGN KEY (usage_report_instance_id) REFERENCES usage_report_instance(id)
);
