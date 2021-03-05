package coretest

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	"github.com/rename-this/vhs/core"
)

// This file contains testing types that can be used throughout other
// packages and by external plugin authors.

// TestSink is a test sink of bytes.
type TestSink struct {
	OptCloseErr error
	mu          sync.Mutex
	data        []byte
}

// Init initializes the sink.
func (*TestSink) Init(core.Context) {}

// Write writes data to the sink.
func (s *TestSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, p...)

	return len(p), nil
}

// Data returns the data from the sink.
func (s *TestSink) Data() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

// Close closes the sink.
func (s *TestSink) Close() error {
	return s.OptCloseErr
}

// TestSinkInt is a test sink of ints.
type TestSinkInt struct {
	OptCloseErr error
	mu          sync.Mutex
	data        []int
}

// Init initializes the sink.
func (*TestSinkInt) Init(core.Context) {}

// Write writes data to the sink.
func (s *TestSinkInt) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i, err := strconv.ParseInt(string(p), 10, 64)
	if err != nil {
		return -1, err
	}

	s.data = append(s.data, int(i))

	return len(p), nil
}

// Data returns the data from the sink.
func (s *TestSinkInt) Data() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

// Close closes the sink.
func (s *TestSinkInt) Close() error {
	return s.OptCloseErr
}

// TestDoubleOutputModifier doubles the output stream.
type TestDoubleOutputModifier struct {
	OptCloseErr error
}

// Wrap wraps.
func (m *TestDoubleOutputModifier) Wrap(w core.OutputWriter) (core.OutputWriter, error) {
	tdom := &testDoubleOutputModifier{w: w}
	if m.OptCloseErr == nil {
		return tdom, nil
	}

	return &errWriteCloser{
		Writer: tdom,
		err:    m.OptCloseErr,
	}, nil
}

type testDoubleOutputModifier struct {
	w io.WriteCloser
}

func (o *testDoubleOutputModifier) Write(p []byte) (int, error) {
	return o.w.Write(append(p, p...))
}

func (o *testDoubleOutputModifier) Close() error {
	return o.w.Close()
}

type errWriteCloser struct {
	io.Writer
	err error
}

func (n *errWriteCloser) Close() error {
	return n.err
}

// NewTestSource creates a new test source.
func NewTestSource(core.Context) (core.Source, error) {
	return &TestSource{
		streams: make(chan core.InputReader),
	}, nil
}

// NewTestSourceData creates a new test source prefilled with
// a certain set of data.
func NewTestSourceData(data []core.InputReader) *TestSource {
	return &TestSource{
		streams: make(chan core.InputReader),
		Data:    data,
	}
}

// TestSource is a test source.
type TestSource struct {
	streams chan core.InputReader
	Data    []core.InputReader
}

// Streams returns a channel of source streams.
func (s *TestSource) Streams() <-chan core.InputReader {
	return s.streams
}

// Init initializes the test source.
func (s *TestSource) Init(ctx core.Context) {
	for _, d := range s.Data {
		s.streams <- d
	}
	close(s.streams)
}

// NewTestInputFormat creates a new test input format.
func NewTestInputFormat(core.Context) (core.InputFormat, error) {
	return &testFormat{
		out: make(chan interface{}),
	}, nil
}

type testFormat struct {
	out chan interface{}
}

func (i *testFormat) Init(ctx core.Context, m core.Middleware, s <-chan core.InputReader) {
	for r := range s {
		go func() {
			defer func() {
				if err := r.Close(); err != nil {
					ctx.Errors <- err
				}
			}()

			go func() {
				s := bufio.NewScanner(r)
				for s.Scan() {
					ii, err := strconv.Atoi(strings.TrimSpace(s.Text()))
					if err != nil {
						ctx.Errors <- err
					}
					i.out <- ii
				}
			}()

			<-ctx.StdContext.Done()
		}()
	}
}

func (i *testFormat) Out() <-chan interface{} { return i.out }

// TestErrOutputModifier is a test output modifier
// that returns an error on wrap.
type TestErrOutputModifier struct {
	Err error
}

// Wrap retuns an error.
func (m *TestErrOutputModifier) Wrap(w core.OutputWriter) (core.OutputWriter, error) {
	fmt.Println("returning err")
	return nil, m.Err
}

// TestDoubleInputModifier doubles the input stream.
type TestDoubleInputModifier struct {
	OptCloseErr error
}

// Wrap wraps.
func (m *TestDoubleInputModifier) Wrap(r core.InputReader) (core.InputReader, error) {
	tdim := &testDoubleInputModifier{r: r}
	if m.OptCloseErr == nil {
		return core.EmptyMeta(ioutil.NopCloser(tdim)), nil
	}

	return core.EmptyMeta(&errReadCloser{
		Reader: ioutil.NopCloser(tdim),
		err:    m.OptCloseErr,
	}), nil
}

// TestErrInputModifier is a test input modifier
// that returns an error on wrap.
type TestErrInputModifier struct {
	Err error
}

// Wrap returns an error.
func (m *TestErrInputModifier) Wrap(core.InputReader) (core.InputReader, error) {
	return nil, m.Err
}

type errReadCloser struct {
	io.Reader
	err error
}

func (n *errReadCloser) Close() error { return n.err }

type testDoubleInputModifier struct {
	r core.InputReader
}

func (i *testDoubleInputModifier) Read(p []byte) (int, error) {
	b, err := ioutil.ReadAll(i.r)
	if err != nil {
		return 0, err
	}
	copy(p, append(b, b...))
	return len(b) * 2, io.EOF
}

// NewTestOutputFormatNoErr creates a new test output format
// and returns it as a single value so errors don't have to be tested.
func NewTestOutputFormatNoErr(ctx core.Context) core.OutputFormat {
	o, _ := NewTestOutputFormat(ctx)
	return o
}

// NewTestOutputFormat creates a new test output format.
func NewTestOutputFormat(ctx core.Context) (core.OutputFormat, error) {
	return &testOutputFormat{
		in:       make(chan interface{}),
		complete: make(chan struct{}, 1),
		buffered: ctx.FlowConfig.BufferOutput,
	}, nil
}

type testOutputFormat struct {
	in       chan interface{}
	complete chan struct{}
	buffered bool
}

func (f *testOutputFormat) Init(ctx core.Context, w io.Writer) {
	defer func() {
		f.complete <- struct{}{}
	}()

	if f.buffered {
		f.initBuffered(ctx, w)
	} else {
		f.initUnbuffered(ctx, w)
	}
}

func (f *testOutputFormat) initBuffered(ctx core.Context, w io.Writer) {
	var total int
	for {
		select {
		case n := <-f.in:
			total += n.(int)
		case <-ctx.StdContext.Done():
			w.Write([]byte(fmt.Sprint(total)))
			return
		}
	}
}

func (f *testOutputFormat) initUnbuffered(ctx core.Context, w io.Writer) {
	for {
		select {
		case n := <-f.in:
			i := n.(int)
			w.Write([]byte(fmt.Sprint(i)))
		case <-ctx.StdContext.Done():
			return
		}
	}
}

func (f *testOutputFormat) In() chan<- interface{} {
	return f.in
}

func (f *testOutputFormat) Complete() <-chan struct{} {
	return f.complete
}
