package handler

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/yahoojapan/authorization-proxy/service"
)

func TestNewDebugHandler(t *testing.T) {
	type args struct {
		authd service.Authorizationd
	}
	tests := []struct {
		name string
		args args
		want *DebugHandler
	}{
		{
			name: "new success",
			args: args{
				authd: nil,
			},
			want: &DebugHandler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDebugHandler(tt.args.authd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDebugHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDebugHandler_GetPolicyCache(t *testing.T) {
	type fields struct {
		authd service.Authorizationd
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	type test struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}
	tests := []test{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dh := &DebugHandler{
				authd: tt.fields.authd,
			}
			if err := dh.GetPolicyCache(tt.args.w, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("DebugHandler.GetPolicyCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
