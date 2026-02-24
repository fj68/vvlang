package mod

import (
	"reflect"
	"testing"
)

func TestParseRemotePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    *RemoteModule
		wantErr bool
	}{
		{
			name: "github simple",
			path: "github.com/user/repo/main.vv",
			want: &RemoteModule{
				Domain:  "github.com",
				User:    "user",
				Repo:    "repo",
				Version: "",
				File:    "main.vv",
			},
			wantErr: false,
		},
		{
			name: "github with version",
			path: "github.com/user/repo@v1.0.0/lib/math.vv",
			want: &RemoteModule{
				Domain:  "github.com",
				User:    "user",
				Repo:    "repo",
				Version: "v1.0.0",
				File:    "lib/math.vv",
			},
			wantErr: false,
		},
		{
			name: "bitbucket simple",
			path: "bitbucket.org/org/proj/file.vv",
			want: &RemoteModule{
				Domain:  "bitbucket.org",
				User:    "org",
				Repo:    "proj",
				Version: "",
				File:    "file.vv",
			},
			wantErr: false,
		},
		{
			name:    "invalid domain",
			path:    "example.com/user/repo/file.vv",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "path too short",
			path:    "github.com/user",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRemotePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRemotePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRemotePath() = %v, want %v", got, tt.want)
			}
		})
	}
}