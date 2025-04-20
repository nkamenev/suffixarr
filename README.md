# Suffix Array Implementation in Go

This project provides an efficient implementation of a suffix array using the **SA-IS algorithm** (Suffix Array Induced Sorting) in Go. It supports both small (≤256 characters) and arbitrary alphabets, optimized for performance and memory usage. The implementation is designed for string processing tasks such as pattern matching, substring search, and text indexing.

## Features

- **SA-IS Algorithm**: Linear-time construction of suffix arrays for small and arbitrary alphabets.
- **Small Alphabet Support**: Optimized for texts with up to 256 unique characters (e.g., ASCII).
- **Arbitrary Alphabet Support**: Handles large or arbitrary alphabets using a map-based bucketing approach.
- **Prefix Search**: Includes methods to find all occurrences of a prefix, with results in lexicographical or text order.
- **Memory Efficiency**: Reuses arrays and minimizes allocations during construction.

## Installation

1. Install the package using:

   ```bash
   go get github.com/nekitakamenev/suffixarr
   ```
2. Ensure the Go environment is set up (Go 1.18 or later recommended).
3. Import the package in your Go code:

   ```go
   import "github.com/nekitakamenev/suffixarr"
   ```

## Usage

### Creating a Suffix Array

```go
package main

import (
	"fmt"
	"github.com/nekitakamenev/suffixarr"
)

func main() {
	// Example text: "banana" as int32 slice (ASCII values)
	text := []int32("banana")
	sa := suffixarr.New(text)
}
```

### Finding Prefix Occurrences

The `Lookup` method returns indices of suffixes starting with a given prefix in lexicographical order. The `LookupTextOrd` method returns the same indices sorted by their position in the text.

```go
// Find all occurrences of prefix "a"
prefix := []int32{'a'}
indices := sa.Lookup(prefix)
fmt.Println("Occurrences of 'a' (lexicographical order):", indices)

// Find occurrences in text order
textOrdered := sa.LookupTextOrd(prefix)
fmt.Println("Occurrences of 'a' (text order):", textOrdered)
```

### Example Output

For text `"banana"`:

- Suffix array: `[5 3 1 0 4 2]` (indices of suffixes `["a", "ana", "anana", "banana", "na", "nana"]`).
- `Lookup("a")`: `[5 3 1]` (suffixes `"a"`, `"ana"`, `"anana"`).
- `LookupTextOrd("a")`: `[1 3 5]` (same indices sorted by text position).

## Algorithm Details

The implementation uses the **SA-IS algorithm**, which constructs a suffix array in O(n) time for a text of length n. Key features:

- **Small Alphabets**: Uses array-based bucketing for efficiency.
- **Arbitrary Alphabets**: Employs a map-based approach with probabilistic counting to handle large alphabets.
- **LMS Substrings**: Leverages Left-Most S-type (LMS) substrings for recursive construction.
- **Induced Sorting**: Efficiently sorts suffixes by inducing L-type and S-type suffixes from LMS positions.

## Performance

- **Time Complexity**: O(n) for suffix array construction, O(m + log n) for prefix lookup (where m is the prefix length).
- **Space Complexity**: O(n) for the suffix array and auxiliary data structures.
- **Optimization**: Minimizes memory allocations by reusing arrays and supports large texts efficiently.

## Testing
To run tests, use:
```bash
go test -v
```

## Contributing

Contributions are welcome! Please submit issues or pull requests for bug fixes, optimizations, or additional features. Ensure code follows Go conventions.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
