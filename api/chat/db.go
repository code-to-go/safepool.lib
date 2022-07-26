package chat

import (
	"encoding/json"
	"time"

	"github.com/code-to-go/safepool.lib/core"
	"github.com/code-to-go/safepool.lib/sql"
)

func sqlSetMessage(pool string, id uint64, author string, m Message, ts time.Time) error {
	message, err := json.Marshal(m)
	if core.IsErr(err, "cannot marshal chat message: %v") {
		return err
	}

	_, err = sql.Exec("SET_CHAT_MESSAGE", sql.Args{"pool": pool, "id": id, "author": author, "message": message, "ts": sql.EncodeTime(ts)})
	core.IsErr(err, "cannot set message %d on db: %v", id)
	return err
}

func sqlGetMessages(pool string, afterId uint64, beforeId uint64, limit int) ([]Message, error) {
	var messages []Message
	rows, err := sql.Query("GET_CHAT_MESSAGES", sql.Args{"pool": pool, "afterId": afterId, "beforeId": beforeId, "limit": limit})
	if err == nil {
		for rows.Next() {
			var data []byte
			var m Message
			err = rows.Scan(&data)
			if !core.IsErr(err, "cannot read message from db: %v", err) {
				err = json.Unmarshal(data, &m)
				if err == nil {
					messages = append(messages, m)
				}
			}
		}
	}

	return messages, err
}

func sqlGetOffset(pool string) time.Time {
	var ts int64
	err := sql.QueryRow("GET_CHAT_OFFSET", sql.Args{"pool": pool}, &ts)
	if err == nil {
		return sql.DecodeTime(ts)
	} else {
		return time.Time{}
	}

}
