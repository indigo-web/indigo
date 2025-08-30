package http

import (
	"io"
	"testing"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

func TestBody(t *testing.T) {
	newBody := func(data ...string) *Body {
		chunks := make([][]byte, len(data))
		for i, chunk := range data {
			chunks[i] = []byte(chunk)
		}

		return NewBody(dummy.NewMockClient(chunks...))
	}

	t.Run("Callback", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			body := newBody("hello", "world")
			written := make([]string, 0, 2)
			err := body.Callback(func(bytes []byte) error {
				written = append(written, string(bytes))
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, []string{"hello", "world"}, written)
		})

		t.Run("error", func(t *testing.T) {
			body := newBody("hello", "world")
			var written []string
			closure := func(bytes []byte) error {
				written = append(written, string(bytes))
				return status.ErrBadRequest
			}
			err := body.Callback(closure)
			require.EqualError(t, err, status.ErrBadRequest.Error())
			require.Equal(t, []string{"hello"}, written)

			err = body.Callback(closure)
			require.EqualError(t, err, status.ErrBadRequest.Error())
			require.Equal(t, []string{"hello"}, written)
		})
	})

	t.Run("Bytes and String", func(t *testing.T) {
		newRequest := func(cfg *config.Config, contentLength int, chunked bool) *Request {
			return &Request{
				cfg: cfg,
				commonHeaders: commonHeaders{
					ContentLength: contentLength,
					Chunked:       chunked,
				},
			}
		}

		test := func(chunked bool) func(t *testing.T) {
			return func(t *testing.T) {
				testHelloworld := func(t *testing.T, cfg *config.Config) {
					body := newBody("hello", "world")
					body.request = newRequest(cfg, len("hello")+len("world"), chunked)
					data, err := body.String()
					require.NoError(t, err)
					require.Equal(t, "helloworld", data)

					data, err = body.String()
					require.NoError(t, err)
					require.Equal(t, "helloworld", data)
				}

				t.Run("happy path", func(t *testing.T) {
					testHelloworld(t, config.Default())
				})

				t.Run("cap buffer", func(t *testing.T) {
					cfg := config.Default()
					cfg.Body.MaxSize = 5
					testHelloworld(t, cfg)
				})
			}
		}

		t.Run("plain", test(false))
		t.Run("chunked", test(true))
	})

	t.Run("JSON", func(t *testing.T) {
		newRequest := func(mime string) *Request {
			return &Request{
				cfg: config.Default(),
				commonHeaders: commonHeaders{
					ContentType: mime,
				},
			}
		}

		type sampleModel struct {
			Hello string `json:"Hello"`
		}

		const jsonSample = `{"Hello": "world"}`

		parseJSON := func(mime, sample string) (sampleModel, error) {
			body := newBody(sample)
			body.request = newRequest(mime)

			var m sampleModel
			err := body.JSON(&m)
			return m, err
		}

		t.Run("happy path", func(t *testing.T) {
			model, err := parseJSON(mime.JSON, jsonSample)
			require.NoError(t, err)
			require.Equal(t, "world", model.Hello)
		})

		t.Run("incompatible mime", func(t *testing.T) {
			_, err := parseJSON(mime.HTTP, jsonSample)
			require.EqualError(t, err, status.ErrUnsupportedMediaType.Error())
		})
	})

	t.Run("Form", func(t *testing.T) {
		newRequest := func(mime string) *Request {
			return &Request{
				cfg: config.Default(),
				commonHeaders: commonHeaders{
					ContentType: mime,
				},
			}
		}

		t.Run("urlencoded", func(t *testing.T) {
			body := newBody("hello=world")
			body.request = newRequest(mime.FormUrlencoded)
			form, err := body.Form()
			require.NoError(t, err)
			world, found := form.Name("hello")
			require.True(t, found)
			require.Equal(t, "world", world.Value)
		})

		t.Run("multipart", func(t *testing.T) {
			body := newBody("--foo\r\nContent-Disposition: form-data; name=foo\r\n\r\nbar\r\n--foo--\r\n")
			body.request = newRequest(mime.Multipart + "; boundary=foo")
			form, err := body.Form()
			require.NoError(t, err)
			bar, found := form.Name("foo")
			require.True(t, found)
			require.Equal(t, "bar", bar.Value)
		})

		t.Run("incompatible", func(t *testing.T) {
			body := newBody("hello")
			body.request = newRequest(mime.HTTP)
			_, err := body.Form()
			require.EqualError(t, err, status.ErrUnsupportedMediaType.Error())
		})
	})

	t.Run("reader", func(t *testing.T) {
		data := dummy.NewMockClient([]byte("Hello, world!"))
		request := &Request{cfg: config.Default()}
		b := NewBody(data)
		b.Reset(request)

		buff := make([]byte, 12)
		n, err := b.Read(buff)
		require.NoError(t, err)
		require.Equal(t, "Hello, world", string(buff[:n]))

		b.Reset(request)
		n, err = b.Read(buff)
		require.Empty(t, string(buff[:n]))
		require.EqualError(t, err, io.EOF.Error())
	})
}
