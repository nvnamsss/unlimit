# Guideline: Implementing Batch Processing with BatchIterator

This guide provides step-by-step instructions for newcomers to implement batch processing using the `Batch` shortcut and `ExecuteWithReturn` as defined in [iterator.go](iterator.go).

---

## 1. Import the Package

Import the `collections` package in your Go file:

```go
import "github.com/nvnamsss/unlimit/collections"
```

---

## 2. Prepare Your Data

Prepare your data as a slice of any type (e.g., `[]int`, `[]string`, or a slice of structs):

```go
data := []int{1, 2, 3, 4, 5, 6}
```

---

## 3. Process Batches Using the Batch Shortcut

Use the `Batch` function to process each batch with a callback function. This is the simplest way to process batches:

```go
err := collections.Batch(data, 2, func(batch []int) error {
    // Process each batch here
    fmt.Println(batch)
    return nil // or return an error to stop processing early
})
if err != nil {
    // Handle error
}
```

---

## 4. Process Batches and Collect Return Values

If you want to process each batch and collect a result for each batch, use `ExecuteWithReturn`:

```go
iterator := collections.NewBatchIterator(data, 2)
results, err := collections.ExecuteWithReturn(iterator, func(batch []int) (string, error) {
    // Transform each batch and return a value
    return fmt.Sprintf("Batch: %v", batch), nil
})
if err != nil {
    // Handle error
}
fmt.Println(results) // Output: ["Batch: [1 2]", "Batch: [3 4]", "Batch: [5 6]"]
```

You can use any return type in the callback, such as `int`, `bool`, `struct`, or even a slice.

---

## 5. Tips

- Use `Batch` for simple batch processing with error handling.
- Use `ExecuteWithReturn` when you need to collect results from each batch.
- You can chain transformations with `Map`, `Filter`, and `Take` before calling `ExecuteWithReturn` if needed.
- Always check for errors returned by `Batch` or `ExecuteWithReturn`.

---

**See also:**  
- [Unit Testing Guidelines](../.co-guidelines/go/unit_test_module.md)  
- [iterator.go source](iterator.go)
You can chain transformations:

```go
mapped := iterator.Map(func(batch []int) []int {
    // Example: double each value
    for i, v := range batch {
        batch[i] = v * 2
    }
    return batch
})

filtered := mapped.Filter(func(batch []int) bool {
    // Example: only keep batches where the first element is even
    return len(batch) > 0 && batch[0]%2 == 0
})
```

---

## 6. Collect Results

Collect all batches into a slice of slices:

```go
allBatches := iterator.Collect()
fmt.Println(allBatches)
```

---

## 7. Reduce Batches

Combine all batches into a single result:

```go
result := iterator.Reduce(func(acc, batch []int) []int {
    return append(acc, batch...)
})
fmt.Println(result)
```

---

## 8. Advanced: ExecuteWithReturn

Transform each batch and collect results of any type:

```go
results, err := collections.ExecuteWithReturn(iterator, func(batch []int) (string, error) {
    return fmt.Sprintf("Batch: %v", batch), nil
})
fmt.Println(results)
```

---

## 9. Error Handling

- If any batch processing function returns an error, processing stops and the error is returned.
- Always check for errors when using `Batch` or `ExecuteWithReturn`.

---

## 10. Tips

- Use `Take(n)` to limit the number of batches processed.
- Use `Filter` and `Map` for flexible batch transformations.
- The iterator works with any slice type, including custom structs.

---

**See also:**  
- [Unit Testing Guidelines](../.co-guidelines/go/unit_test_module.md)  
- [iterator.go source](iterator.go)
