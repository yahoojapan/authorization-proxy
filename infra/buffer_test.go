package infra

import (
	"testing"

	"fmt"
	"reflect"
	"sync"
)

// NotEqualError reports the name of the field having different value and their values.
type NotEqualError struct {
	Field string
	Got   interface{}
	Want  interface{}
}

// Error formats NotEqualError.
func (e *NotEqualError) Error() string {
	return fmt.Sprintf("%s got = %v, want %v", e.Field, e.Got, e.Want)
}

func TestNewBuffer(t *testing.T) {
	type args struct {
		size uint64
	}
	type testcase struct {
		name      string
		args      args
		want      *buffer
		checkFunc func(got, want *buffer) error
	}
	tests := []testcase{
		{
			name: "Check newBuffer, with 0 size",
			args: args{
				size: 0,
			},
			want: nil,
			checkFunc: func(got, want *buffer) error {
				if !reflect.DeepEqual(got, want) {
					return &NotEqualError{"", got, want}
				}
				return nil
			},
		},
		{
			name: "Check newBuffer, positive size",
			args: args{
				size: 37,
			},
			want: &buffer{
				size: func(i uint64) *uint64 { return &i }(37),
			},
			checkFunc: func(got, want *buffer) error {
				if *(got.size) != *(want.size) {
					return &NotEqualError{"size", *(got.size), *(want.size)}
				}

				buffer := got.Get()
				if uint64(cap(buffer)) != *(want.size) {
					return &NotEqualError{"pool", cap(buffer), *(want.size)}
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := NewBuffer(tt.args.size)

			if got == nil && tt.want == nil {
				// skip on both nil
				return
			}
			if err := tt.checkFunc(got.(*buffer), tt.want); err != nil {
				t.Errorf("newBuffer() %v", err)
				return
			}
		})
	}
}

func Test_buffer_Get(t *testing.T) {
	type fields struct {
		pool sync.Pool
		size *uint64
	}
	type testcase struct {
		name   string
		fields fields
		want   []byte
	}
	tests := []testcase{
		{
			name: "Check buffer Get, get from internal pool",
			fields: fields{
				pool: sync.Pool{
					New: func() interface{} {
						return []byte("pool-new-91")
					},
				},
			},
			want: []byte("pool-new-91"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			b := &buffer{
				pool: tt.fields.pool,
				size: tt.fields.size,
			}

			got := b.Get()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buffer.Get() %v", &NotEqualError{"", got, tt.want})
				return
			}
		})
	}
}

func Test_buffer_Put(t *testing.T) {
	type fields struct {
		pool sync.Pool
		size *uint64
	}
	type args struct {
		buf []byte
	}
	type testcase struct {
		name      string
		fields    fields
		args      args
		checkFunc func(got *buffer) error
	}
	tests := []testcase{
		{
			name: "Check buffer Put, with 0 size",
			fields: fields{
				pool: sync.Pool{New: func() interface{} { return make([]byte, 0, 134) }},
				size: func(i uint64) *uint64 { return &i }(135),
			},
			args: args{
				buf: make([]byte, 0),
			},
			checkFunc: func(got *buffer) error {
				wantSize := uint64(135)
				wantBufLen := 0
				wantBufCap := 0
				mayWantBufCap := 134

				gotSize := *(got.size)
				if gotSize != wantSize {
					return &NotEqualError{"size", gotSize, wantSize}
				}

				gotBuffer := got.Get() // for sync.Pool, get may returns from Put() or from New()
				gotBufLen := len(gotBuffer)
				if gotBufLen != wantBufLen {
					return &NotEqualError{"buffer len", gotBufLen, wantBufLen}
				}
				gotBufCap := cap(gotBuffer)
				if gotBufCap != wantBufCap && gotBufCap != mayWantBufCap {
					return &NotEqualError{"buffer cap", gotBufCap, wantBufCap}
				}
				return nil
			},
		},
		{
			name: "Check buffer Put, with buffer len and cap > current size",
			fields: fields{
				pool: sync.Pool{New: func() interface{} { return make([]byte, 0, 165) }},
				size: func(i uint64) *uint64 { return &i }(166),
			},
			args: args{
				buf: make([]byte, 169),
			},
			checkFunc: func(got *buffer) error {
				wantSize := uint64(169)
				wantBufLen := 0
				wantBufCap := 169
				mayWantBufCap := 165

				gotSize := *(got.size)
				if gotSize != wantSize {
					return &NotEqualError{"size", gotSize, wantSize}
				}

				gotBuffer := got.Get() // for sync.Pool, get may returns from Put() or from New()
				gotBufLen := len(gotBuffer)
				if gotBufLen != wantBufLen {
					return &NotEqualError{"len(buffer)", gotBufLen, wantBufLen}
				}
				gotBufCap := cap(gotBuffer)
				if gotBufCap != wantBufCap && gotBufCap != mayWantBufCap {
					return &NotEqualError{"cap(buffer)", gotBufCap, wantBufCap}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			b := &buffer{
				pool: tt.fields.pool,
				size: tt.fields.size,
			}

			b.Put(tt.args.buf)
			if err := tt.checkFunc(b); err != nil {
				t.Errorf("buffer.Put() %v", err)
				return
			}
		})
	}
}

func Test_max(t *testing.T) {

	type args struct {
		x uint64
		y uint64
	}
	type testcase struct {
		name string
		args args
		want uint64
	}
	tests := []testcase{
		{
			name: "Check max, x < y",
			args: args{
				x: uint64(227),
				y: uint64(228),
			},
			want: uint64(228),
		},
		{
			name: "Check max, x == y",
			args: args{
				x: uint64(235), y: uint64(235),
			},
			want: uint64(235),
		},
		{
			name: "Check max, x > y",
			args: args{
				y: uint64(242),
				x: uint64(243),
			},
			want: uint64(243),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := max(tt.args.x, tt.args.y)
			if got != tt.want {
				t.Errorf("max() %v", &NotEqualError{"", got, tt.want})
				return
			}
		})
	}
}
