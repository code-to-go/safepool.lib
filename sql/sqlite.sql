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
CREATE TABLE IF NOT EXISTS files (
    domain VARCHAR(128) NOT NULL, 
    name VARCHAR(8192) NOT NULL, 
    firstId INTEGER NOT NULL,
    lastId INTEGER NOT NULL,
    author VARCHAR(64) NOT NULL, 
    alt char(256),
    hash VARCHAR(128) NOT NULL, 
    modtime INTEGER,
    state INTEGER,
    hashsplit BLOB,
    CONSTRAINT pk_domain_name_owner PRIMARY KEY(domain,name,author)
)

-- INIT
CREATE INDEX IF NOT EXISTS idx_files_modtime ON files(modtime);

-- GET_FILES
SELECT name, hash, author, modTime, state FROM files " +
	"WHERE domain=:domain

-- GET_FILES_WITH_UPDATES
SELECT name, author, alt, modTime, state FROM files " +
	"WHERE domain=:domain AND state != 0

-- SET_FILE
INSERT INTO files(domain,name,firstId,lastId,author,alt,hash,modtime,state) VALUES(:domain,:name,:firstId,:lastId,:author,:alt,:hash,:modtime,:state)
	ON CONFLICT(domain,name,author) DO UPDATE SET hash=:hash,modtime=:modtime,state=:state
	WHERE domain=:domain AND name=:name AND author=:author

-- GET_FILE
SELECT firstId,lastId,alt, hash, modTime, state FROM files " +
	"WHERE domain=:domain AND name=:name AND author=:author

-- GET_FILE_BY_HASH
SELECT domain,name,author,firstId,lastId,alt,modTime,state FROM files " +
	"WHERE hash=:hash

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
    timestamp INTEGER,
    PRIMARY KEY (name)
);

-- GET_CHANGE
SELECT name, timestamp FROM change WHERE domain=:domain AND name > :base

-- ADD_CHANGE
INSERT INTO change(domain, name, timestamp) VALUES(:domain, :name, :timestamp)

-- INIT
CREATE TABLE IF NOT EXISTS encKey (
    domain VARCHAR(128) NOT NULL, 
    keyId INTEGER, 
    keyValue VARCHAR(128),
    CONSTRAINT pk_domain_keyId PRIMARY KEY(domain,keyId)
);

-- GET_ENC_KEYS_BY_DOMAIN
SELECT keyId, keyValue FROM encKey WHERE domain=:domain ORDER BY keyValue

-- GET_LAST_ENC_KEY_BY_DOMAIN
SELECT keyId, keyValue FROM encKey WHERE domain=:domain ORDER BY keyValue DESC LIMIT 1

-- SET_ENC_KEY
INSERT INTO encKey(domain,keyId,keyValue) VALUES(:domain,:keyId,:keyValue)
    ON CONFLICT(domain,keyId) DO UPDATE SET keyValue=:keyValue
	    WHERE domain=:domain AND keyId=:keyId

-- INIT
CREATE TABLE IF NOT EXISTS user (
    domain VARCHAR(128) NOT NULL, 
    identity VARCHAR(128), 
    nick VARCHAR(128),
    admin INTEGER,
    active INTEGER,
    CONSTRAINT pk_domain_identity PRIMARY KEY(domain,identity)
);

-- GET_USERS_ID_BY_DOMAIN
SELECT identity FROM user WHERE domain=:domain AND active=:active

-- GET_ADMIN_ID_BY_DOMAIN
SELECT identity FROM user WHERE domain=:domain AND admin=TRUE AND active=:active

-- GET_USERS_BY_DOMAIN
SELECT identity,admin,active FROM user WHERE domain=:domain ORDER BY nick

-- SET_USER
INSERT INTO user(domain,identity,admin,active) VALUES(:domain,:identity,:admin,:active)
    ON CONFLICT(domain,identity) DO UPDATE SET admin=:admin,active=:active
	    WHERE domain=:domain AND identity=:identity

