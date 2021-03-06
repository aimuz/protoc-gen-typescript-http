package plugin

import (
	"encoding/binary"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func descriptorTypeName(desc protoreflect.Descriptor) string {
	name := string(desc.Name())
	var prefix string
	if desc.Parent() != desc.ParentFile() {
		prefix = descriptorTypeName(desc.Parent()) + "_"
	}
	return prefix + name
}

func rangeFields(message protoreflect.MessageDescriptor, f func(field protoreflect.FieldDescriptor)) {
	for i := 0; i < message.Fields().Len(); i++ {
		f(message.Fields().Get(i))
	}
}

func rangeEnumValues(enum protoreflect.EnumDescriptor, f func(value protoreflect.EnumValueDescriptor)) {
	for i := 0; i < enum.Values().Len(); i++ {
		f(enum.Values().Get(i))
	}
}

func rangeFileMessages(file protoreflect.FileDescriptor, f func(message protoreflect.MessageDescriptor)) {
	for i := 0; i < file.Messages().Len(); i++ {
		f(file.Messages().Get(i))
		rangeNestedMessages(file.Messages().Get(i), f)
	}
}

func rangeNestedMessages(msg protoreflect.MessageDescriptor, f func(message protoreflect.MessageDescriptor)) {
	for i := 0; i < msg.Messages().Len(); i++ {
		nested := msg.Messages().Get(i)
		f(msg.Messages().Get(i))
		rangeNestedMessages(nested, f)
	}
}

func t(n int) string {
	return strings.Repeat("\t", n)
}

type fileSourcePath = int32

const (
	fileSourcePathMessage fileSourcePath = 4
	fileSourcePathEnum                   = 5
)

type messageSourcePath = int32

const (
	messageSourcePathField   messageSourcePath = 2
	messageSourcePathMessage                   = 3
	messageSourcePathEnum                      = 4
)

type enumSourcePath = int32

const (
	enumSourcePathValue enumSourcePath = 2
)

func descriptorSourcePath(desc protoreflect.Descriptor) protoreflect.SourcePath {
	if _, ok := desc.(protoreflect.FileDescriptor); ok {
		return nil
	}
	var path protoreflect.SourcePath
	switch desc.Parent().(type) {
	case protoreflect.FileDescriptor:
		switch v := desc.(type) {
		case protoreflect.MessageDescriptor:
			path = protoreflect.SourcePath{fileSourcePathMessage, int32(v.Index())}
		case protoreflect.EnumDescriptor:
			path = protoreflect.SourcePath{fileSourcePathEnum, int32(v.Index())}
		}
	case protoreflect.MessageDescriptor:
		switch v := desc.(type) {
		case protoreflect.FieldDescriptor:
			path = protoreflect.SourcePath{messageSourcePathField, int32(v.Index())}
		case protoreflect.MessageDescriptor:
			path = protoreflect.SourcePath{messageSourcePathMessage, int32(v.Index())}
		case protoreflect.EnumDescriptor:
			path = protoreflect.SourcePath{messageSourcePathEnum, int32(v.Index())}
		}
	case protoreflect.EnumDescriptor:
		switch v := desc.(type) {
		case protoreflect.EnumValueDescriptor:
			path = protoreflect.SourcePath{enumSourcePathValue, int32(v.Index())}
		}
	}
	return append(descriptorSourcePath(desc.Parent()), path...)
}

func descriptorSourceLocation(desc protoreflect.Descriptor, path protoreflect.SourcePath) (protoreflect.SourceLocation, bool) {
	locs := desc.ParentFile().SourceLocations()
	key := newPathKey(path)
	for i := 0; i < locs.Len(); i++ {
		loc := locs.Get(i)
		if newPathKey(loc.Path) == key {
			return loc, true
		}
	}
	return protoreflect.SourceLocation{}, false
}

// A pathKey is a representation of a location path suitable for use as a map key.
type pathKey string

// newPathKey converts a location path to a pathKey.
func newPathKey(path protoreflect.SourcePath) pathKey {
	buf := make([]byte, 4*len(path))
	for i, x := range path {
		binary.LittleEndian.PutUint32(buf[i*4:], uint32(x))
	}
	return pathKey(buf)
}
