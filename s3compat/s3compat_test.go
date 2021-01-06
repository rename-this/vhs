package s3compat

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/rename-this/vhs/internal/smoke"
	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
)

const (
	accessKey = "minio111"
	secretKey = "minio11111"
)

var endpoint string

func TestMain(m *testing.M) {
	resource, cleanup := smoke.SetupDockertestPool(smoke.Config{
		ReadinessPath: "/minio/health/live",
		ReadinessPort: "9000",
		RunOptions: &dockertest.RunOptions{
			Repository: "minio/minio",
			Tag:        "latest",
			Cmd:        []string{"server", "/data"},
			Env:        []string{"MINIO_ACCESS_KEY=" + accessKey, "MINIO_SECRET_KEY=" + secretKey},
		},
	})

	endpoint = "localhost:" + resource.GetPort("9000/tcp")

	code := m.Run()

	cleanup()

	os.Exit(code)
}

func TestSourceWithSink(t *testing.T) {
	cases := []struct {
		desc                string
		data                string
		optSourceEndpoint   string
		optSinkEndpoint     string
		optSourceBucketName string
		optSinkBucketName   string
		errContains         string
	}{
		{
			desc: "success",
			data: "111",
		},
		{
			desc:              "failed to get source client",
			optSourceEndpoint: "---",
			errContains:       "falied to create S3-compatible source client",
		},
		{
			desc:            "failed to get sink client",
			optSinkEndpoint: "---",
			errContains:     "falied to create S3-compatible sink client",
		},
		{
			desc:                "missing source bucket",
			optSourceBucketName: "notfoundsource",
			data:                "111",
			errContains:         "bucket does not exist",
		},
		{
			desc:              "missing sink bucket",
			optSinkBucketName: "notfoundsink",
			data:              "111",
			errContains:       "bucket does not exist",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				safeDesc   = strings.ReplaceAll(c.desc, " ", "")
				bucketName = "bucket-" + safeDesc
				sessionID  = "session-" + safeDesc
			)
			minioClient, err := minio.New(endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
				Secure: false,
			})
			assert.NilError(t, err)
			err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: "tmp"})
			assert.NilError(t, err)

			err = func() error {
				ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{
					S3CompatEndpoint:   endpoint,
					S3CompatAccessKey:  accessKey,
					S3CompatSecretKey:  secretKey,
					S3CompatSecure:     false,
					S3CompatBucketName: bucketName,
				}, nil)

				if c.optSinkBucketName != "" {
					ctx.FlowConfig.S3CompatBucketName = c.optSinkBucketName
				}
				if c.optSinkEndpoint != "" {
					ctx.FlowConfig.S3CompatEndpoint = c.optSinkEndpoint
				}

				ctx.SessionID = sessionID

				defer ctx.Cancel()

				snk, err := NewSink(ctx)
				if err != nil {
					return err
				}

				_, err = snk.Write([]byte(c.data))
				if err != nil {
					return err
				}

				err = snk.Close()
				if err != nil {
					return err
				}

				ctx.FlowConfig.S3CompatObjectName = sessionID
				if c.optSourceBucketName != "" {
					ctx.FlowConfig.S3CompatBucketName = c.optSourceBucketName
				}
				if c.optSourceEndpoint != "" {
					ctx.FlowConfig.S3CompatEndpoint = c.optSourceEndpoint
				}

				src, err := NewSource(ctx)
				if err != nil {
					return err
				}

				go src.Init(ctx)

				strm := <-src.Streams()

				defer strm.Close()

				assert.Equal(t, sessionID, strm.Meta().SourceID)

				data, err := ioutil.ReadAll(strm)
				if err != nil {
					return err
				}

				assert.Equal(t, string(data), c.data)

				ctx.Cancel()

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
