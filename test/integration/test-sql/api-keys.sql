INSERT INTO api_key (owner, api_key) VALUES ('Test', 'valid-key-no-permission');

INSERT INTO api_key (owner, api_key) VALUES ('Test2', 'valid-key-with-permission');
INSERT INTO api_key_permission(owner, permission) VALUES ('Test2', 'create-subscription');
