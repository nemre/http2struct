// Package http2struct provides functionality to automatically map HTTP request data
// into Go struct fields using struct tags.
//
// It supports mapping from various sources:
// - JSON request body
// - Form fields
// - URL query parameters
// - Path parameters
// - HTTP headers
// - File uploads (both multipart and binary)
package http2struct

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// File represents an uploaded file from an HTTP request
type File struct {
	Name    string // Original filename provided by the client
	Size    int64  // Size of the file in bytes
	Content []byte // Raw content of the file
}

// Convert maps data from an HTTP request into a struct.
// The destination must be a pointer to a struct with appropriate tags.
//
// Supported struct tags:
// - `json:"field_name"` - Maps JSON body fields
// - `form:"field_name"` - Maps form fields
// - `query:"param_name"` - Maps URL query parameters
// - `path:"param_name"` - Maps URL path parameters
// - `header:"Header-Name"` - Maps HTTP headers
// - `file:"field_name"` - Maps uploaded files from multipart forms
// - `file:"binary"` - Maps the entire request body as a file
func Convert(request *http.Request, destination any) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	destinationType := reflect.TypeOf(destination)

	if destinationType == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	if destinationType.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	destinationType = destinationType.Elem()

	if destinationType.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a struct")
	}

	if err := convertBody(request, destination, destinationType); err != nil {
		return fmt.Errorf("failed to convert body: %w", err)
	}

	v := reflect.ValueOf(destination).Elem()

	for i := range destinationType.NumField() {
		field := destinationType.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		fieldValue.SetZero()

		tag, ok := field.Tag.Lookup("form")
		if ok && tag != "" && tag != "-" {
			if request.PostForm == nil {
				if err := request.ParseMultipartForm(32 << 20); err != nil {
					return fmt.Errorf("failed to parse request multipart form: %w", err)
				}
			}

			var v string

			if p := request.PostForm[tag]; len(p) > 0 {
				v = p[0]
			}

			if err := convert(fieldValue, field.Type, v); err != nil {
				return fmt.Errorf("failed to convert %q form to %q field: %w", tag, field.Name, err)
			}

			continue
		}

		tag, ok = field.Tag.Lookup("file")
		if ok && tag != "" && tag != "-" && tag != "binary" {
			if field.Type.Kind() != reflect.Pointer && field.Type != reflect.TypeOf(File{}) {
				return fmt.Errorf("%q type is not supported for %q field", fieldValue.Type().String(), field.Name)
			}

			if field.Type.Kind() == reflect.Pointer && field.Type != reflect.TypeOf(&File{}) {
				return fmt.Errorf("%q type is not supported for %q field", fieldValue.Type().String(), field.Name)
			}

			base, _, _ := strings.Cut(request.Header.Get("Content-Type"), ";")

			if strings.TrimSpace(base) != "multipart/form-data" {
				continue
			}

			file, fileHeader, err := request.FormFile(tag)
			if errors.Is(err, http.ErrMissingFile) {
				continue
			}
			if err != nil {
				return fmt.Errorf("failed to get %q form file for %q field: %w", tag, field.Name, err)
			}

			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read %q form file content for %q field: %w", tag, field.Name, err)
			}

			f := File{
				Name:    fileHeader.Filename,
				Size:    fileHeader.Size,
				Content: content,
			}

			if field.Type.Kind() == reflect.Pointer {
				fieldValue.Set(reflect.ValueOf(&f))

				continue
			}

			fieldValue.Set(reflect.ValueOf(f))

			continue
		}

		tag, ok = field.Tag.Lookup("file")
		if ok && tag == "binary" {
			if field.Type.Kind() != reflect.Pointer && field.Type != reflect.TypeOf(File{}) {
				return fmt.Errorf("%q type is not supported for %q field", fieldValue.Type().String(), field.Name)
			}

			if field.Type.Kind() == reflect.Pointer && field.Type != reflect.TypeOf(&File{}) {
				return fmt.Errorf("%q type is not supported for %q field", fieldValue.Type().String(), field.Name)
			}

			if request.ContentLength == 0 {
				return nil
			}

			contentDisposition := request.Header.Get("Content-Disposition")

			_, params, err := mime.ParseMediaType(contentDisposition)
			if err != nil {
				continue
			}

			filename := params["filename"]
			if filename == "" {
				filename = params["filename*"]
			}

			if filename == "" {
				continue
			}

			content, err := io.ReadAll(request.Body)
			if err != nil {
				return fmt.Errorf("failed to read %q raw body for %q field: %w", tag, field.Name, err)
			}

			f := File{
				Name:    filename,
				Size:    request.ContentLength,
				Content: content,
			}

			if field.Type.Kind() == reflect.Pointer {
				fieldValue.Set(reflect.ValueOf(&f))

				continue
			}

			fieldValue.Set(reflect.ValueOf(f))

			continue
		}

		tag, ok = field.Tag.Lookup("header")
		if ok && tag != "" && tag != "-" {
			v := request.Header.Get(tag)

			if err := convert(fieldValue, field.Type, v); err != nil {
				return fmt.Errorf("failed to convert %q header to %q field: %w", tag, field.Name, err)
			}

			continue
		}

		tag, ok = field.Tag.Lookup("query")
		if ok && tag != "" && tag != "-" {
			v := request.URL.Query().Get(tag)

			if err := convert(fieldValue, field.Type, v); err != nil {
				return fmt.Errorf("failed to convert %q query to %q field: %w", tag, field.Name, err)
			}

			continue
		}

		tag, ok = field.Tag.Lookup("path")
		if ok && tag != "" && tag != "-" {
			v := request.PathValue(tag)

			if err := convert(fieldValue, field.Type, v); err != nil {
				return fmt.Errorf("failed to convert %q path to %q field: %w", tag, field.Name, err)
			}

			continue
		}
	}

	return nil
}

func convertBody(request *http.Request, destination any, destinationType reflect.Type) error {
	if request.ContentLength == 0 {
		return nil
	}

	base, _, _ := strings.Cut(request.Header.Get("Content-Type"), ";")

	if strings.TrimSpace(base) != "application/json" {
		return nil
	}

	for i := range destinationType.NumField() {
		field := destinationType.Field(i)

		if !field.IsExported() {
			continue
		}

		tag, ok := field.Tag.Lookup("json")
		if !ok {
			continue
		}

		if tag == "-" {
			continue
		}

		if err := json.NewDecoder(request.Body).Decode(destination); err != nil {
			return fmt.Errorf("failed to decode request body: %w", err)
		}

		break
	}

	return nil
}

func convert(field reflect.Value, fieldType reflect.Type, value string) error {
	if value == "" {
		return nil
	}

	var err error

	switch field.Kind() {
	case reflect.Bool:
		var v bool

		v, err = strconv.ParseBool(value)
		if err == nil {
			field.SetBool(v)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var v int64

		v, err = strconv.ParseInt(value, 10, fieldType.Bits())
		if err == nil {
			field.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var v uint64

		v, err = strconv.ParseUint(value, 10, fieldType.Bits())
		if err == nil {
			field.SetUint(v)
		}
	case reflect.Float32, reflect.Float64:
		var v float64

		v, err = strconv.ParseFloat(value, fieldType.Bits())
		if err == nil {
			field.SetFloat(v)
		}
	case reflect.Complex64, reflect.Complex128:
		var v complex128

		v, err = strconv.ParseComplex(value, fieldType.Bits())
		if err == nil {
			field.SetComplex(v)
		}
	case reflect.Slice:
		element := fieldType.Elem()

		if element.Kind() == reflect.Slice {
			return fmt.Errorf("slice element kind %q is not supported", element.Kind().String())
		}

		parts := strings.Split(value, ",")
		slice := reflect.MakeSlice(fieldType, len(parts), len(parts))

		for i, part := range parts {
			if err := convert(slice.Index(i), element, part); err != nil {
				return fmt.Errorf("failed to convert slice element for index %d: %w", i, err)
			}
		}

		field.Set(slice)
	case reflect.String:
		field.SetString(value)
	default:
		return fmt.Errorf("kind %q is not supported", field.Kind().String())
	}

	if err != nil {
		return fmt.Errorf("failed to parse value to %q: %w", field.Kind().String(), err)
	}

	return nil
}
