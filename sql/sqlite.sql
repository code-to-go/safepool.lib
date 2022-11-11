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
    safe VARCHAR(128) NOT NULL, 
    k VARCHAR(64) NOT NULL, 
    s VARCHAR(64) NOT NULL,
    i INTEGER NOT NULL,
    b TEXT,
    CONSTRAINT pk_safe_key PRIMARY KEY(safe,k)
);

-- GET_CONFIG
SELECT s, i, b FROM config WHERE safe=:safe AND k=:key

-- SET_CONFIG
INSERT INTO config(safe,k,s,i,b) VALUES(:safe,:key,:s,:i,:b)
	ON CONFLICT(safe,k) DO UPDATE SET s=:s,i=:i,b=:b
	WHERE safe=:safe AND k=:key

-- INIT
CREATE TABLE IF NOT EXISTS heads (
    safe VARCHAR(128) NOT NULL, 
    id INTEGER NOT NULL,
    name VARCHAR(8192) NOT NULL, 
    modtime INTEGER,
    size INTEGER,
    hash VARCHAR(128) NOT NULL, 
    CONSTRAINT pk_safe_id PRIMARY KEY(safe,id)
)

-- INIT
CREATE INDEX IF NOT EXISTS idx_heads_id ON heads(id);

-- GET_HEADS
SELECT id, name, modtime, size, hash,ts FROM heads WHERE safe=:safe AND id > :after AND ts> :afterTime ORDER BY id

-- SET_HEAD
INSERT INTO heads(safe,id,name,modtime,size,hash,ts) VALUES(:safe,:id,:name,:modtime,:size,:hash,:ts)

-- INIT
CREATE TABLE IF NOT EXISTS keystore (
    safe VARCHAR(128) NOT NULL, 
    keyId INTEGER, 
    keyValue VARCHAR(128),
    CONSTRAINT pk_safe_keyId PRIMARY KEY(safe,keyId)
);

-- GET_KEYSTORE
SELECT keyId, keyValue FROM keystore WHERE safe=:safe

-- GET_KEY
SELECT keyValue FROM keystore WHERE safe=:safe AND keyId=:keyId

-- SET_KEY
INSERT INTO keystore(safe,keyId,keyValue) VALUES(:safe,:keyId,:keyValue)
    ON CONFLICT(safe,keyId) DO UPDATE SET keyValue=:keyValue
	    WHERE safe=:safe AND keyId=:keyId

-- INIT
CREATE TABLE IF NOT EXISTS safe_config (
    name VARCHAR(128),
    configs BLOB,
    CONSTRAINT PRIMARY KEY(name)
);

-- GET_SAFE
SELECT configs FROM safe_config WHERE name=:name

-- LIST_SAFE
SELECT name FROM safe_config

-- SET_SAFE
INSERT INTO safe_config(name,configs) VALUES(:name,:configs)
    ON CONFLICT(name) DO UPDATE SET configs=:configs
	    WHERE name=:name

-- INIT
CREATE TABLE IF NOT EXISTS safe_identity (
    safe VARCHAR(128),
    signatureKey VARCHAR(128),
    encryptionKey VARCHAR(128),
    since INTEGER,
    CONSTRAINT pk_safe_sig_enc PRIMARY KEY(safe,signatureKey,encryptionKey)
);

-- GET_TRUSTED_ON_SAFE
SELECT i.signatureKey, i.encryptionKey,nick,ts FROM identity i INNER JOIN safe_identity s WHERE s.safe=:safe AND i.signatureKey = s.signatureKey AND i.trusted

-- GET_IDENTITY_ON_SAFE
SELECT i.signatureKey, i.encryptionKey,nick,since,ts FROM identity i INNER JOIN safe_identity s WHERE s.safe=:safe AND i.signatureKey = s.signatureKey

-- SET_IDENTITY_ON_SAFE
INSERT INTO safe_identity(signatureKey,encryptionKey,since,safe) VALUES(:signatureKey,:encryptionKey,:since,:safe)
    ON CONFLICT(signatureKey,encryptionKey,safe) DO NOTHING

-- DEL_IDENTITY_ON_SAFE
DELETE FROM safe_identity WHERE signatureKey=:signatureKey AND encryptionKey=:encryptionKey AND safe=:safe

-- INIT
CREATE TABLE IF NOT EXISTS feed (
    safe VARCHAR(128),
    feedTime INTEGER,
    CONSTRAINT PRIMARY KEY(safe)
);

-- SET_FEED_TIME
INSERT INTO feed(safe,feedTime) VALUES(:safe,:feedTime)
    ON CONFLICT(safe,feedTime) DO UPDATE SET safe=:safe
	    WHERE safe=:safe AND keyId=:keyId

-- GET_FEED_TIME
SELECT feedTime FROM feed WHERE safe=:safe

-- INIT
CREATE TABLE IF NOT EXISTS chat (
    safe VARCHAR(128),
    name string,
    author string,
    message BLOB,
    CONSTRAINT PRIMARY KEY(safe)
);

-- SET_CHAT_MESSAGE
INSERT INTO chat(safe,name,author,messag, ts) VALUES(:safe,:name,:author,:message, :ts)
    ON CONFLICT(safe,name,author) DO UPDATE SET message=:message
	    WHERE safe=:safe AND name=:name AND author=:author

-- GET_CHAT_MESSAGES
SELECT message FROM chat WHERE safe=:safe AND ts > :after ORDER BY id DESC LIMIT :limit

-- GET_CHAT_OFFSET 
SELECT max(ts)