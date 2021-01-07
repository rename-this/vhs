package gcs

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

func TestNewSink(t *testing.T) {
	bucketName := "bucket-111"

	server := fakestorage.NewServer([]fakestorage.Object{{BucketName: bucketName}})
	defer server.Stop()

	cases := []struct {
		desc        string
		bucketName  string
		sessionID   string
		data        string
		errContains string
		newClientFn newClientFn
	}{
		{
			desc:       "success",
			bucketName: bucketName,
			sessionID:  "111",
			data:       "data-111",
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return server.Client(), nil
			},
		},
		{
			desc:        "error on client",
			errContains: "111",
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return nil, errors.New("111")
			},
		},
		{
			desc:        "missing bucket",
			bucketName:  "none",
			errContains: "failed to find bucket",
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return server.Client(), nil
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := func() error {
				ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{
					GCSBucketName: c.bucketName,
				}, nil)
				ctx.SessionID = c.sessionID
				defer ctx.Cancel()

				s, err := newSink(ctx, c.newClientFn)
				if err != nil {
					return err
				}

				_, err = s.Write([]byte(c.data))
				if err != nil {
					return err
				}

				err = s.Close()
				if err != nil {
					return err
				}

				o := server.Client().Bucket(c.bucketName).Object(c.sessionID)
				r, err := o.NewReader(context.Background())
				if err != nil {
					return err
				}

				defer r.Close()

				data, err := ioutil.ReadAll(r)
				if err != nil {
					return err
				}

				assert.Equal(t, string(data), c.data)
				return nil
			}()

			if c.errContains == "" {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}

func TestNewSource(t *testing.T) {
	var (
		bucketName = "bucket-111"
		objectName = "object-111"
		objectData = "111"
	)

	server := fakestorage.NewServer([]fakestorage.Object{
		{
			BucketName: bucketName,
			Name:       objectName,
			Content:    []byte(objectData),
		},
	})
	defer server.Stop()

	cases := []struct {
		desc        string
		bucketName  string
		objectName  string
		errContains string
		newClientFn newClientFn
	}{

		{
			desc:       "success",
			bucketName: bucketName,
			objectName: objectName,
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return server.Client(), nil
			},
		},
		{
			desc:        "error on client",
			errContains: "111",
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return nil, errors.New("111")
			},
		},
		{
			desc:        "missing bucket",
			bucketName:  "none",
			errContains: "failed to find bucket",
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return server.Client(), nil
			},
		},
		{
			desc:        "missing object",
			bucketName:  bucketName,
			objectName:  "none",
			errContains: "object doesn't exist",
			newClientFn: func(_ session.Context) (*storage.Client, error) {
				return server.Client(), nil
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			errs := make(chan error, 10)
			ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{
				GCSBucketName: c.bucketName,
				GCSObjectName: c.objectName,
			}, errs)

			s := newSource(ctx, c.newClientFn)

			go s.Init(ctx)

			if c.errContains != "" {
				time.Sleep(time.Second)
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
				return
			}

			r := <-s.Streams()
			assert.Equal(t, objectName, r.Meta().SourceID)

			data, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			assert.Equal(t, string(data), objectData)
		})
	}
}

func TestNewSinkFail(t *testing.T) {
	ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{}, nil)
	_, err := NewSink(ctx)
	assert.Assert(t, err != nil)
}
