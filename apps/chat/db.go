package chat

import (
	"encoding/json"
	"time"
	"weshare/core"
	"weshare/sql"
)

func sqlSetMessage(safe string, id uint64, author []byte, m Message, ts time.Time) error {
	_, err := sql.Exec("SET_CHAT_MESSAGE", sql.Args{"ïd": id, "author": author, "ts": sql.EncodeTime(ts)})
	core.IsErr(err, "cannot set message %d on db: %v", id)
	return err
}

func sqlGetMessages(safe string, beforeId uint64, limit int) []Message {
	var messages []Message
	rows, err := sql.Query("GET_CHAT_MESSAGES", sql.Args{"safe": safe, "beforeId": beforeId, "limit": limit})
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

	return messages
}

func sqlGetOffset(safe string) time.Time {
	var ts int64
	err := sql.QueryRow("GET_CHAT_OFFSET", sql.Args{"safe": safe}, &ts)
	if err == nil {
		return sql.DecodeTime(ts)
	} else {
		return time.Time{}
	}

}
