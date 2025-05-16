# http2struct: Go Library for Converting / Binding HTTP Requests to Structs

[![Go Reference](https://pkg.go.dev/badge/github.com/nemre/http2struct.svg)](https://pkg.go.dev/github.com/nemre/http2struct)
[![Go Report Card](https://goreportcard.com/badge/github.com/nemre/http2struct)](https://goreportcard.com/report/github.com/nemre/http2struct)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/nemre/http2struct.svg)](https://github.com/nemre/http2struct/releases)

`http2struct` is a lightweight, zero-dependency Go library that simplifies HTTP request processing by allowing you to easily transfer data from HTTP requests directly into Go structs. The library handles data from multiple sources including headers, URL query parameters, path parameters, form data, file uploads, and JSON body.

This streamlined approach to request binding eliminates boilerplate code and helps you write more readable, maintainable, and error-resistant applications.

## Features

- **Zero Dependencies:** Built using only Go's standard library
- **Single Function API:** Converts an HTTP request to a struct with a single function call
- **Comprehensive Source Support:** 
  - JSON body data (`json` tag)
  - Form data (`form` tag)
  - URL query parameters (`query` tag)
  - Path parameters (`path` tag)
  - HTTP headers (`header` tag)
  - File uploads - both multipart form (`file` tag) and binary (`file:"binary"` tag)
- **Automatic Type Conversion:** Handles conversion to various Go types:
  - Boolean: `bool`
  - Integers: `int`, `int8`, `int16`, `int32`, `int64`
  - Unsigned integers: `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`
  - Floating point: `float32`, `float64`
  - Complex numbers: `complex64`, `complex128`
  - Strings: `string`
  - Slices of the above types (comma-separated values are automatically split)
- **File Upload Handling:** Manages both multipart form files and binary file uploads with the built-in `File` struct
- **Extensive Error Reporting:** Provides detailed error messages for debugging
- **Smart Data Binding:** Unlike some other binders, only binds fields with data present in the request, preventing invisible problems

## Benefits

| Benefits                | Before (using standard net/http)            | After (using nemre/http2struct)                                                                |
| ----------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------------- |
| ⌛️ Developer Time      | 😫 Expensive (too much parsing code)        | 🚀 **Faster** (define the struct and leave parsing to http2struct)                             |
| ♻️ Code Repetition     | 😞 High                                    | 😍 **Lower** (eliminates repetitive request parsing code)                                      |
| 📖 Code Readability     | 😟 Poor                                    | 🤩 **Highly readable** (declarative approach with struct tags)                                 |
| 🔨 Maintainability      | 😡 Poor                                    | 🥰 **Highly maintainable** (centralized request handling logic)                                |
| 🐞 Error Handling       | 😖 Manual for each field                   | 😎 **Comprehensive** (detailed errors for debugging)                                           |
| 🔄 Type Safety          | 😨 Manual type conversion                  | 😌 **Automatic** (type-safe conversions with validation)                                       |

## Installation

```bash
go get -u github.com/nemre/http2struct
```

## Basic Usage

Define a struct with appropriate tags and convert your request:

```go
type UserRequest struct {
    Name      string   `json:"name"`           // From JSON body
    Age       int      `json:"age"`            // From JSON body
    Token     string   `header:"Authorization"` // From request header
    Page      int      `query:"page"`          // From URL query parameter
    UserID    uint64   `path:"user_id"`        // From path parameter
    Nickname  string   `form:"nickname"`       // From form data
    Tags      []string `query:"tags"`          // Handles comma-separated values
    Avatar    File     `file:"avatar"`         // File upload from multipart form
}

func handler(w http.ResponseWriter, r *http.Request) {
    var req UserRequest
    
    if err := http2struct.Convert(r, &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Now req is populated with data from the request
    fmt.Fprintf(w, "Hello, %s!", req.Name)
}
```

## Advanced Usage

### File Uploads

`http2struct` provides a built-in `File` struct for handling file uploads:

```go
type File struct {
    Name    string // Original filename
    Size    int64  // File size in bytes
    Content []byte // File content
}
```

#### Multipart Form File Uploads

```go
type UploadRequest struct {
    // As a value
    Avatar File `file:"avatar"`
    
    // Or as a pointer
    Document *File `file:"document"`
}
```

#### Binary File Upload (Entire Request Body)

```go
type BinaryUploadRequest struct {
    // Use the special "binary" tag value
    File File `file:"binary"`
    
    // Additional metadata can come from headers
    ContentType string `header:"Content-Type"`
    Filename    string `header:"X-Filename"`
}
```

### Handling Multiple Data Sources

`http2struct` allows you to combine data from multiple sources in a single request:

```go
type ComplexRequest struct {
    // User data from JSON body
    User struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    } `json:"user"`
    
    // Configuration from query parameters
    Page  int  `query:"page"`
    Limit int  `query:"limit"`
    
    // Authentication from headers
    Token string `header:"Authorization"`
    
    // Resource identifier from path
    ID uint64 `path:"id"`
    
    // File uploads
    Avatar    File  `file:"avatar"`
    Documents []File `file:"documents"` // Not currently supported, shown for future consideration
}
```

## Error Handling

The `Convert` function returns detailed errors to help diagnose issues:

```go
err := http2struct.Convert(r, &req)
if err != nil {
    // Handle the error
    log.Printf("Request conversion error: %v", err)
    http.Error(w, "Bad request format", http.StatusBadRequest)
    return
}
```

Error messages are descriptive, indicating:
- Invalid destination types
- Field conversion failures
- Unsupported types
- Form parsing errors
- JSON decoding issues

## Best Practices

- **Validate Input Data**: While `http2struct` handles conversion, you should still validate the business logic of the data
- **Use Appropriate Types**: Choose struct field types that match the expected data to avoid conversion errors
- **Consider Performance**: For large file uploads, process files directly rather than loading them all into memory
- **Set Default Values**: Initialize struct fields with default values before conversion for optional parameters

## Comparison with Other Libraries

| Feature                 | http2struct             | gorilla/schema         | gin binding           |
|-------------------------|-------------------------|------------------------|----------------------|
| Zero Dependencies       | ✅                      | ❌                     | ❌                   |
| JSON Body Support       | ✅                      | ❌                     | ✅                   |
| Form Data Support       | ✅                      | ✅                     | ✅                   |
| Query Parameter Support | ✅                      | ✅                     | ✅                   |
| Path Parameter Support  | ✅                      | ❌                     | ❌                   |
| Header Support          | ✅                      | ❌                     | ❌                   |
| File Upload Support     | ✅                      | ❌                     | ✅                   |
| Binary File Support     | ✅                      | ❌                     | ❌                   |
| Detailed Error Messages | ✅                      | ❌                     | ✅                   |

## Contributing

Contributions to improve `http2struct` are welcome! Here's how you can contribute:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## FAQ

### Q: Can I use http2struct with other web frameworks?
**A:** Yes, the library works with any framework that uses the standard `net/http.Request` object, including Gin, Echo, Chi, etc.

### Q: How does http2struct handle arrays or slices of values?
**A:** For query parameters, path parameters, headers, and form values, comma-separated strings are automatically split and converted to slices of the appropriate type.

### Q: What happens if a field can't be converted to the target type?
**A:** The library will return a detailed error explaining which field failed conversion and why.

### Q: Can I use nested structs?
**A:** Yes, JSON body data can be mapped to nested structs. Other sources (query, path, header, form) work with flat structures.
