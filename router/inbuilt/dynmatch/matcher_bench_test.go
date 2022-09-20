package main

import (
	"context"
	"regexp"
	"testing"
)

func BenchmarkMatch(b *testing.B) {
	staticSample := "/hello/world/length/does/not/matter"
	staticTemplate, _ := Parse(staticSample)

	shortTemplateSample := "/hello/{world}"
	shortSample := "/hello/some-very-long-world"
	shortTemplate, _ := Parse(shortTemplateSample)

	mediumTemplateSample := "/hello/{world}/very/{good}/{ok}"
	mediumSample := "/hello/world-finally-became/very/good-as-fuck/ok-let-it-be"
	mediumTemplate, _ := Parse(mediumTemplateSample)

	longTemplateSample := "/hello/{world}/very/{good}/{ok}/{wanna}/somestatic/{finally}/good"
	longSample := "/hello/world-finally-became/very/good-as-fuck/ok-let-it-be/i-wanna-/somestatic/finally-matcher-is-here/good"
	longTemplate, _ := Parse(longTemplateSample)

	b.Run("OnlyStatic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			staticTemplate.Match(context.Background(), staticSample)
		}
	})

	b.Run("Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			shortTemplate.Match(context.Background(), shortSample)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mediumTemplate.Match(context.Background(), mediumSample)
		}
	})

	b.Run("Long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			longTemplate.Match(context.Background(), longSample)
		}
	})
}

func BenchmarkRegexp(b *testing.B) {
	staticSample := `\/hello\/world\/length\/does\/not\/matter$`
	static, _ := regexp.Compile(staticSample)

	shortTemplateSample := `\/hello\/\w+$`
	shortSample := "/hello/some-very-long-world"
	short, _ := regexp.Compile(shortTemplateSample)

	mediumTemplateSample := `\/hello\/\w+/very/\w+/\w+$`
	mediumSample := "/hello/world-finally-became/very/good-as-fuck/ok-let-it-be"
	medium, _ := regexp.Compile(mediumTemplateSample)

	longTemplateSample := `\/hello\/\w+\/very\/\w+/\w+\/\w+\/somestatic\/\w+\/good$`
	longSample := "/hello/world-finally-became/very/good-as-fuck/ok-let-it-be/i-wanna-/somestatic/finally-matcher-is-here/good"
	long, _ := regexp.Compile(longTemplateSample)

	b.Run("StaticPositive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			static.MatchString(staticSample)
		}
	})

	b.Run("StaticNegative", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			static.MatchString(staticSample + "ok")
		}
	})

	b.Run("ShortPositive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			short.MatchString(shortSample)
		}
	})

	b.Run("ShortNegative", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			short.MatchString(shortSample + "ok")
		}
	})

	b.Run("MediumPositive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			medium.MatchString(mediumSample)
		}
	})

	b.Run("MediumNegative", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			medium.MatchString(mediumSample + "ok")
		}
	})

	b.Run("LongPositive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			long.MatchString(longSample)
		}
	})

	b.Run("LongNegative", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			long.MatchString(longSample + "ok")
		}
	})
}
