/******** USERS *********/

CREATE OR REPLACE TABLE users
(
    id      VARCHAR(36) PRIMARY KEY,
    created DATETIME NOT NULL DEFAULT current_timestamp()
);

CREATE OR REPLACE TABLE user_keys
(
    hash       VARCHAR(64) NOT NULL,
    id         BLOB        NOT NULL,
    aaguid     VARCHAR(36) NOT NULL,
    user_id    VARCHAR(36) NOT NULL,
    credential BLOB        NOT NULL,
    created    DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    last_used  DATETIME,
    INDEX user_keys_hash_user_id (hash, user_id),
    INDEX user_keys_user_id (user_id),
    CONSTRAINT FOREIGN KEY user_keys_user_id (user_id) REFERENCES users (id) ON DELETE CASCADE
);


CREATE OR REPLACE PROCEDURE create_user(IN user_id VARCHAR(36))
BEGIN
    INSERT INTO users(id, created) VALUES (user_id, current_timestamp());
    SELECT user_id;
END;

CREATE OR REPLACE PROCEDURE delete_user(IN user_id VARCHAR(36))
BEGIN
    DELETE FROM users WHERE id = user_id;
END;

CREATE OR REPLACE PROCEDURE delete_user_key(in user_id VARCHAR(36), IN hash VARCHAR(64))
BEGIN
    DELETE FROM user_keys WHERE user_keys.user_id = user_id AND user_keys.hash = hash;
END;

CREATE OR REPLACE PROCEDURE get_user(IN user_id VARCHAR(36))
BEGIN
    SELECT id, created FROM users WHERE id = user_id;
END;

CREATE OR REPLACE PROCEDURE list_users()
BEGIN
    SELECT id, created FROM users;
END;

CREATE OR REPLACE PROCEDURE list_users_with_keys()
BEGIN
    SELECT users.id             as userId,
           users.created        as userCreated,
           user_keys.hash       as keyHash,
           user_keys.id         as keyId,
           user_keys.aaguid     as keyAAGUID,
           user_keys.created    as keyCreated,
           user_keys.last_used  as keyUsed,
           user_keys.credential as keyCredential
    FROM users
             LEFT JOIN user_keys ON users.id = user_keys.user_id
    ORDER BY users.id;
END;

CREATE OR REPLACE PROCEDURE get_user_with_keys(IN user_id VARCHAR(36))
BEGIN
    SELECT users.id             as userId,
           users.created        as userCreated,
           user_keys.hash       as keyHash,
           user_keys.id         as keyId,
           user_keys.aaguid     as keyAAGUID,
           user_keys.created    as keyCreated,
           user_keys.last_used  as keyUsed,
           user_keys.credential as keyCredential
    FROM users
             LEFT JOIN user_keys ON users.id = user_keys.user_id
    WHERE users.id = user_id
    ORDER BY users.id;
END;


CREATE OR REPLACE PROCEDURE get_user_keys(IN user_id VARCHAR(36))
BEGIN
    SELECT hash, id, user_id, credential, created, last_used FROM user_keys WHERE user_keys.user_id = user_id;
END;

CREATE OR REPLACE PROCEDURE get_user_key(IN user_id VARCHAR(36), IN hash VARCHAR(64))
BEGIN
    SELECT hash, id, aaguid, user_id, credential, created, last_used
    FROM user_keys
    WHERE user_keys.user_id = user_id
      AND user_keys.hash = hash;
END;

CREATE OR REPLACE PROCEDURE add_user_key(IN hash VARCHAR(64), IN key_id BLOB, IN aaguid VARCHAR(36),
                                         IN user_id VARCHAR(36),
                                         IN credential BLOB)
BEGIN
    INSERT INTO user_keys(hash, id, aaguid, user_id, credential) VALUES (hash, key_id, aaguid, user_id, credential);
END;

CREATE OR REPLACE PROCEDURE ping_user_key(IN hash VARCHAR(64), IN credential BLOB)
BEGIN
    UPDATE user_keys u SET last_used = current_timestamp(), u.credential = credential WHERE u.hash = hash;
END;

CREATE OR REPLACE PROCEDURE get_key(IN hash VARCHAR(64))
BEGIN
    SELECT hash, id, aaguid, user_id, credential, created, last_used FROM user_keys WHERE user_keys.hash = hash;
END;

/******** SERVER KEYS *********/

CREATE OR REPLACE TABLE server_keys
(
    kid         VARCHAR(16) PRIMARY KEY,
    type        ENUM ('EC', 'RSA')                                          NOT NULL DEFAULT 'RSA',
    alg         ENUM ('ES256', 'ES384', 'ES512', 'RS256', 'RS384', 'RS512') NOT NULL DEFAULT 'RS256',
    created     DATETIME                                                    NOT NULL DEFAULT current_timestamp(),
    private_key VARCHAR(2048)                                               NOT NULL,
    public_key  VARCHAR(2048)                                               NOT NULL
);
CREATE OR REPLACE PROCEDURE get_available_algorithms()
BEGIN
    SELECT DISTINCT alg FROM server_keys;
END;

CREATE OR REPLACE PROCEDURE create_server_key(IN kid VARCHAR(36), IN type VARCHAR(15), IN alg VARCHAR(15),
                                              IN private_key VARCHAR(2048),
                                              IN public_key VARCHAR(2048))
BEGIN
    INSERT INTO server_keys(kid, alg, type, private_key, public_key) VALUES (kid, alg, type, private_key, public_key);
    SELECT kid;
END;

CREATE OR REPLACE PROCEDURE get_server_key(IN id VARCHAR(36))
BEGIN
    SELECT kid, type, alg, created, private_key, public_key FROM server_keys WHERE kid = id;
END;

CREATE OR REPLACE PROCEDURE get_server_key_with_alg(IN alg ENUM ('ES256', 'ES384', 'ES512', 'RS256', 'RS384', 'RS512'))
BEGIN
    SELECT kid, type, alg, created, private_key, public_key
    FROM server_keys s
    WHERE s.alg = alg
    ORDER BY created DESC
    LIMIT 1;
END;

CREATE OR REPLACE PROCEDURE list_server_keys()
BEGIN
    SELECT kid, type, alg, created, private_key, public_key FROM server_keys;
END;

CREATE OR REPLACE PROCEDURE delete_server_key(IN id VARCHAR(36))
BEGIN
    DELETE FROM server_keys WHERE kid = id;
END;

/******** APPLICATIONS *********/

CREATE OR REPLACE TABLE applications
(
    id                    VARCHAR(36) PRIMARY KEY,
    created               DATETIME                                                    NOT NULL DEFAULT current_timestamp(),
    name                  VARCHAR(100)                                                NOT NULL,
    secret                VARCHAR(36)                                                 NOT NULL,
    description           VARCHAR(1024)                                               NOT NULL,
    icon                  VARCHAR(2048)                                               NOT NULL,
    is_admin              BOOLEAN                                                     NOT NULL DEFAULT FALSE,
    ciba_mode             ENUM ('poll', 'push', 'ping')                               NOT NULL DEFAULT 'poll',
    notification_endpoint VARCHAR(2048)                                               NOT NULL DEFAULT '',
    alg                   ENUM ('ES256', 'ES384', 'ES512', 'RS256', 'RS384', 'RS512') NOT NULL DEFAULT 'RS256',
    kid                   VARCHAR(16)                                                 NOT NULL,
    CONSTRAINT FOREIGN KEY applications_kid (kid) REFERENCES server_keys (kid)
);

CREATE OR REPLACE TABLE application_redirect_urls
(
    application_id VARCHAR(36)  NOT NULL,
    url            VARCHAR(250) NOT NULL,
    CONSTRAINT FOREIGN KEY application_redirect_urls_application_id (application_id) REFERENCES applications (id) ON DELETE CASCADE
);

CREATE OR REPLACE PROCEDURE create_app(IN app_id VARCHAR(36), IN secret VARCHAR(36), IN app_name VARCHAR(100),
                                       IN description VARCHAR(250), IN icon VARCHAR(1024),
                                       IN ciba_mode VARCHAR(20),
                                       IN notification_endpoint VARCHAR(2048),
                                       IN alg ENUM ('ES256', 'ES384', 'ES512', 'RS256', 'RS384', 'RS512'),
                                       IN kid VARCHAR(16), IN is_admin BOOLEAN)
BEGIN
    INSERT INTO applications (id, name, secret, description, icon, alg, kid, is_admin, ciba_mode, notification_endpoint)
    VALUES (app_id, app_name, secret, description, icon, alg, kid, is_admin, ciba_mode, notification_endpoint);
    SELECT app_id, secret;
END;

CREATE OR REPLACE PROCEDURE create_app_redirect_url(IN app_id VARCHAR(36), IN url VARCHAR(250))
BEGIN
    INSERT INTO application_redirect_urls(application_id, url) VALUES (app_id, url);
END;

CREATE OR REPLACE PROCEDURE get_app(IN app_id VARCHAR(36))
BEGIN
    SELECT id,
           created,
           name,
           secret,
           description,
           icon,
           ciba_mode,
           notification_endpoint,
           is_admin,
           alg,
           kid
    FROM applications
    WHERE id = app_id
    LIMIT 1;
END;

CREATE OR REPLACE PROCEDURE get_app_redirect_urls(IN app_id VARCHAR(36))
BEGIN
    SELECT url FROM application_redirect_urls WHERE application_redirect_urls.application_id = app_id;
END;

CREATE OR REPLACE TABLE application_user_meta
(
    app_id    VARCHAR(36) NOT NULL,
    user_id   VARCHAR(36) NOT NULL,
    last_auth DATETIME,
    CONSTRAINT FOREIGN KEY application_user_meta_app_id (app_id) REFERENCES applications (id) ON DELETE CASCADE,
    CONSTRAINT FOREIGN KEY application_user_meta_user_id (user_id) REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (app_id, user_id)
);

CREATE OR REPLACE PROCEDURE update_user_auth_time(IN user_id VARCHAR(36), IN app_id VARCHAR(36))
BEGIN
    INSERT INTO application_user_meta (app_id, user_id, last_auth)
    VALUES (app_id, user_id, current_timestamp())
    ON DUPLICATE KEY UPDATE last_auth = current_timestamp();
END;

CREATE OR REPLACE PROCEDURE get_user_auth_time(IN user_id VARCHAR(36), IN app_id VARCHAR(36))
BEGIN
    SELECT last_auth FROM application_user_meta a WHERE a.user_id = user_id AND a.app_id = app_id;
END;


/******* CHALLENGES *******/

CREATE OR REPLACE TABLE challenges
(
    created        DATETIME                                                     NOT NULL DEFAULT current_timestamp(),
    id             VARCHAR(36) PRIMARY KEY,
    type           VARCHAR(36)                                                  NOT NULL,
    app_id         VARCHAR(36),
    expire         DATETIME                                                     NOT NULL,
    public_data    BLOB                                                         NOT NULL,
    private_data   BLOB                                                         NOT NULL,

    signature_text TEXT                                                         NOT NULL COLLATE utf8mb4_unicode_ci,
    signature_data BLOB,
    nonce          VARCHAR(16)                                                  NOT NULL COLLATE utf8mb4_unicode_ci,

    signature      BLOB,
    credential     BLOB,
    signed         DATETIME,

    redirect_url   VARCHAR(250)                                                 NOT NULL,
    oauth2_context VARCHAR(1024)                                                NOT NULL DEFAULT '',
    status         ENUM ('pending', 'viewed', 'signed', 'collected','rejected') NOT NULL DEFAULT 'pending',
    secret         VARCHAR(36),
    CONSTRAINT FOREIGN KEY challenges_app_id (app_id) REFERENCES applications (id) ON DELETE SET NULL
);

CREATE OR REPLACE TABLE challenge_codes
(
    code         VARCHAR(36) PRIMARY KEY,
    challenge_id VARCHAR(36),
    CONSTRAINT FOREIGN KEY challenge_codes_challenge_id (challenge_id) REFERENCES challenges (id) ON DELETE CASCADE
);

CREATE OR REPLACE TABLE challenge_ciba_request_ids
(
    request_id   VARCHAR(36) PRIMARY KEY,
    challenge_id VARCHAR(36),
    CONSTRAINT FOREIGN KEY challenge_ciba_request_ids_challenge_id (challenge_id) REFERENCES challenges (id) ON DELETE CASCADE
);

CREATE OR REPLACE PROCEDURE create_ciba_request_id(IN request_id VARCHAR(36), IN challenge_id VARCHAR(36))
BEGIN
    INSERT challenge_ciba_request_ids VALUE (request_id, challenge_id);
END;

CREATE OR REPLACE PROCEDURE get_challenge_by_ciba_request_id(IN request_id VARCHAR(36))
BEGIN
    SELECT created,
           id,
           type,
           app_id,
           expire,
           public_data,
           private_data,
           signature_text,
           signature_data,
           nonce,
           signature,
           credential,
           signed,
           redirect_url,
           oauth2_context,
           status
    FROM challenge_ciba_request_ids c
             RIGHT JOIN challenges c2 on c.challenge_id = c2.id
    WHERE c.request_id = request_id;
END;

CREATE OR REPLACE PROCEDURE delete_ciba_request(IN request_id VARCHAR(36))
BEGIN
    DELETE FROM challenge_ciba_request_ids WHERE challenge_ciba_request_ids.request_id = request_id;
END;

CREATE OR REPLACE PROCEDURE create_code(IN code VARCHAR(36), IN challenge_id VARCHAR(36))
BEGIN
    INSERT challenge_codes VALUE (code, challenge_id);
END;

CREATE OR REPLACE PROCEDURE get_challenge_by_code(IN code VARCHAR(36))
BEGIN
    SELECT created,
           id,
           type,
           app_id,
           expire,
           public_data,
           private_data,
           signature_text,
           signature_data,
           nonce,
           signature,
           credential,
           signed,
           redirect_url,
           oauth2_context,
           status,
           secret
    FROM challenge_codes AS c
             RIGHT JOIN challenges AS c2 on c.challenge_id = c2.id
    WHERE c.code = code;
END;

CREATE OR REPLACE PROCEDURE delete_code(IN code VARCHAR(36))
BEGIN
    DELETE FROM challenge_codes WHERE challenge_codes.code = code;
END;

CREATE OR REPLACE PROCEDURE create_challenge(IN challenge_id VARCHAR(36), IN type VARCHAR(36), IN app_id VARCHAR(36),
                                             IN expire DATETIME,
                                             IN public_data BLOB,
                                             IN private_data BLOB,
                                             IN signature_text TEXT COLLATE utf8mb4_unicode_ci,
                                             IN signature_data BLOB,
                                             IN nonce VARCHAR(16) COLLATE utf8mb4_unicode_ci,
                                             IN redirect_url VARCHAR(250), secret VARCHAR(36))
BEGIN
    INSERT INTO challenges(id, type, app_id, expire, public_data, private_data, signature_text, signature_data, nonce,
                           redirect_url, secret)
    VALUES (challenge_id, type, app_id, expire, public_data, private_data, signature_text, signature_data, nonce,
            redirect_url, secret);
    SELECT challenge_id;
END;

CREATE OR REPLACE PROCEDURE set_challenge_status(IN challenge_id VARCHAR(36),
                                                 IN status ENUM ('pending', 'viewed', 'signed', 'collected', 'rejected'))
BEGIN
    UPDATE challenges AS c SET c.status = status WHERE c.id = challenge_id;
END;

CREATE OR REPLACE PROCEDURE set_oauth2_context(IN challenge_id VARCHAR(36), IN oauth2_context VARCHAR(1024))
BEGIN
    UPDATE challenges AS c SET c.oauth2_context = oauth2_context WHERE c.id = challenge_id;
END;

CREATE OR REPLACE PROCEDURE get_challenge(IN challenge_id VARCHAR(36))
BEGIN
    SELECT created,
           id,
           type,
           app_id,
           expire,
           public_data,
           private_data,
           signature_text,
           signature_data,
           nonce,
           signature,
           credential,
           signed,
           redirect_url,
           oauth2_context,
           status,
           secret
    FROM challenges
    WHERE id = challenge_id;
END;

CREATE OR REPLACE PROCEDURE sign_challenge(IN challenge_id VARCHAR(36), IN signature BLOB, IN credential BLOB)
BEGIN
    UPDATE challenges
    SET challenges.signature  = signature,
        challenges.credential = credential,
        signed                = current_timestamp()
    WHERE id = challenge_id;
END;

CREATE OR REPLACE PROCEDURE delete_challenge(IN challenge_id VARCHAR(36))
BEGIN
    DELETE FROM challenges WHERE id = challenge_id;
END;


CREATE OR REPLACE TABLE sessions
(
    id               VARCHAR(16) PRIMARY KEY,
    user_id          VARCHAR(36)   NOT NULL,
    app_id           VARCHAR(36)   NOT NULL,
    requested_scopes VARCHAR(1024) NOT NULL DEFAULT '',
    counter          INT UNSIGNED  NOT NULL DEFAULT 0,
    created_at       DATETIME      NOT NULL DEFAULT current_timestamp(),
    expire_at        DATETIME,
    CONSTRAINT FOREIGN KEY sessions_user_id (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT FOREIGN KEY sessions_app_id (app_id) REFERENCES applications (id) ON DELETE CASCADE
);

CREATE OR REPLACE PROCEDURE create_session(IN session_id VARCHAR(16), IN user_id VARCHAR(36), IN app_id VARCHAR(36),
                                           IN requested_scopes VARCHAR(1024),
                                           IN expire_at DATETIME)
BEGIN
    INSERT INTO sessions(id, user_id, app_id, requested_scopes, expire_at)
    VALUES (session_id, user_id, app_id, requested_scopes, expire_at);
END;

CREATE OR REPLACE PROCEDURE get_session(IN session_id VARCHAR(18))
BEGIN
    SELECT id, user_id, app_id, requested_scopes, counter, created_at, expire_at
    FROM sessions
    WHERE sessions.id = session_id
      AND (sessions.expire_at > current_timestamp() OR sessions.expire_at IS NULL);
END;

CREATE OR REPLACE PROCEDURE delete_session(IN session_id VARCHAR(18))
BEGIN
    DELETE FROM sessions WHERE sessions.id = session_id;
END;

CREATE OR REPLACE PROCEDURE rotate_session(IN session_id VARCHAR(18), IN expire DATETIME)
BEGIN
    UPDATE sessions
    SET sessions.expire_at = expire,
        sessions.counter   = sessions.counter + 1
    WHERE sessions.id = session_id;
END;

CREATE OR REPLACE PROCEDURE get_sessions_for_user(IN user_id VARCHAR(36))
BEGIN
    SELECT id, user_id, app_id, requested_scopes, counter, created_at, expire_at
    FROM sessions
    WHERE sessions.user_id = user_id
      AND (sessions.expire_at > current_timestamp() OR sessions.expire_at IS NULL);
END;

CREATE OR REPLACE PROCEDURE list_sessions_for_user(IN user_id VARCHAR(36))
BEGIN
    SELECT id, user_id, app_id, requested_scopes, counter, created_at, expire_at
    FROM sessions
    WHERE sessions.user_id = user_id
      AND (sessions.expire_at > current_timestamp() OR sessions.expire_at IS NULL);
END;