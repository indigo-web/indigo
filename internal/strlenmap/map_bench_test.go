package strlenmap

import "testing"

func BenchmarkMap(b *testing.B) {
	// skipping benchmarks of inserting. It's known that it's slow and allocation-rich, however
	// it isn't considered as important, because the map is usually filled just once on initialization
	// and after that used only for retrieving values

	b.Run("single value per bucket", func(b *testing.B) {
		m := New[string]()
		// this one actually serves no purpose, however used to not let the "hello" entry
		// feel itself lonely
		m.Insert("hi", "anything")
		m.Insert("hello", "everything")
		var key = "hello"

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Get(key)
		}
	})

	b.Run("multiple in a same bucket", func(b *testing.B) {
		m := New[string]()
		// this one actually serves no purpose, however used to not let the "hello" entry
		// feel itself lonely
		m.Insert("hello", "everything")
		m.Insert("hallo", "anything")
		m.Insert("hola!", "this one is in spanish")
		var key = "hola!"

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Get(key)
		}
	})
}
