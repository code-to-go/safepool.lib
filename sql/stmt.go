package sql

const stmtGetConfig = "SELECT s, i FROM config WHERE domain=:domain AND key=:key"
const stmtSetConfig = "INSERT INTO config(domain,key,s,i) VALUES(:domain,:key,:s,:i)" +
	" ON CONFLICT(domain,key) DO UPDATE SET s=:s,i=:i" +
	" WHERE domain=:domain AND key=:key"

const stmtGetFiles = "SELECT name, hash, modTime, status FROM files " +
	"WHERE domain=:domain"

const stmtSetFile = "INSERT IN name, hash, modTime, status FROM files " +
	"WHERE domain=:domain"
