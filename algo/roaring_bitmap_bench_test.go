package algo

import (
	"fmt"
	"math/rand"
	"testing"
)

// BenchmarkRoaringBitmap_Add benchmarks adding values to the bitmap
func BenchmarkRoaringBitmap_Add(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		rb := NewRoaringBitmap()
		i := uint32(0)

		for b.Loop() {
			rb.Add(i)
			i++
		}
	})

	b.Run("Sparse", func(b *testing.B) {
		rb := NewRoaringBitmap()
		i := uint32(0)

		for b.Loop() {
			rb.Add(i * 1000) // Sparse values
			i++
		}
	})

	b.Run("Dense", func(b *testing.B) {
		rb := NewRoaringBitmap()
		i := uint32(0)

		for b.Loop() {
			rb.Add(i % 10000) // Dense within a range
			i++
		}
	})
}

// BenchmarkRoaringBitmap_Contains benchmarks checking if values exist
func BenchmarkRoaringBitmap_Contains(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup: Add values (excluded from timing with b.Loop())
			rb := NewRoaringBitmap()
			for i := 0; i < size; i++ {
				rb.Add(uint32(i))
			}

			b.ReportAllocs()
			i := 0

			for b.Loop() {
				rb.Contains(uint32(i % size))
				i++
			}
		})
	}
}

// BenchmarkRoaringBitmap_Remove benchmarks removing values from the bitmap
func BenchmarkRoaringBitmap_Remove(b *testing.B) {
	b.Run("FromArrayContainer", func(b *testing.B) {
		// Setup: Add sparse values (array container)
		rb := NewRoaringBitmap()
		for i := 0; i < 1000; i++ {
			rb.Add(uint32(i * 100))
		}

		i := 0
		for b.Loop() {
			value := uint32((i % 1000) * 100)
			rb.Remove(value)
			// Re-add to maintain consistent state
			rb.Add(value)
			i++
		}
	})

	b.Run("FromBitmapContainer", func(b *testing.B) {
		// Setup: Add dense values (bitmap container)
		rb := NewRoaringBitmap()
		for i := 0; i < arrayToBitmapThreshold+1000; i++ {
			rb.Add(uint32(i))
		}

		i := 0
		for b.Loop() {
			value := uint32(i % (arrayToBitmapThreshold + 1000))
			rb.Remove(value)
			// Re-add to maintain consistent state
			rb.Add(value)
			i++
		}
	})
}

// BenchmarkRoaringBitmap_Union benchmarks union operations
func BenchmarkRoaringBitmap_Union(b *testing.B) {
	b.Run("ArrayContainers", func(b *testing.B) {
		// Setup: Create bitmaps with sparse data
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		for i := 0; i < 1000; i += 2 {
			rb1.Add(uint32(i))
		}
		for i := 1; i < 1000; i += 2 {
			rb2.Add(uint32(i))
		}

		b.ReportAllocs()

		for b.Loop() {
			_ = rb1.Union(rb2)
		}
	})

	b.Run("BitmapContainers", func(b *testing.B) {
		// Setup: Create bitmaps with dense data
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		for i := 0; i < arrayToBitmapThreshold+1000; i += 2 {
			rb1.Add(uint32(i))
		}
		for i := 1; i < arrayToBitmapThreshold+1000; i += 2 {
			rb2.Add(uint32(i))
		}

		b.ReportAllocs()

		for b.Loop() {
			_ = rb1.Union(rb2)
		}
	})

	b.Run("MixedContainers", func(b *testing.B) {
		// Setup: rb1 with sparse data, rb2 with dense data
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		for i := 0; i < 1000; i++ {
			rb1.Add(uint32(i * 100))
		}
		for i := 0; i < arrayToBitmapThreshold+1000; i++ {
			rb2.Add(uint32(i))
		}

		b.ReportAllocs()

		for b.Loop() {
			_ = rb1.Union(rb2)
		}
	})
}

// BenchmarkRoaringBitmap_Intersection benchmarks intersection operations
func BenchmarkRoaringBitmap_Intersection(b *testing.B) {
	b.Run("ArrayContainers", func(b *testing.B) {
		// Setup: Create overlapping sparse data
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		for i := 0; i < 2000; i++ {
			rb1.Add(uint32(i))
		}
		for i := 1000; i < 3000; i++ {
			rb2.Add(uint32(i))
		}

		b.ReportAllocs()

		for b.Loop() {
			_ = rb1.Intersection(rb2)
		}
	})

	b.Run("BitmapContainers", func(b *testing.B) {
		// Setup: Create overlapping dense data
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		for i := 0; i < arrayToBitmapThreshold+2000; i++ {
			rb1.Add(uint32(i))
		}
		for i := 1000; i < arrayToBitmapThreshold+3000; i++ {
			rb2.Add(uint32(i))
		}

		b.ReportAllocs()

		for b.Loop() {
			_ = rb1.Intersection(rb2)
		}
	})
}

// BenchmarkRoaringBitmap_Difference benchmarks difference operations
func BenchmarkRoaringBitmap_Difference(b *testing.B) {
	// Setup
	rb1 := NewRoaringBitmap()
	rb2 := NewRoaringBitmap()
	for i := 0; i < 10000; i++ {
		rb1.Add(uint32(i))
	}
	for i := 2000; i < 8000; i++ {
		rb2.Add(uint32(i))
	}

	b.ReportAllocs()

	for b.Loop() {
		_ = rb1.Difference(rb2)
	}
}

// BenchmarkRoaringBitmap_ToArray benchmarks converting bitmap to array
func BenchmarkRoaringBitmap_ToArray(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup: Add values
			rb := NewRoaringBitmap()
			for i := 0; i < size; i++ {
				rb.Add(uint32(i))
			}

			b.ReportAllocs()

			for b.Loop() {
				_ = rb.ToArray()
			}
		})
	}
}

// BenchmarkRoaringBitmap_Cardinality benchmarks counting elements
func BenchmarkRoaringBitmap_Cardinality(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup: Add values
			rb := NewRoaringBitmap()
			for i := 0; i < size; i++ {
				rb.Add(uint32(i))
			}

			for b.Loop() {
				_ = rb.Cardinality()
			}
		})
	}
}

// BenchmarkRoaringBitmap_ContainerConversion benchmarks container type conversions
func BenchmarkRoaringBitmap_ContainerConversion(b *testing.B) {
	b.Run("ArrayToBitmap", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			rb := NewRoaringBitmap()

			// Add values up to threshold
			for j := 0; j < arrayToBitmapThreshold; j++ {
				rb.Add(uint32(j))
			}
			b.StartTimer()

			// Trigger conversion by adding one more
			rb.Add(uint32(arrayToBitmapThreshold))
		}
	})

	b.Run("BitmapToArray", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			rb := NewRoaringBitmap()

			// Create bitmap container
			for j := 0; j < arrayToBitmapThreshold+1000; j++ {
				rb.Add(uint32(j))
			}

			// Remove most values to near threshold
			for j := bitmapToArrayThreshold; j < arrayToBitmapThreshold+1000; j++ {
				rb.Remove(uint32(j))
			}
			b.StartTimer()

			// Trigger conversion by removing one more
			rb.Remove(uint32(bitmapToArrayThreshold - 1))
		}
	})
}

// BenchmarkRoaringBitmap_ParallelOperations benchmarks concurrent operations
func BenchmarkRoaringBitmap_ParallelOperations(b *testing.B) {
	b.Run("ParallelAdd", func(b *testing.B) {
		rb := NewRoaringBitmap()

		b.RunParallel(func(pb *testing.PB) {
			i := uint32(0)
			for pb.Next() {
				rb.Add(i)
				i++
			}
		})
	})

	b.Run("ParallelContains", func(b *testing.B) {
		// Setup: Pre-populate bitmap
		rb := NewRoaringBitmap()
		for i := 0; i < 100000; i++ {
			rb.Add(uint32(i))
		}

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				rb.Contains(uint32(i % 100000))
				i++
			}
		})
	})
}

// BenchmarkRoaringBitmap_MixedOperations benchmarks realistic usage patterns
func BenchmarkRoaringBitmap_MixedOperations(b *testing.B) {
	b.Run("AddContainsRemove", func(b *testing.B) {
		rb := NewRoaringBitmap()
		i := 0

		for b.Loop() {
			value := uint32(i % 10000)

			rb.Add(value)
			_ = rb.Contains(value)
			if i%10 == 0 { // Remove occasionally
				rb.Remove(value)
			}
			i++
		}
	})

	b.Run("BuildAndQuery", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			rb := NewRoaringBitmap()

			// Build phase
			for j := 0; j < 1000; j++ {
				rb.Add(uint32(j * 10))
			}
			b.StartTimer()

			// Query phase
			for j := 0; j < 100; j++ {
				_ = rb.Contains(uint32(j * 100))
			}
		}
	})
}

// BenchmarkRoaringBitmap_SetOperationsChain benchmarks chained set operations
func BenchmarkRoaringBitmap_SetOperationsChain(b *testing.B) {
	// Setup
	rb1 := NewRoaringBitmap()
	rb2 := NewRoaringBitmap()
	rb3 := NewRoaringBitmap()
	for i := 0; i < 5000; i++ {
		rb1.Add(uint32(i))
		rb2.Add(uint32(i + 2500))
		rb3.Add(uint32(i + 5000))
	}

	b.ReportAllocs()

	for b.Loop() {
		// Chain operations: (A ∪ B) ∩ C
		union := rb1.Union(rb2)
		_ = union.Intersection(rb3)
	}
}

// BenchmarkRoaringBitmap_MemoryUsage benchmarks memory allocation patterns
func BenchmarkRoaringBitmap_MemoryUsage(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		rb := NewRoaringBitmap()

		// Add a mix of sparse and dense data
		for j := 0; j < 1000; j++ {
			rb.Add(uint32(j))        // Dense
			rb.Add(uint32(j * 1000)) // Sparse
		}

		// Perform some operations
		_ = rb.Cardinality()
		_ = rb.ToArray()
	}
}

// BenchmarkRoaringBitmapVsHashtable_Add compares addition performance
func BenchmarkRoaringBitmapVsHashtable_Add(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		// Sequential data
		b.Run(fmt.Sprintf("Sequential-Size-%d", size), func(b *testing.B) {
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					rb := NewRoaringBitmap()
					b.StartTimer()

					for j := 0; j < size; j++ {
						rb.Add(uint32(j))
					}
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					ht := make(map[uint32]bool)
					b.StartTimer()

					for j := 0; j < size; j++ {
						ht[uint32(j)] = true
					}
				}
			})
		})

		// Sparse data
		b.Run(fmt.Sprintf("Sparse-Size-%d", size), func(b *testing.B) {
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					rb := NewRoaringBitmap()
					b.StartTimer()

					for j := 0; j < size; j++ {
						rb.Add(uint32(j * 1000)) // Sparse values
					}
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					ht := make(map[uint32]bool)
					b.StartTimer()

					for j := 0; j < size; j++ {
						ht[uint32(j*1000)] = true
					}
				}
			})
		})

		// Random data
		b.Run(fmt.Sprintf("Random-Size-%d", size), func(b *testing.B) {
			// Pre-generate random values for fair comparison
			r := rand.New(rand.NewSource(42)) // Fixed seed for reproducibility
			randomValues := make([]uint32, size)
			for j := 0; j < size; j++ {
				randomValues[j] = r.Uint32()
			}

			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					rb := NewRoaringBitmap()
					b.StartTimer()

					for _, val := range randomValues {
						rb.Add(val)
					}
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					ht := make(map[uint32]bool)
					b.StartTimer()

					for _, val := range randomValues {
						ht[val] = true
					}
				}
			})
		})
	}
}

// BenchmarkRoaringBitmapVsHashtable_Contains compares lookup performance
func BenchmarkRoaringBitmapVsHashtable_Contains(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		// Sequential data lookup
		b.Run(fmt.Sprintf("Sequential-Size-%d", size), func(b *testing.B) {
			// Pre-generate lookup keys for fair comparison
			lookupKeys := make([]uint32, size)
			for j := 0; j < size; j++ {
				lookupKeys[j] = uint32(j)
			}

			// Setup RoaringBitmap
			rb := NewRoaringBitmap()
			for _, key := range lookupKeys {
				rb.Add(key)
			}

			// Setup Hashtable
			ht := make(map[uint32]bool)
			for _, key := range lookupKeys {
				ht[key] = true
			}

			b.ResetTimer()
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				i := 0
				for b.Loop() {
					rb.Contains(lookupKeys[i%size])
					i++
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				i := 0
				for b.Loop() {
					_ = ht[lookupKeys[i%size]]
					i++
				}
			})
		})

		// Sparse data lookup
		b.Run(fmt.Sprintf("Sparse-Size-%d", size), func(b *testing.B) {
			// Pre-generate lookup keys for fair comparison
			lookupKeys := make([]uint32, size)
			for j := 0; j < size; j++ {
				lookupKeys[j] = uint32(j * 1000)
			}

			// Setup RoaringBitmap
			rb := NewRoaringBitmap()
			for _, key := range lookupKeys {
				rb.Add(key)
			}

			// Setup Hashtable
			ht := make(map[uint32]bool)
			for _, key := range lookupKeys {
				ht[key] = true
			}

			b.ResetTimer()
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				i := 0
				for b.Loop() {
					rb.Contains(lookupKeys[i%size])
					i++
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				i := 0
				for b.Loop() {
					_ = ht[lookupKeys[i%size]]
					i++
				}
			})
		})

		// Random lookups (mix of existing and non-existing)
		b.Run(fmt.Sprintf("RandomLookup-Size-%d", size), func(b *testing.B) {
			// Setup data - generate k random values in range [0, 2^31)
			r := rand.New(rand.NewSource(42))
			values := make([]uint32, size)
			for j := 0; j < size; j++ {
				values[j] = r.Uint32() & 0x7FFFFFFF // Limit to 2^31-1
			}

			// Setup RoaringBitmap
			rb := NewRoaringBitmap()
			for _, val := range values {
				rb.Add(val)
			}

			// Setup Hashtable
			ht := make(map[uint32]bool)
			for _, val := range values {
				ht[val] = true
			}

			// Pre-generate lookup values array - k random values in range [0, 2^31)
			lookupValues := make([]uint32, size)
			for j := 0; j < size; j++ {
				lookupValues[j] = r.Uint32() & 0x7FFFFFFF // Random values in range [0, 2^31)
			}

			b.ResetTimer()
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				i := 0
				for b.Loop() {
					rb.Contains(lookupValues[i%size])
					i++
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				i := 0
				for b.Loop() {
					_ = ht[lookupValues[i%size]]
					i++
				}
			})
		})
	}
}

// BenchmarkRoaringBitmapVsHashtable_Memory compares memory usage
func BenchmarkRoaringBitmapVsHashtable_Memory(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Sequential-Size-%d", size), func(b *testing.B) {
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					rb := NewRoaringBitmap()
					for j := 0; j < size; j++ {
						rb.Add(uint32(j))
					}
					// Keep reference to prevent GC
					_ = rb.Cardinality()
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					ht := make(map[uint32]bool, size) // Pre-allocate capacity
					for j := 0; j < size; j++ {
						ht[uint32(j)] = true
					}
					// Keep reference to prevent GC
					_ = len(ht)
				}
			})
		})

		b.Run(fmt.Sprintf("Sparse-Size-%d", size), func(b *testing.B) {
			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					rb := NewRoaringBitmap()
					for j := 0; j < size; j++ {
						rb.Add(uint32(j * 1000))
					}
					_ = rb.Cardinality()
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					ht := make(map[uint32]bool, size)
					for j := 0; j < size; j++ {
						ht[uint32(j*1000)] = true
					}
					_ = len(ht)
				}
			})
		})
	}
}

// BenchmarkRoaringBitmapVsHashtable_MixedOperations compares realistic usage patterns
func BenchmarkRoaringBitmapVsHashtable_MixedOperations(b *testing.B) {
	b.Run("AddContainsPattern", func(b *testing.B) {
		b.Run("RoaringBitmap", func(b *testing.B) {
			rb := NewRoaringBitmap()
			i := 0

			for b.Loop() {
				value := uint32(i % 100000)

				// Add every 3rd iteration
				if i%3 == 0 {
					rb.Add(value)
				}

				// Check contains every iteration
				_ = rb.Contains(value)
				i++
			}
		})

		b.Run("Hashtable", func(b *testing.B) {
			ht := make(map[uint32]bool)
			i := 0

			for b.Loop() {
				value := uint32(i % 100000)

				// Add every 3rd iteration
				if i%3 == 0 {
					ht[value] = true
				}

				// Check contains every iteration
				_ = ht[value]
				i++
			}
		})
	})

	b.Run("BuildThenQuery", func(b *testing.B) {
		b.Run("RoaringBitmap", func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				rb := NewRoaringBitmap()

				// Build phase
				for j := 0; j < 10000; j++ {
					rb.Add(uint32(j))
				}
				b.StartTimer()

				// Query phase
				for j := 0; j < 1000; j++ {
					_ = rb.Contains(uint32(j * 10))
				}
			}
		})

		b.Run("Hashtable", func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				ht := make(map[uint32]bool)

				// Build phase
				for j := 0; j < 10000; j++ {
					ht[uint32(j)] = true
				}
				b.StartTimer()

				// Query phase
				for j := 0; j < 1000; j++ {
					_ = ht[uint32(j*10)]
				}
			}
		})
	})
}

// BenchmarkRoaringBitmapVsHashtable_SetOperations compares set operations
func BenchmarkRoaringBitmapVsHashtable_SetOperations(b *testing.B) {
	// Setup test data
	size := 10000

	// RoaringBitmap setup
	rb1 := NewRoaringBitmap()
	rb2 := NewRoaringBitmap()
	for i := 0; i < size; i += 2 {
		rb1.Add(uint32(i))
	}
	for i := 1; i < size; i += 2 {
		rb2.Add(uint32(i))
	}

	// Hashtable setup
	ht1 := make(map[uint32]bool)
	ht2 := make(map[uint32]bool)
	for i := 0; i < size; i += 2 {
		ht1[uint32(i)] = true
	}
	for i := 1; i < size; i += 2 {
		ht2[uint32(i)] = true
	}

	b.Run("Union", func(b *testing.B) {
		b.Run("RoaringBitmap", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = rb1.Union(rb2)
			}
		})

		b.Run("Hashtable", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				result := make(map[uint32]bool)
				for k := range ht1 {
					result[k] = true
				}
				for k := range ht2 {
					result[k] = true
				}
				_ = result
			}
		})
	})

	b.Run("Intersection", func(b *testing.B) {
		// Create overlapping sets for intersection
		rb3 := NewRoaringBitmap()
		rb4 := NewRoaringBitmap()
		ht3 := make(map[uint32]bool)
		ht4 := make(map[uint32]bool)

		for i := 0; i < size; i++ {
			rb3.Add(uint32(i))
			ht3[uint32(i)] = true
		}
		for i := size / 2; i < size+size/2; i++ {
			rb4.Add(uint32(i))
			ht4[uint32(i)] = true
		}

		b.Run("RoaringBitmap", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = rb3.Intersection(rb4)
			}
		})

		b.Run("Hashtable", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				result := make(map[uint32]bool)
				for k := range ht3 {
					if ht4[k] {
						result[k] = true
					}
				}
				_ = result
			}
		})
	})
}

// BenchmarkRoaringBitmapVsHashtable_Cardinality compares counting operations
func BenchmarkRoaringBitmapVsHashtable_Cardinality(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			rb := NewRoaringBitmap()
			ht := make(map[uint32]bool)

			for i := 0; i < size; i++ {
				value := uint32(i)
				rb.Add(value)
				ht[value] = true
			}

			b.Run("RoaringBitmap", func(b *testing.B) {
				for b.Loop() {
					_ = rb.Cardinality()
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				for b.Loop() {
					_ = len(ht)
				}
			})
		})
	}
}

// BenchmarkRoaringBitmapVsHashtable_Iteration compares iteration performance
func BenchmarkRoaringBitmapVsHashtable_Iteration(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			rb := NewRoaringBitmap()
			ht := make(map[uint32]bool)

			for i := 0; i < size; i++ {
				value := uint32(i)
				rb.Add(value)
				ht[value] = true
			}

			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					values := rb.ToArray()
					// Process values to prevent optimization
					sum := uint64(0)
					for _, v := range values {
						sum += uint64(v)
					}
					_ = sum
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					// Process values to prevent optimization
					sum := uint64(0)
					for k := range ht {
						sum += uint64(k)
					}
					_ = sum
				}
			})
		})
	}
}

// BenchmarkRoaringBitmapVsHashtable_ParallelContains compares concurrent lookup performance
func BenchmarkRoaringBitmapVsHashtable_ParallelContains(b *testing.B) {
	size := 100000

	// Setup
	rb := NewRoaringBitmap()
	ht := make(map[uint32]bool)

	for i := 0; i < size; i++ {
		value := uint32(i)
		rb.Add(value)
		ht[value] = true
	}

	b.Run("RoaringBitmap", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				rb.Contains(uint32(i % size))
				i++
			}
		})
	})

	b.Run("Hashtable", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_ = ht[uint32(i%size)]
				i++
			}
		})
	})
}

// BenchmarkSetAlgebra_RoaringVsHashtable proves "Roaring beats hashtable for set algebra operations"
// Tests union, intersection, and difference operations with varying set sizes
func BenchmarkSetAlgebra_RoaringVsHashtable(b *testing.B) {
	sizes := []int{10000, 50000, 100000, 500000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Union-Size-%d", size), func(b *testing.B) {
			// Setup: Two overlapping sets, each with 'size' elements
			// Set1: [0, size), Set2: [size/2, size+size/2)
			// Overlap: [size/2, size) - about 50% overlap

			// RoaringBitmap setup
			rb1 := NewRoaringBitmap()
			rb2 := NewRoaringBitmap()
			for i := 0; i < size; i++ {
				rb1.Add(uint32(i))
			}
			for i := size / 2; i < size+size/2; i++ {
				rb2.Add(uint32(i))
			}

			// Hashtable setup
			ht1 := make(map[uint32]bool)
			ht2 := make(map[uint32]bool)
			for i := 0; i < size; i++ {
				ht1[uint32(i)] = true
			}
			for i := size / 2; i < size+size/2; i++ {
				ht2[uint32(i)] = true
			}

			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = rb1.Union(rb2)
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					result := make(map[uint32]bool, len(ht1)+len(ht2))
					for k := range ht1 {
						result[k] = true
					}
					for k := range ht2 {
						result[k] = true
					}
					_ = result
				}
			})
		})

		b.Run(fmt.Sprintf("Intersection-Size-%d", size), func(b *testing.B) {
			// Setup: Two overlapping sets for intersection
			rb1 := NewRoaringBitmap()
			rb2 := NewRoaringBitmap()
			for i := 0; i < size; i++ {
				rb1.Add(uint32(i))
			}
			for i := size / 2; i < size+size/2; i++ {
				rb2.Add(uint32(i))
			}

			ht1 := make(map[uint32]bool)
			ht2 := make(map[uint32]bool)
			for i := 0; i < size; i++ {
				ht1[uint32(i)] = true
			}
			for i := size / 2; i < size+size/2; i++ {
				ht2[uint32(i)] = true
			}

			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = rb1.Intersection(rb2)
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					result := make(map[uint32]bool)
					for k := range ht1 {
						if ht2[k] {
							result[k] = true
						}
					}
					_ = result
				}
			})
		})

		b.Run(fmt.Sprintf("Difference-Size-%d", size), func(b *testing.B) {
			// Setup: Two overlapping sets for difference
			rb1 := NewRoaringBitmap()
			rb2 := NewRoaringBitmap()
			for i := 0; i < size; i++ {
				rb1.Add(uint32(i))
			}
			for i := size / 3; i < 2*size/3; i++ {
				rb2.Add(uint32(i))
			}

			ht1 := make(map[uint32]bool)
			ht2 := make(map[uint32]bool)
			for i := 0; i < size; i++ {
				ht1[uint32(i)] = true
			}
			for i := size / 3; i < 2*size/3; i++ {
				ht2[uint32(i)] = true
			}

			b.Run("RoaringBitmap", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = rb1.Difference(rb2)
				}
			})

			b.Run("Hashtable", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					result := make(map[uint32]bool)
					for k := range ht1 {
						if !ht2[k] {
							result[k] = true
						}
					}
					_ = result
				}
			})
		})
	}
}

// BenchmarkSparseData_RoaringVsHashtable proves "Roaring beats hashtable for sparse data"
// Tests scenarios with millions of possible integers but only thousands present
func BenchmarkSparseData_RoaringVsHashtable(b *testing.B) {
	scenarios := []struct {
		name     string
		universe uint32 // Total possible range
		present  int    // Actually present values
		sparsity string // Description
	}{
		{"VerySparse-1M-Universe-1K-Present", 1000000, 1000, "0.1%"},
		{"Sparse-10M-Universe-5K-Present", 10000000, 5000, "0.05%"},
		{"UltraSparse-100M-Universe-10K-Present", 100000000, 10000, "0.01%"},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			// Generate sparse random data
			r := rand.New(rand.NewSource(42))
			values := make([]uint32, scenario.present)
			valueSet := make(map[uint32]bool)

			// Generate unique random values in the universe
			for i := 0; i < scenario.present; {
				val := r.Uint32() % scenario.universe
				if !valueSet[val] {
					values[i] = val
					valueSet[val] = true
					i++
				}
			}

			b.Run("Add", func(b *testing.B) {
				b.Run("RoaringBitmap", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						rb := NewRoaringBitmap()

						for _, val := range values {
							rb.Add(val)
						}
					}
				})

				b.Run("Hashtable", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						ht := make(map[uint32]bool)

						for _, val := range values {
							ht[val] = true
						}
					}
				})
			})

			b.Run("Contains", func(b *testing.B) {
				// Pre-setup data structures
				rb := NewRoaringBitmap()
				ht := make(map[uint32]bool)
				for _, val := range values {
					rb.Add(val)
					ht[val] = true
				}

				// Generate random lookup values (mix of existing and non-existing)
				lookupValues := make([]uint32, 1000)
				for i := 0; i < 500; i++ {
					lookupValues[i] = values[i%len(values)] // 50% existing
				}
				for i := 500; i < 1000; i++ {
					lookupValues[i] = r.Uint32() % scenario.universe // 50% random (mostly non-existing)
				}

				b.Run("RoaringBitmap", func(b *testing.B) {
					b.ReportAllocs()
					i := 0
					for b.Loop() {
						rb.Contains(lookupValues[i%len(lookupValues)])
						i++
					}
				})

				b.Run("Hashtable", func(b *testing.B) {
					b.ReportAllocs()
					i := 0
					for b.Loop() {
						_ = ht[lookupValues[i%len(lookupValues)]]
						i++
					}
				})
			})
		})
	}
}

// BenchmarkClusteredData_RoaringVsHashtable proves "Roaring beats hashtable for clustered data"
// Tests scenarios with long runs and repeated ranges where RoaringBitmap excels
func BenchmarkClusteredData_RoaringVsHashtable(b *testing.B) {
	scenarios := []struct {
		name        string
		description string
		generator   func() []uint32
	}{
		{
			"LongRuns-10K-Elements",
			"10 runs of 1000 consecutive integers each",
			func() []uint32 {
				var values []uint32
				for run := 0; run < 10; run++ {
					start := uint32(run * 100000) // Space runs apart
					for i := uint32(0); i < 1000; i++ {
						values = append(values, start+i)
					}
				}
				return values
			},
		},
		{
			"RepeatedRanges-50K-Elements",
			"50 ranges of 1000 consecutive integers",
			func() []uint32 {
				var values []uint32
				for run := 0; run < 50; run++ {
					start := uint32(run * 50000) // Space runs apart
					for i := uint32(0); i < 1000; i++ {
						values = append(values, start+i)
					}
				}
				return values
			},
		},
		{
			"MixedClusters-25K-Elements",
			"Mixed: 20% clustered, 80% random within clusters",
			func() []uint32 {
				r := rand.New(rand.NewSource(42))
				var values []uint32

				// 5 clusters, each containing 5000 values
				for cluster := 0; cluster < 5; cluster++ {
					clusterStart := uint32(cluster * 1000000)
					clusterSize := uint32(100000) // Cluster spans 100K range

					for i := 0; i < 5000; i++ {
						// Add some consecutive values (clustered)
						if i < 1000 {
							values = append(values, clusterStart+uint32(i))
						} else {
							// Add random values within cluster range
							val := clusterStart + r.Uint32()%clusterSize
							values = append(values, val)
						}
					}
				}
				return values
			},
		},
	}

	for _, scenario := range scenarios {
		values := scenario.generator()

		b.Run(scenario.name, func(b *testing.B) {
			b.Run("Add", func(b *testing.B) {
				b.Run("RoaringBitmap", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						rb := NewRoaringBitmap()

						for _, val := range values {
							rb.Add(val)
						}
					}
				})

				b.Run("Hashtable", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						ht := make(map[uint32]bool, len(values))

						for _, val := range values {
							ht[val] = true
						}
					}
				})
			})

			b.Run("Union", func(b *testing.B) {
				// Split values into two sets for union operation
				mid := len(values) / 2
				values1 := values[:mid]
				values2 := values[mid:]

				// Setup RoaringBitmaps
				rb1 := NewRoaringBitmap()
				rb2 := NewRoaringBitmap()
				for _, val := range values1 {
					rb1.Add(val)
				}
				for _, val := range values2 {
					rb2.Add(val)
				}

				// Setup Hashtables
				ht1 := make(map[uint32]bool)
				ht2 := make(map[uint32]bool)
				for _, val := range values1 {
					ht1[val] = true
				}
				for _, val := range values2 {
					ht2[val] = true
				}

				b.Run("RoaringBitmap", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						_ = rb1.Union(rb2)
					}
				})

				b.Run("Hashtable", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						result := make(map[uint32]bool, len(ht1)+len(ht2))
						for k := range ht1 {
							result[k] = true
						}
						for k := range ht2 {
							result[k] = true
						}
						_ = result
					}
				})
			})
		})
	}
}

// BenchmarkMemoryEfficiency_RoaringVsHashtable proves "Roaring beats hashtable for memory efficiency"
// Tests memory allocations and usage patterns
func BenchmarkMemoryEfficiency_RoaringVsHashtable(b *testing.B) {
	scenarios := []struct {
		name   string
		size   int
		values func(size int) []uint32
	}{
		{
			"Sequential-Dense",
			100000,
			func(size int) []uint32 {
				values := make([]uint32, size)
				for i := 0; i < size; i++ {
					values[i] = uint32(i)
				}
				return values
			},
		},
		{
			"Sparse-Scattered",
			10000,
			func(size int) []uint32 {
				values := make([]uint32, size)
				for i := 0; i < size; i++ {
					values[i] = uint32(i * 1000) // Every 1000th number
				}
				return values
			},
		},
		{
			"Clustered-Runs",
			50000,
			func(size int) []uint32 {
				var values []uint32
				runsCount := 50
				runSize := size / runsCount

				for run := 0; run < runsCount; run++ {
					start := uint32(run * 100000) // Space runs far apart
					for i := 0; i < runSize; i++ {
						values = append(values, start+uint32(i))
					}
				}
				return values
			},
		},
	}

	for _, scenario := range scenarios {
		values := scenario.values(scenario.size)

		b.Run(fmt.Sprintf("%s-Size-%d", scenario.name, scenario.size), func(b *testing.B) {
			b.Run("Construction", func(b *testing.B) {
				b.Run("RoaringBitmap", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						rb := NewRoaringBitmap()
						for _, val := range values {
							rb.Add(val)
						}
						// Keep reference to prevent GC optimization
						_ = rb.Cardinality()
					}
				})

				b.Run("Hashtable", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						ht := make(map[uint32]bool, len(values))
						for _, val := range values {
							ht[val] = true
						}
						// Keep reference to prevent GC optimization
						_ = len(ht)
					}
				})
			})

			b.Run("Copy", func(b *testing.B) {
				// Pre-create source data structures
				rb := NewRoaringBitmap()
				ht := make(map[uint32]bool)
				for _, val := range values {
					rb.Add(val)
					ht[val] = true
				}

				b.Run("RoaringBitmap", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						// Create copy via union with empty bitmap
						empty := NewRoaringBitmap()
						_ = rb.Union(empty)
					}
				})

				b.Run("Hashtable", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						copy := make(map[uint32]bool, len(ht))
						for k, v := range ht {
							copy[k] = v
						}
						_ = copy
					}
				})
			})

			b.Run("Serialization", func(b *testing.B) {
				rb := NewRoaringBitmap()
				ht := make(map[uint32]bool)
				for _, val := range values {
					rb.Add(val)
					ht[val] = true
				}

				b.Run("RoaringBitmap-ToArray", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						_ = rb.ToArray()
					}
				})

				b.Run("Hashtable-ToSlice", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						slice := make([]uint32, 0, len(ht))
						for k := range ht {
							slice = append(slice, k)
						}
						_ = slice
					}
				})
			})
		})
	}
}
