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
		require.Equal(t, 1, len(template.staticParts), "only 1 static part is expected")
		require.Empty(t, template.markerNames)
		require.Equal(t, sample, template.staticParts[0])
	})

	t.Run("OneStaticOneDynamic", func(t *testing.T) {
		sample := "/hello/{world}"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 1, len(template.staticParts), "only 1 static part is expected")
		require.Equal(t, 1, len(template.markerNames), "only 1 marker name is expected")
		require.Equal(t, "/hello/", template.staticParts[0])
		require.Equal(t, "world", template.markerNames[0])
	})

	t.Run("TwoStaticOneDynamic", func(t *testing.T) {
		sample := "/hello/{world}/greet"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 2, len(template.staticParts), "only 2 static parts are expected")
		require.Equal(t, 1, len(template.markerNames), "only 1 marker name is expected")
		require.Equal(t, "/hello/", template.staticParts[0])
		require.Equal(t, "/greet", template.staticParts[1])
		require.Equal(t, "world", template.markerNames[0])
	})

	t.Run("TwoStaticTwoDynamic", func(t *testing.T) {
		sample := "/hello/{world}/greet/{name}"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 2, len(template.staticParts), "only 2 static parts are expected")
		require.Equal(t, 2, len(template.markerNames), "only 2 marker names are expected")
		require.Equal(t, "/hello/", template.staticParts[0])
		require.Equal(t, "/greet/", template.staticParts[1])
		require.Equal(t, "world", template.markerNames[0])
		require.Equal(t, "name", template.markerNames[1])
	})

	t.Run("StaticPrefixInsideOfPart", func(t *testing.T) {
		sample := "/hello/name-of-{world}"
		template, err := Parse(sample)
		require.NoError(t, err)
		require.Equal(t, 1, len(template.staticParts), "only 1 static part is expected")
		require.Equal(t, 1, len(template.markerNames), "only 1 marker name is expected")
		require.Equal(t, "/hello/name-of-", template.staticParts[0])
		require.Equal(t, "world", template.markerNames[0])
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
		require.EqualError(t, err, ErrMustEndWithSlash.Error())
	})
}
