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
    identity VARCHAR(512), 
    nick VARCHAR(128),
    active INTEGER,
    trusted INTEGER,
    CONSTRAINT pk_domain_identity PRIMARY KEY(domain,identity)
);

-- GET_USERS_ID_BY_DOMAIN
SELECT identity FROM user WHERE domain=:domain AND active=:active

-- GET_USERS_BY_NICK
SELECT identity FROM user WHERE domain=:domain AND active=:active AND nick=:nick

-- GET_USERS_BY_DOMAIN
SELECT identity,active FROM user WHERE domain=:domain ORDER BY nick

-- GET_ALL_TRUSTED
SELECT identity FROM user WHERE domain=:domain AND trusted AND active

-- SET_TRUSTED
UPDATE user SET trusted=:trusted WHERE domain=:domain AND identity=:identity

-- SET_USER
INSERT INTO user(domain,identity,nick,active) VALUES(:domain,:identity,:nick,:active)
    ON CONFLICT(domain,identity) DO UPDATE SET active=:active, nick=:nick
	    WHERE domain=:domain AND identity=:identity
