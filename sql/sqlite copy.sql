-- INIT
CREATE TABLE IF NOT EXISTS config (
    domain VARCHAR(128) NOT NULL, 
    name VARCHAR(64) NOT NULL, 
    s VARCHAR(64) NOT NULL,
    i INTEGER NOT NULL,
    b TEXT,
    CONSTRAINT pk_domain_key PRIMARY KEY(domain,name)
);

-- GET_CONFIG
SELECT s, i, b FROM config WHERE domain=:domain AND name=:name

-- SET_CONFIG
INSERT INTO config(domain,name,s,i,b) VALUES(:domain,:name,:s,:i,:b)
	ON CONFLICT(domain,name) DO UPDATE SET s=:s,i=:i,b=:b
	WHERE domain=:domain AND name=:name

-- INIT
CREATE TABLE IF NOT EXISTS heads (
    safe VARCHAR(128) NOT NULL, 
    id INTEGER NOT NULL,
    name VARCHAR(8192) NOT NULL, 
    modtime INTEGER,
    hash VARCHAR(128) NOT NULL, 
    CONSTRAINT pk_heads PRIMARY KEY(safe,id)
)

-- INIT
CREATE INDEX IF NOT EXISTS idx_heads_id ON heads(id);

-- GET_HEADS
SELECT id, name, modTime, size, hash FROM heads WHERE safe=:safe AND id > :after ORDER BY id DESC LIMIT :limit

-- SET_HEAD
INSERT INTO heads(safe,id,name,modtime,hash) VALUES(:domain,:name,:firstId,:lastId,:author,:alt,:hash,:modtime,:state)

-- INIT
CREATE TABLE IF NOT EXISTS files (
    domain VARCHAR(128) NOT NULL, 
    name VARCHAR(8192) NOT NULL, 
    id INTEGER NOT NULL,
    firstId INTEGER NOT NULL,
    author VARCHAR(64) NOT NULL, 
    modtime INTEGER,
    state INTEGER,
    hash VARCHAR(128) NOT NULL, 
    hashsplit BLOB,
    CONSTRAINT pk_domain_name_owner PRIMARY KEY(domain,name,author)
)

-- INIT
CREATE INDEX IF NOT EXISTS idx_files_modtime ON files(modtime);

-- GET_FILES
SELECT domain, name, id, firstId, author, modTime, state, hash FROM files " +
	"WHERE domain=:domain

-- GET_FILES_WITH_UPDATES
SELECT domain, name, id, firstId, author, modTime, state, hash FROM files " +
	"WHERE domain=:domain AND state != 0

-- GET_FILES_BY_FIRSTID
SELECT domain, name, id, firstId, author, modTime, state, hash FROM files " +
	"WHERE domain=:domain AND firstId=:firstId

-- GET_FILE_BY_HASH
SELECT domain,name,author,firstId,lastId,alt,modTime,state FROM files " +
	"WHERE hash=:hash

-- GET_FILE_BY_NAME
SELECT domain, name, id, firstId, author, modTime, state, hash FROM files " +
	"WHERE domain=:domain AND name=:name

-- SET_FILE
INSERT INTO files(domain,name,firstId,lastId,author,alt,hash,modtime,state) VALUES(:domain,:name,:firstId,:lastId,:author,:alt,:hash,:modtime,:state)
	ON CONFLICT(domain,name,author) DO UPDATE SET hash=:hash,modtime=:modtime,state=:state
	WHERE domain=:domain AND name=:name AND author=:author

-- INIT
CREATE TABLE IF NOT EXISTS access (
    domain VARCHAR(128) NOT NULL, 
    granted INTEGER,
    config BLOB,
    PRIMARY KEY (domain)
);

-- GET_DOMAINS
SELECT domain FROM access;

-- GET_ACCESS
SELECT config FROM access WHERE domain=:domain

-- SET_ACCESS
INSERT INTO access(domain,granted,config) VALUES(:domain,:granted,:config)
	ON CONFLICT(domain) DO UPDATE SET granted=granted,config=:config
	WHERE domain=:domain

-- INIT
CREATE TABLE IF NOT EXISTS change (
    domain VARCHAR(128) NOT NULL, 
    name VARCHAR(128) NOT NULL, 
    exchange VARCHAR(128) NOT NULL,

    modTime INTEGER,
    PRIMARY KEY (name)
);

-- GET_LAST_CHANGE_NAME 
SELECT name FROM change WHERE domain=:domain AND exchange=:exchange 

-- GET_CHANGE_SINCE
SELECT name, timestamp FROM change WHERE domain=:domain AND name > :base

-- GET_CHANGE
SELECT id_, fatherId, modTime FROM change WHERE domain=:domain AND name = :name

-- ADD_CHANGE
INSERT INTO change(domain, name, timestamp) VALUES(:domain, :name, :timestamp)

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
    CONSTRAINT PRIMARY KEY(signatureKey)
);

-- GET_TRUSTED_ON_SAFE
SELECT signatureKey, encryptionKey,nick FROM identity i INNER JOIN safe_identity t WHERE s.safe=:safe AND i.signatureKey = s.signatureKey AND i.trusted

-- GET_IDENTITY_ON_SAFE
SELECT signatureKey, encryptionKey,nick FROM identity i INNER JOIN safe_identity t WHERE s.safe=:safe AND i.signatureKey = s.signatureKey

-- SET_IDENTITY_ON_SAFE
INSERT INTO safe_identity(signatureKey,encryptionKey,safe) VALUES(:signatureKey,:encryptionKey,:safe)
    ON CONFLICT(signatureKey,encryptionKey) DO NOTHING

-- DET_IDENTITY_ON_SAFE
DELETE FROM safe_identity WHERE signatureKey=:signatureKey AND encryptionKey=:encryptionKey AND safe=:safe

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
UPDATE identity SET trusted=:trusted WHERE domain=:domain AND identity=:identity

-- INSERT_IDENTITY
INSERT INTO identity(signatureKey,encryptionKey,nick) VALUES(:signatureKey,:encryptionKey,:nick)
    ON CONFLICT(signatureKey,encryptionKey) DO NOTHING
