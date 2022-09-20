package main

import (
	"context"
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
