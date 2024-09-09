package slogmock

import (
	"context"
	"slices"

	"log/slog"

	"github.com/samber/lo"
)

type Option struct {
	// optional
	Enabled func(ctx context.Context, level slog.Level) bool
	// optional
	Handle func(ctx context.Context, record slog.Record) error
}

func (o Option) NewMockHandler() slog.Handler {
	return &MockHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*MockHandler)(nil)

type MockHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

func (h *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.option.Enabled == nil {
		return true
	}

	return h.option.Enabled(ctx, level)
}

func (h *MockHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.option.Handle == nil {
		return nil
	}

	// Clone the record and add the handlers attributes to the new record.
	// I could not just do `record.AddAttrs(h.attrs...)` because h.Attrs must be added before record.Attrs.
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(h.attrs...)

	attrs := []slog.Attr{}
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	for i := range h.groups {
		k := h.groups[len(attrs)-1-i]
		v := attrs
		attrs = []slog.Attr{
			slog.Group(k, lo.ToAnySlice(v)...),
		}
	}
	newRecord.AddAttrs(attrs...)

	return h.option.Handle(ctx, newRecord)
}

func (h *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MockHandler{
		option: h.option,
		attrs:  appendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *MockHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &MockHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func appendAttrsToGroup(groups []string, actualAttrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	actualAttrs = slices.Clone(actualAttrs)

	if len(groups) == 0 {
		return append(actualAttrs, newAttrs...)
	}

	for i := range actualAttrs {
		attr := actualAttrs[i]
		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			actualAttrs[i] = slog.Group(groups[0], lo.ToAnySlice(appendAttrsToGroup(groups[1:], attr.Value.Group(), newAttrs...))...)
			return actualAttrs
		}
	}

	return append(
		actualAttrs,
		slog.Group(
			groups[0],
			lo.ToAnySlice(appendAttrsToGroup(groups[1:], []slog.Attr{}, newAttrs...))...,
		),
	)
}
