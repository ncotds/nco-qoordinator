package client

import (
	"reflect"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/google/uuid"
)

const (
	TestManager     = "TEST_E2E_MANAGER"
	TestAgent       = "TEST_E2E_AGENT"
	TestAlertsTable = "alerts.status"
)

func SentenceFactory() string {
	return faker.Sentence()
}

func AlertStatusRecordFactory() AlertStatusRecord {
	var alert AlertStatusRecord
	err := faker.FakeData(&alert, options.WithRandomStringLength(255))
	if err != nil {
		panic(err) // fatal, cannot run tests without fixture
	}
	alert.Manager = TestManager
	alert.Agent = TestAgent
	alert.AlertGroup = uuid.NewString()
	return alert
}

type AlertStatusRecord struct {
	Identifier      string `faker:"uuid_hyphenated" db:"Identifier"`
	Node            string `faker:"domain_name" db:"Node"`
	NodeAlias       string `faker:"ipv4" db:"NodeAlias"`
	Agent           string `faker:"-" db:"Agent"`
	Manager         string `faker:"-" db:"Manager"`
	AlertGroup      string `faker:"-" db:"AlertGroup"`
	AlertKey        string `faker:"uuid_digit" db:"AlertKey"`
	Type            uint8  `faker:"oneof: 1" db:"Type"`
	Severity        uint8  `faker:"oneof: 2, 3, 4, 5" db:"Severity"`
	Summary         string `faker:"sentence" db:"Summary"`
	FirstOccurrence int64  `faker:"unix_time" db:"FirstOccurrence"`
	URL             string `faker:"url" db:"URL"`
	ExtendedAttr    string `faker:"paragraph" db:"ExtendedAttr"`
}

func NewAlertStatusRecordFromMap(params map[string]any) AlertStatusRecord {
	var result = AlertStatusRecord{}
	recordFields := reflect.VisibleFields(reflect.TypeOf(result))
	for _, field := range recordFields {
		if param, ok := params[field.Name]; ok {
			value := reflect.ValueOf(param).Convert(field.Type)
			reflect.ValueOf(&result).Elem().FieldByName(field.Name).Set(value)
		}
	}
	return result
}
