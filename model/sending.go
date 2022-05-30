package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
	"os"
	"strings"
	"time"
)

// Sending keeps sending data.
type (
	Filter struct {
		Tags  []string `json:"tags,omitempty"`
		Codes []int    `json:"codes,omitempty"`
	}
	Sending struct {
		ID      uuid.UUID `json:"id" yaml:"id"`
		StartAt time.Time `json:"start_at,omitempty" yaml:"start_at"`
		Text    string    `json:"text" yaml:"text"`
		//Filter  string    `json:"filter,omitempty" yaml:"filter"` // TODO: invent something instead of string
		Filter Filter    `json:"filter,omitempty" yaml:"filter"`
		StopAt time.Time `json:"stop_at,omitempty" yaml:"stop_at"`
	}
	Sendings []Sending
)

// do not need here, just for example
func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

// do not need here, just for example
func (f *Filter) String() string {
	// ('{"vip1","vip2"}','{911, 912, 913}')
	//Tags := []string{"vip1", "vip2", "vip3"}
	//Codes := []int{911, 912}
	t := strings.Join(f.Tags, "\",\"")
	c := arrayToString(f.Codes, ",")
	return fmt.Sprintf("'{\"%v\"}','{%v}'", t, c)
}

func (dst *Filter) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		return errors.New("NULL values can't be decoded. Scan into a &*MyType to handle NULLs")
	}

	if err := (pgtype.CompositeFields{&dst.Tags, &dst.Codes}).DecodeBinary(ci, src); err != nil {
		return err
	}

	return nil
}

func (src Filter) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) (newBuf []byte, err error) {
	var a pgtype.TextArray
	err = a.Set(src.Tags)
	if err != nil {
		return nil, err
	}

	var b pgtype.Int8Array
	err = b.Set(src.Codes)
	if err != nil {
		return nil, err
	}

	return (pgtype.CompositeFields{&a, &b}).EncodeBinary(ci, buf)
}

func (*Sending) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (Sendings) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Sending) Bind(r *http.Request) error {
	//if s.ID == uuid.Nil {
	//	s.ID, _ = uuid.NewUUID()
	//}

	if s.Text == "" {
		return fmt.Errorf("text is a required field")
	}

	if s.StartAt.IsZero() {
		return fmt.Errorf("start_at is a required field")
	}
	if s.StopAt.IsZero() {
		return fmt.Errorf("stop_at is a required field")
	}
	return nil
}

// GetLoggerContext enriches logger context with essential fields.
func (s *Sending) GetLoggerContext(logCtx zerolog.Context) zerolog.Context {
	if s.ID != uuid.Nil {
		logCtx = logCtx.Str(logging.SendingIDKey, s.ID.String())
	}

	return logCtx
}

func qq() {
	qq := Sending{
		ID:      uuid.UUID{},
		StartAt: time.Time{},
		Text:    "",
		Filter: Filter{
			Tags:  []string{"vip1", "vip2"},
			Codes: []int{911, 912},
		},
		StopAt: time.Time{},
	}
	b, err := json.Marshal(qq)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)

	var qqq Sending
	json.Unmarshal(b, &qqq)

}
