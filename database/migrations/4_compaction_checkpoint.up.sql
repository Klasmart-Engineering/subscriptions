ALTER TABLE subscription ADD COLUMN created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

CREATE TABLE compaction_checkpoint (
    subscription_id UUID NOT NULL,
    succeeded_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (subscription_id)
);
