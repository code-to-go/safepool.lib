-- INIT
CREATE TABLE IF NOT EXISTS identity (
    signatureKey VARCHAR(128),
    encryptionKey VARCHAR(128),
    nick VARCHAR(128),
    trusted INTEGER,
    CONSTRAINT pk_identity_sign_enc PRIMARY KEY(signatureKey, encryptionKey)
);

-- INIT
CREATE INDEX IF NOT EXISTS idx_identity_trust ON identity(trusted);

-- GET_IDENTITY
SELECT signatureKey,encryptionKey,nick FROM identity

-- GET_NICK
SELECT nick FROM identity WHERE signatureKey=:signatureKey AND encryptionKey=:encryptionKey

-- GET_TRUSTED
SELECT signatureKey,encryptionKey,nick FROM identity WHERE trusted

-- SET_TRUSTED
UPDATE identity SET trusted=:trusted WHERE signatureKey=:signatureKey AND encryptionKey=:encryptionKey

-- INSERT_IDENTITY
INSERT INTO identity(signatureKey,encryptionKey,nick) VALUES(:signatureKey,:encryptionKey,:nick)
    ON CONFLICT(signatureKey,encryptionKey) DO NOTHING

-- INIT
CREATE TABLE IF NOT EXISTS config (
    pool VARCHAR(128) NOT NULL, 
    k VARCHAR(64) NOT NULL, 
    s VARCHAR(64) NOT NULL,
    i INTEGER NOT NULL,
    b TEXT,
    CONSTRAINT pk_safe_key PRIMARY KEY(pool,k)
);

-- GET_CONFIG
SELECT s, i, b FROM config WHERE pool=:pool AND k=:key

-- SET_CONFIG
INSERT INTO config(pool,k,s,i,b) VALUES(:pool,:key,:s,:i,:b)
	ON CONFLICT(pool,k) DO UPDATE SET s=:s,i=:i,b=:b
	WHERE pool=:pool AND k=:key

-- INIT
CREATE TABLE IF NOT EXISTS heads (
    pool VARCHAR(128) NOT NULL, 
    id INTEGER NOT NULL,
    name VARCHAR(8192) NOT NULL, 
    modtime INTEGER,
    size INTEGER,
    hash VARCHAR(128) NOT NULL, 
    ts INTEGER,
    CONSTRAINT pk_safe_id PRIMARY KEY(pool,id)
)

-- INIT
CREATE INDEX IF NOT EXISTS idx_heads_id ON heads(id);

-- GET_HEADS
SELECT id, name, modtime, size, hash, ts FROM heads WHERE pool=:pool AND id > :after ORDER BY id

-- SET_HEAD
INSERT INTO heads(pool,id,name,modtime,size,hash,ts) VALUES(:pool,:id,:name,:modtime,:size,:hash,:ts)

-- INIT
CREATE TABLE IF NOT EXISTS keystore (
    pool VARCHAR(128) NOT NULL, 
    keyId INTEGER, 
    keyValue VARCHAR(128),
    CONSTRAINT pk_safe_keyId PRIMARY KEY(pool,keyId)
);

-- GET_KEYSTORE
SELECT keyId, keyValue FROM keystore WHERE pool=:pool

-- GET_KEY
SELECT keyValue FROM keystore WHERE pool=:pool AND keyId=:keyId

-- SET_KEY
INSERT INTO keystore(pool,keyId,keyValue) VALUES(:pool,:keyId,:keyValue)
    ON CONFLICT(pool,keyId) DO UPDATE SET keyValue=:keyValue
	    WHERE pool=:pool AND keyId=:keyId

-- INIT
CREATE TABLE IF NOT EXISTS safe_config (
    name VARCHAR(128),
    configs BLOB,
    PRIMARY KEY(name)
);

-- GET_POOL
SELECT configs FROM safe_config WHERE name=:name

-- LIST_POOL
SELECT name FROM safe_config

-- SET_POOL
INSERT INTO safe_config(name,configs) VALUES(:name,:configs)
    ON CONFLICT(name) DO UPDATE SET configs=:configs
	    WHERE name=:name

-- INIT
CREATE TABLE IF NOT EXISTS safe_identity (
    pool VARCHAR(128),
    signatureKey VARCHAR(128),
    encryptionKey VARCHAR(128),
    since INTEGER,
    ts INTEGER,
    CONSTRAINT pk_safe_sig_enc PRIMARY KEY(pool,signatureKey,encryptionKey)
);

-- GET_TRUSTED_ON_POOL
SELECT i.signatureKey, i.encryptionKey,nick,ts FROM identity i INNER JOIN safe_identity s WHERE s.pool=:pool AND i.signatureKey = s.signatureKey AND i.trusted

-- GET_IDENTITY_ON_POOL
SELECT i.signatureKey, i.encryptionKey,nick,since,ts FROM identity i INNER JOIN safe_identity s WHERE s.pool=:pool AND i.signatureKey = s.signatureKey

-- SET_IDENTITY_ON_POOL
INSERT INTO safe_identity(signatureKey,encryptionKey,since,pool) VALUES(:signatureKey,:encryptionKey,:since,:pool)
    ON CONFLICT(signatureKey,encryptionKey,pool) DO NOTHING

-- DEL_IDENTITY_ON_POOL
DELETE FROM safe_identity WHERE signatureKey=:signatureKey AND encryptionKey=:encryptionKey AND pool=:pool

-- INIT
CREATE TABLE IF NOT EXISTS feed (
    pool VARCHAR(128),
    feedTime INTEGER,
    CONSTRAINT PRIMARY KEY(pool)
);

-- SET_FEED_TIME
INSERT INTO feed(pool,feedTime) VALUES(:pool,:feedTime)
    ON CONFLICT(pool,feedTime) DO UPDATE SET pool=:pool
	    WHERE pool=:pool AND keyId=:keyId

-- GET_FEED_TIME
SELECT feedTime FROM feed WHERE pool=:pool

-- INIT
CREATE TABLE IF NOT EXISTS chat (
    pool VARCHAR(128),
    name string,
    author string,
    message BLOB,
    CONSTRAINT PRIMARY KEY(pool)
);

-- SET_CHAT_MESSAGE
INSERT INTO chat(pool,name,author,messag, ts) VALUES(:pool,:name,:author,:message, :ts)
    ON CONFLICT(pool,name,author) DO UPDATE SET message=:message
	    WHERE pool=:pool AND name=:name AND author=:author

-- GET_CHAT_MESSAGES
SELECT message FROM chat WHERE pool=:pool AND ts > :after ORDER BY id DESC LIMIT :limit

-- GET_CHAT_OFFSET 
SELECT max(ts)