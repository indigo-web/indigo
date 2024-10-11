package internal

import (
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMultipart(t *testing.T) {
	t.Run("real-world example", func(t *testing.T) {
		data := "------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; " +
			"name=\"username\"\r\n\r\nAlice\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nCo" +
			"ntent-Disposition: form-data; name=\"profile_pic\"; filename=\"profile.png\"\r\n" +
			"Content-Type: image/png\r\n\r\n[binary file content]\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--\r\n"
		parsed, err := ParseMultipart(nil, []byte(data), "----WebKitFormBoundary7MA4YWxkTrZu0gW")
		require.Equal(t, 2, len(parsed))
		require.Equal(t, form.Data{
			Name:     "username",
			Filename: "",
			Type:     mime.Plain,
			Charset:  "utf8",
			Value:    "Alice",
		}, parsed[0])
		require.Equal(t, form.Data{
			Name:     "profile_pic",
			Filename: "profile.png",
			Type:     mime.PNG,
			Charset:  "utf8",
			Value:    "[binary file content]",
		}, parsed[1])
		require.NoError(t, err)
	})

	t.Run("prelude and postlude", func(t *testing.T) {
		data := "Hello, world!--boundary\r\nContent-Disposition: form-data; " +
			"name=username\r\n\r\nAlice\r\n--boundary--\r\nAre you still reading?"
		parsed, err := ParseMultipart(nil, []byte(data), "boundary")
		require.NoError(t, err)
		require.Equal(t, 1, len(parsed))
		require.Equal(t, form.Data{
			Name:    "username",
			Type:    mime.Plain,
			Charset: "utf8",
			Value:   "Alice",
		}, parsed[0])
	})

	t.Run("set global coding", func(t *testing.T) {
		data := "--boundary\r\nContent-Disposition: form-data; " +
			"name=_charset_\r\n\r\ncp1252\r\n--boundary\r\nContent-Disposition: " +
			"form-data; name=username\r\n\r\nAlice\r\n--boundary--\r\n"
		parsed, err := ParseMultipart(nil, []byte(data), "boundary")
		require.NoError(t, err)
		require.Equal(t, 1, len(parsed))
		require.Equal(t, form.Data{
			Name:    "username",
			Type:    mime.Plain,
			Charset: "cp1252",
			Value:   "Alice",
		}, parsed[0])
	})

	t.Run("set charset via Content-Type", func(t *testing.T) {
		data := "--boundary\r\n" +
			"Content-Disposition: form-data; name=username\r\n" +
			"Content-Type: application/octet-stream; charset=cp1252\r\n" +
			"\r\nAlice\r\n--boundary--\r\n"
		parsed, err := ParseMultipart(nil, []byte(data), "boundary")
		require.NoError(t, err)
		require.Equal(t, 1, len(parsed))
		require.Equal(t, form.Data{
			Name:    "username",
			Type:    mime.OctetStream,
			Charset: "cp1252",
			Value:   "Alice",
		}, parsed[0])
	})
}

func TestMultipartNegative(t *testing.T) {
	for i, tc := range []string{
		"--boundary\r\n\r\nAlice\r\n--boundary--\r\n",
		"--boundary\r\nContent-Disposition: form?\r\n\r\nAlice\r\n--boundary--\r\n",
		"--boundary\r\nContent-Disposition:\r\n\r\nAlice\r\n--boundary--\r\n",
		"--boundary\r\nContent-Disposition\r\n\r\nAlice\r\n--boundary--\r\n",
		"--boundary\r\nContent-Disposition: form-data; name=\r\n\r\nAlice\r\n--boundary--\r\n",
		"--boundary\r\nContent-Disposition: form-data;\r\n\r\nAlice\r\n--boundary--\r\n",
		"--boundary\r\nContent-Disposition: form-data; name=_charset_\r\n\r\n\r\n--boundary\r\nContent-Disposition: " +
			"form-data; name=username\r\n\r\nAlice\r\n--boundary--\r\n",
		"",
		"prelude only",
		"--boundary\r\nContent-Disposition: form-data; name=\r\n\r\nAlice--boundary--\r\n",
	} {
		_, err := ParseMultipart(nil, []byte(tc), "boundary")
		require.EqualErrorf(t, err, status.ErrBadRequest.Error(),
			"Test case %d: wanted status.ErrBadRequest, got instead %s", i+1, err)
	}
}
