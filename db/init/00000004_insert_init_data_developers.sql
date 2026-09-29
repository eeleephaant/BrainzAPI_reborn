\connect brainz_developers;


INSERT INTO
    role (id, title)
VALUES
    (0, 'user');

INSERT INTO
    role (id, title)
VALUES
    (1, 'admin');

INSERT INTO
    role_api_key_permission (role_id, action, institution_id)
VALUES
    (0, 'read', NULL),
    (1, 'read', NULL),
    (1, 'write', NULL);
