package testhelper

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// NewInputFormat creates a new format.
func NewInputFormat(_ *session.Context) (format.Input, error) {
	return &inputFormat{
		out: make(chan interface{}),
	}, nil
}

type inputFormat struct {
	out chan interface{}
}

func (i *inputFormat) Init(ctx *session.Context, m *middleware.Middleware, r io.ReadCloser) error {
	defer r.Close()

	s := bufio.NewScanner(r)
	for s.Scan() {
		ii, err := strconv.Atoi(strings.TrimSpace(s.Text()))
		if err != nil {
			return err
		}
		i.out <- ii
	}
	return nil
}

func (i *inputFormat) Out() <-chan interface{} { return i.out }
