package proto

import (
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// docsExamined is renamed from nscannedObjects in 3.2.0
// https://docs.mongodb.com/manual/reference/database-profiler/#system.profile.docsExamined
type SystemProfile struct {
	AllUsers        []interface{} `bson:"allUsers"`
	Client          string        `bson:"client"`
	CursorExhausted bool          `bson:"cursorExhausted"`
	DocsExamined    int           `bson:"docsExamined"`
	NscannedObjects int           `bson:"nscannedObjects"`
	ExecStats       struct {
		Advanced                    int `bson:"advanced"`
		ExecutionTimeMillisEstimate int `bson:"executionTimeMillisEstimate"`
		InputStage                  struct {
			Advanced                    int    `bson:"advanced"`
			Direction                   string `bson:"direction"`
			DocsExamined                int    `bson:"docsExamined"`
			ExecutionTimeMillisEstimate int    `bson:"executionTimeMillisEstimate"`
			Filter                      struct {
				Date struct {
					Eq string `bson:"$eq"`
				} `bson:"date"`
			} `bson:"filter"`
			Invalidates  int    `bson:"invalidates"`
			IsEOF        int    `bson:"isEOF"`
			NReturned    int    `bson:"nReturned"`
			NeedTime     int    `bson:"needTime"`
			NeedYield    int    `bson:"needYield"`
			RestoreState int    `bson:"restoreState"`
			SaveState    int    `bson:"saveState"`
			Stage        string `bson:"stage"`
			Works        int    `bson:"works"`
		} `bson:"inputStage"`
		Invalidates  int    `bson:"invalidates"`
		IsEOF        int    `bson:"isEOF"`
		LimitAmount  int    `bson:"limitAmount"`
		NReturned    int    `bson:"nReturned"`
		NeedTime     int    `bson:"needTime"`
		NeedYield    int    `bson:"needYield"`
		RestoreState int    `bson:"restoreState"`
		SaveState    int    `bson:"saveState"`
		Stage        string `bson:"stage"`
		Works        int    `bson:"works"`
	} `bson:"execStats"`
	KeyUpdates   int `bson:"keyUpdates"`
	KeysExamined int `bson:"keysExamined"`
	Locks        struct {
		Collection struct {
			AcquireCount struct {
				R int `bson:"R"`
			} `bson:"acquireCount"`
		} `bson:"Collection"`
		Database struct {
			AcquireCount struct {
				R int `bson:"r"`
			} `bson:"acquireCount"`
		} `bson:"Database"`
		Global struct {
			AcquireCount struct {
				R int `bson:"r"`
			} `bson:"acquireCount"`
		} `bson:"Global"`
		MMAPV1Journal struct {
			AcquireCount struct {
				R int `bson:"r"`
			} `bson:"acquireCount"`
		} `bson:"MMAPV1Journal"`
	} `bson:"locks"`
	Millis             int       `bson:"millis"`
	Nreturned          int       `bson:"nreturned"`
	Ns                 string    `bson:"ns"`
	NumYield           int       `bson:"numYield"`
	Op                 string    `bson:"op"`
	Protocol           string    `bson:"protocol"`
	Query              BsonD     `bson:"query"`
	UpdateObj          BsonD     `bson:"updateobj"`
	Command            BsonD     `bson:"command"`
	OriginatingCommand BsonD     `bson:"originatingCommand"`
	ResponseLength     int       `bson:"responseLength"`
	Ts                 time.Time `bson:"ts"`
	User               string    `bson:"user"`
	WriteConflicts     int       `bson:"writeConflicts"`
}

func NewExampleQuery(doc SystemProfile) ExampleQuery {
	return ExampleQuery{
		Ns:                 doc.Ns,
		Op:                 doc.Op,
		Query:              doc.Query,
		Command:            doc.Command,
		OriginatingCommand: doc.OriginatingCommand,
		UpdateObj:          doc.UpdateObj,
	}
}

// ExampleQuery is a subset of SystemProfile
type ExampleQuery struct {
	Ns                 string `bson:"ns" json:"ns"`
	Op                 string `bson:"op" json:"op"`
	Query              BsonD  `bson:"query,omitempty" json:"query,omitempty"`
	Command            BsonD  `bson:"command,omitempty" json:"command,omitempty"`
	OriginatingCommand BsonD  `bson:"originatingCommand,omitempty" json:"originatingCommand,omitempty"`
	UpdateObj          BsonD  `bson:"updateobj,omitempty" json:"updateobj,omitempty"`
}

func (self ExampleQuery) Db() string {
	ns := strings.SplitN(self.Ns, ".", 2)
	if len(ns) > 0 {
		return ns[0]
	}
	return ""
}

func (self ExampleQuery) ExplainCmd() bson.D {
	cmd := self.Command

	switch self.Op {
	case "query":
		if cmd.Len() == 0 {
			cmd = self.Query
		}

		// MongoDB 2.6:
		//
		// "query" : {
		//   "query" : {
		//
		//   },
		//	 "$explain" : true
		// },
		if _, ok := cmd.Map()["$explain"]; ok {
			cmd = BsonD{
				{"explain", ""},
			}
			break
		}

		if cmd.Len() == 0 || cmd[0].Name != "find" {
			var filter interface{}
			if cmd.Len() > 0 && cmd[0].Name == "query" {
				filter = cmd[0].Value
			} else {
				filter = cmd
			}

			coll := ""
			s := strings.SplitN(self.Ns, ".", 2)
			if len(s) == 2 {
				coll = s[1]
			}

			cmd = BsonD{
				{"find", coll},
				{"filter", filter},
			}
		} else {
			for i := range cmd {
				// drop $db param as it is not supported in MongoDB 3.0
				if cmd[i].Name == "$db" {
					if len(cmd)-1 == i {
						cmd = cmd[:i]
					} else {
						cmd = append(cmd[:i], cmd[i+1:]...)
					}
					break
				}
			}
		}
	case "update":
		s := strings.SplitN(self.Ns, ".", 2)
		coll := ""
		if len(s) == 2 {
			coll = s[1]
		}
		if cmd.Len() == 0 {
			cmd = BsonD{
				{Name: "q", Value: self.Query},
				{Name: "u", Value: self.UpdateObj},
			}
		}
		cmd = BsonD{
			{Name: "update", Value: coll},
			{Name: "updates", Value: []interface{}{cmd}},
		}
	case "remove":
		s := strings.SplitN(self.Ns, ".", 2)
		coll := ""
		if len(s) == 2 {
			coll = s[1]
		}
		if cmd.Len() == 0 {
			cmd = BsonD{
				{Name: "q", Value: self.Query},
				// we can't determine if limit was 1 or 0 so we assume 0
				{Name: "limit", Value: 0},
			}
		}
		cmd = BsonD{
			{Name: "delete", Value: coll},
			{Name: "deletes", Value: []interface{}{cmd}},
		}
	case "insert":
		if cmd.Len() == 0 {
			cmd = self.Query
		}
		if cmd.Len() == 0 || cmd[0].Name != "insert" {
			coll := ""
			s := strings.SplitN(self.Ns, ".", 2)
			if len(s) == 2 {
				coll = s[1]
			}

			cmd = BsonD{
				{"insert", coll},
			}
		}
	case "getmore":
		if self.OriginatingCommand.Len() > 0 {
			cmd = self.OriginatingCommand
			for i := range cmd {
				// drop $db param as it is not supported in MongoDB 3.0
				if cmd[i].Name == "$db" {
					if len(cmd)-1 == i {
						cmd = cmd[:i]
					} else {
						cmd = append(cmd[:i], cmd[i+1:]...)
					}
					break
				}
			}
		} else {
			cmd = BsonD{
				{Name: "getmore", Value: ""},
			}
		}
	case "command":
		if cmd.Len() == 0 || cmd[0].Name != "group" {
			break
		}

		if group, ok := cmd[0].Value.(BsonD); ok {
			for i := range group {
				// for MongoDB <= 3.2
				// "$reduce" : function () {}
				// It is then Unmarshaled as empty value, so in essence not working
				//
				// for MongoDB >= 3.4
				// "$reduce" : {
				//    "code" : "function () {}"
				// }
				// It is then properly Unmarshaled but then explain fails with "not code"
				//
				// The $reduce function shouldn't affect explain execution plan (e.g. what indexes are picked)
				// so we ignore it for now until we find better way to handle this issue
				if group[i].Name == "$reduce" {
					group[i].Value = "{}"
					cmd[0].Value = group
					break
				}
			}
		}
	}

	return bson.D{
		{
			Name:  "explain",
			Value: cmd,
		},
	}
}
