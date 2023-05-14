package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Logs []Log

type Log struct {
	Timestamp  int
	SessionID  string
	ScreenName string
}

func (l Log) String() string {
	return fmt.Sprintf("%d, %s,%s\n", l.Timestamp, l.SessionID, l.ScreenName)
}

func ReadFromCSV(file *os.File) (Logs, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3
	reader.Comment = '#'
	reader.Comma = ';'

	csvLines, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Error reading csv record from file: %s", err)
	}

	logs := []Log{}
	for _, line := range csvLines {

		timestamp, err := strconv.ParseFloat(line[0], 32)
		if err != nil {
			return nil, fmt.Errorf("Error parsing time: %s", err)
		}
		screenName := strings.Replace(line[2], "/ca/ca", "", -1)
		log := Log{
			Timestamp:  int(timestamp),
			SessionID:  strings.TrimSpace(line[1]),
			ScreenName: strings.TrimSpace(screenName),
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (logs Logs) ToSessions() map[string][]Log {

	// Chop all unordered logs into session
	sessions := map[string][]Log{}
	for _, log := range logs {
		session, found := sessions[log.SessionID]
		if !found {
			session = []Log{}
		}
		session = append(session, log)
		sessions[log.SessionID] = session
	}

	for name, session := range sessions {
		// Order each session on time
		sort.Slice(session, func(i, j int) bool {
			return session[i].Timestamp < session[j].Timestamp
		})

		// prepend start-node to each session
		session = append([]Log{
			{
				Timestamp:  session[0].Timestamp - 1,
				SessionID:  name,
				ScreenName: "start",
			},
		}, session...)

		// append end-node to each session
		session = append(session, Log{
			Timestamp:  session[len(session)-1].Timestamp + 1,
			SessionID:  name,
			ScreenName: "end",
		})
		sessions[name] = session
	}

	return sessions
}
