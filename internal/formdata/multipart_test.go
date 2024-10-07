package formdata

import (
	"fmt"
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
		fmt.Println(parsed)
		require.NoError(t, err)
	})
}
