package radix

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		sample := "/hello/world"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 2, len(template.segments), "2 segments are expected")

		require.False(t, template.segments[0].IsDynamic)
		require.Equal(t, "hello", template.segments[0].Payload)

		require.False(t, template.segments[1].IsDynamic)
		require.Equal(t, "world", template.segments[1].Payload)
	})

	t.Run("OneStaticOneDynamic", func(t *testing.T) {
		sample := "/hello/{world}"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 2, len(template.segments), "2 segments are expected")

		require.False(t, template.segments[0].IsDynamic)
		require.Equal(t, "hello", template.segments[0].Payload)

		require.True(t, template.segments[1].IsDynamic)
		require.Equal(t, "world", template.segments[1].Payload)
	})

	t.Run("TwoStaticOneDynamic", func(t *testing.T) {
		sample := "/hello/{world}/greet"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 3, len(template.segments), "3 segments are expected")

		require.False(t, template.segments[0].IsDynamic)
		require.Equal(t, "hello", template.segments[0].Payload)

		require.True(t, template.segments[1].IsDynamic)
		require.Equal(t, "world", template.segments[1].Payload)

		require.False(t, template.segments[2].IsDynamic)
		require.Equal(t, "greet", template.segments[2].Payload)
	})

	t.Run("TwoStaticTwoDynamic", func(t *testing.T) {
		sample := "/hello/{world}/greet/{name}"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 4, len(template.segments), "4 segments are expected")

		require.False(t, template.segments[0].IsDynamic)
		require.Equal(t, "hello", template.segments[0].Payload)

		require.True(t, template.segments[1].IsDynamic)
		require.Equal(t, "world", template.segments[1].Payload)

		require.False(t, template.segments[2].IsDynamic)
		require.Equal(t, "greet", template.segments[2].Payload)

		require.True(t, template.segments[3].IsDynamic)
		require.Equal(t, "name", template.segments[3].Payload)
	})
}

func TestParse_Negative(t *testing.T) {
	t.Run("EmptyPath", func(t *testing.T) {
		sample := ""
		_, err := Parse(sample)
		require.EqualError(t, err, ErrEmptyPath.Error())
	})

	t.Run("NoLeadingSlash", func(t *testing.T) {
		sample := "hello/world"
		_, err := Parse(sample)
		require.EqualError(t, err, ErrNeedLeadingSlash.Error())
	})

	t.Run("SlashInsideOfPartName", func(t *testing.T) {
		sample := "/hello/{world/something else}"
		_, err := Parse(sample)
		require.EqualError(t, err, ErrInvalidPartName.Error())
	})

	t.Run("FBraceInsideOfPartName", func(t *testing.T) {
		sample := "/hello/{world {another name}}"
		_, err := Parse(sample)
		require.EqualError(t, err, ErrInvalidPartName.Error())
	})

	t.Run("FBraceInsideOfPartName", func(t *testing.T) {
		sample := "/hello/{world {another name}}"
		_, err := Parse(sample)
		require.EqualError(t, err, ErrInvalidPartName.Error())
	})

	t.Run("NoSlashAfterDynamicPart", func(t *testing.T) {
		sample := "/hello/{world}name/greet"
		_, err := Parse(sample)
		require.EqualError(t, err, ErrDynamicMustBeWholeSection.Error())
	})

	t.Run("DynamicWithPrefix", func(t *testing.T) {
		sample := "/hello-{world}/greet"
		_, err := Parse(sample)
		require.EqualError(t, err, ErrDynamicMustBeWholeSection.Error())
	})
}
