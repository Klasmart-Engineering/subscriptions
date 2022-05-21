INSERT INTO subscription_account (id, last_processed, run_frequency_minutes, state) VALUES
    ('2f797c16-053e-41ab-b40d-24356480e61e', NULL, 30, 1),
    ('4c7e63ee-43a9-486d-ae38-d3e086593613', NULL, 30, 2),
    ('5859fc2f-9eed-4e09-b653-5a63d3b100c0', NULL, 30, 3);

INSERT INTO subscription_account_product (subscription_id, product, type, action, threshold) VALUES
     ('2f797c16-053e-41ab-b40d-24356480e61e', 'Test Product', 'Uncapped', 'API Call', 30),
     ('4c7e63ee-43a9-486d-ae38-d3e086593613', 'Test Product', 'Uncapped', 'API Call', 30),
     ('5859fc2f-9eed-4e09-b653-5a63d3b100c0', 'Test Product', 'Uncapped', 'API Call', 30);
