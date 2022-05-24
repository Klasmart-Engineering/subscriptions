INSERT INTO subscription_account (id, account_id, last_processed, run_frequency_minutes, state) VALUES
    ('2f797c16-053e-41ab-b40d-24356480e61e', gen_random_uuid(), NULL, 30, 2);

INSERT INTO subscription_account_product (subscription_id, product, type, action, threshold) VALUES
    ('2f797c16-053e-41ab-b40d-24356480e61e', 'Test Product', 'Uncapped', 'API Call', 30);

INSERT INTO subscription_account_log (subscription_id, action_type, usage, product_name, interaction_at) VALUES
    ('2f797c16-053e-41ab-b40d-24356480e61e', 'API Call', 20, 'Test Product', NOW());
