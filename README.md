# http2struct: Go Library for Converting HTTP Requests to Structs

[![Go Reference](https://pkg.go.dev/badge/github.com/nemre/http2struct.svg)](https://pkg.go.dev/github.com/nemre/http2struct)

`http2struct` is a Go library that allows you to easily transfer data from HTTP requests (headers, URL query parameters, path parameters, and JSON body) directly into Go structs.

This simplifies HTTP request processing and helps you write more readable and maintainable code.

## Features

- **Easy to Use:** Converts an HTTP request to a struct with a single function call.
- **Multiple Source Support:** Reads data from request headers (`header` tag), URL query parameters (`query` tag), path parameters (`path` tag), and JSON request body (`json` tag).
- **Automatic Type Conversion:** Automatically converts to all data types (bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128, slice, string).
- **Slice Support:** Can convert comma-separated values into slices.
- **Flexible Tagging:** You can specify data sources through custom tags in your struct fields.
- **Error Handling:** Provides detailed error reporting during the conversion process.
- **Handles Invisible Problems:**

## Benefits

| Benefits                | Before (use net/http package)              | After (use nemre/http2struct package)                                                              |
| ----------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------------- |
| ⌛️ Developer Time      | 😫 Expensive (too much parsing stuff code) | 🚀 **Faster** (define the struct for receiving input data and leave the parsing job to http2struct) |
| ♻️ Code Repetition Rate | 😞 High                                    | 😍 **Lower**                                                                                   |
| 📖 Code Readability     | 😟 Poor                                    | 🤩 **Highly readable**                                                                         |
| 🔨 Maintainability      | 😡 Poor                                    | 🥰 **Highly maintainable**    

## Usage

To add the `http2struct` library to your project, use the following command:
```bash
go get -u github.com/nemre/http2struct
````
Define a struct like this:
```go
type Model struct {
    BodyField   string     `json:"foo"`
    HeaderField uint16     `header:"foo"`
    QueryField  []bool     `query:"foo"`
    PathField   complex128 `path:"foo"`
}
```
Then convert:
```go
var model Model

http2struct.Convert(request, &model)
```
That's all! 👌
