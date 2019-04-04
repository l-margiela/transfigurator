package transform

import "testing"

func Test_isNomadFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "valid",
			path: "/Volumes/TARDIS/Deluxe/repos/docker-consul-lb/deploy.tmpl.nomad",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNomadFile(tt.path); got != tt.want {
				t.Errorf(`isNomadFile("%s") = %v, want %v`, tt.path, got, tt.want)
			}
		})
	}
}
