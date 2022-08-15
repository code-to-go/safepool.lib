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
    author VARCHAR(64) NOT NULL, 
    alt char(256),
    hash VARCHAR(128) NOT NULL, 
    modtime INTEGER,
    state INTEGER,
    merkleTree BLOB,
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
INSERT INTO files(domain,name,author,alt,hash,modtime,state) VALUES(:domain,:name,:author,:alt,:hash,:modtime,:state)
	ON CONFLICT(domain,name,author) DO UPDATE SET hash=:hash,modtime=:modtime,state=:state
	WHERE domain=:domain AND name=:name AND author=:author

-- GET_FILE
SELECT alt, hash, modTime, state FROM files " +
	"WHERE domain=:domain AND name=:name AND author=:author

-- GET_FILE_BY_HASH
SELECT domain, name, author, alt, modTime, state FROM files " +
	"WHERE hash=:hash

-- INIT
CREATE TABLE IF NOT EXISTS domain (
    name VARCHAR(128) NOT NULL, 
    config BLOB,
    PRIMARY KEY (name)
);

-- GET_DOMAINS
SELECT name FROM domain;

-- GET_DOMAIN
SELECT config FROM domain WHERE name=:name

-- SET_DOMAIN
INSERT INTO domain(name,config) VALUES(:name,:config)
	ON CONFLICT(name) DO UPDATE SET config=:config
	WHERE name=:name

-- INIT
CREATE TABLE IF NOT EXISTS log (
    domain VARCHAR(128) NOT NULL, 
    name VARCHAR(128) NOT NULL, 
    timestamp INTEGER,
    PRIMARY KEY (name)
);

-- GET_LOG
SELECT name, timestamp FROM log WHERE domain=:domain AND name > base:

-- ADD_LOG
INSERT INTO LOG(domain, name, timestamp) VALUES(:domain, :name, :timestamp)
