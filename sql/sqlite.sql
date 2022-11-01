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
    name VARCHAR(64) NOT NULL, 
    s VARCHAR(64) NOT NULL,
    i INTEGER NOT NULL,
    b TEXT,
    CONSTRAINT pk_safe_key PRIMARY KEY(safe,name)
);

-- GET_CONFIG
SELECT s, i, b FROM config WHERE safe=:safe AND name=:name

-- SET_CONFIG
INSERT INTO config(safe,name,s,i,b) VALUES(:safe,:name,:s,:i,:b)
	ON CONFLICT(safe,name) DO UPDATE SET s=:s,i=:i,b=:b
	WHERE safe=:safe AND name=:name

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
SELECT id, name, modtime, size, hash FROM heads WHERE safe=:safe AND id > :after ORDER BY id

-- SET_HEAD
INSERT INTO heads(safe,id,name,modtime,size,hash) VALUES(:safe,:id,:name,:modtime,:size,:hash)

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
CREATE TABLE IF NOT EXISTS safe_identity (
    safe VARCHAR(128),
    signatureKey VARCHAR(128),
    encryptionKey VARCHAR(128),
    since INTEGER,
    CONSTRAINT pk_safe_sig_enc PRIMARY KEY(safe,signatureKey,encryptionKey)
);

-- GET_TRUSTED_ON_SAFE
SELECT i.signatureKey, i.encryptionKey,nick FROM identity i INNER JOIN safe_identity s WHERE s.safe=:safe AND i.signatureKey = s.signatureKey AND i.trusted

-- GET_IDENTITY_ON_SAFE
SELECT i.signatureKey, i.encryptionKey,nick,since FROM identity i INNER JOIN safe_identity s WHERE s.safe=:safe AND i.signatureKey = s.signatureKey

-- SET_IDENTITY_ON_SAFE
INSERT INTO safe_identity(signatureKey,encryptionKey,since,safe) VALUES(:signatureKey,:encryptionKey,:since,:safe)
    ON CONFLICT(signatureKey,encryptionKey,safe) DO NOTHING

-- DEL_IDENTITY_ON_SAFE
DELETE FROM safe_identity WHERE signatureKey=:signatureKey AND encryptionKey=:encryptionKey AND safe=:safe