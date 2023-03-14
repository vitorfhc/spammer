package spammer

import "testing"

func Test_addPathToHost(t *testing.T) {
	type args struct {
		host string
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should add a path with format /*/",
			args: args{
				host: "http://example.com",
				path: "/test/",
			},
			want:    "http://example.com/test/",
			wantErr: false,
		},
		{
			name: "should add a path with format /*",
			args: args{
				host: "https://example.com",
				path: "/test",
			},
			want:    "https://example.com/test",
			wantErr: false,
		},
		{
			name: "should add a path with format *",
			args: args{
				host: "https://example.com",
				path: "test",
			},
			want:    "https://example.com/test",
			wantErr: false,
		},
		{
			name: "should add a path with format /*/*",
			args: args{
				host: "https://example.com",
				path: "/test/another",
			},
			want:    "https://example.com/test/another",
			wantErr: false,
		},
		{
			name: "should add a path with format /*/*/",
			args: args{
				host: "https://example.com",
				path: "/test/another/",
			},
			want:    "https://example.com/test/another/",
			wantErr: false,
		},
		{
			name: "should add a path with format */*",
			args: args{
				host: "https://example.com",
				path: "test/another",
			},
			want:    "https://example.com/test/another",
			wantErr: false,
		},
		{
			name: "should add a path to a host with no scheme",
			args: args{
				host: "example.com",
				path: "test/another",
			},
			want:    "https://example.com/test/another",
			wantErr: false,
		},
		{
			name: "should add a path to a host with no scheme and no path",
			args: args{
				host: "example.com",
				path: "/test/another",
			},
			want:    "https://example.com/test/another",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeHostAndAddPath(tt.args.host, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("addPathToHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("addPathToHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
