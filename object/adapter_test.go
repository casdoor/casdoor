package object

import (
	"testing"
)

func Test_getCollectionName(t *testing.T) {
	type args struct {
		objectPtr interface{}
	}

	applications := []*Application{}
	applications = append(applications, &Application{Name: "app1"})
	applications = append(applications, &Application{Name: "app2"})
	tests := []struct {
		name     string
		args     args
		wantName string
	}{
		{
			name:     "one application",
			args:     args{&Application{Name: "app"}},
			wantName: "application",
		},
		{
			name:     "applications",
			args:     args{applications},
			wantName: "application",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotName := getCollectionName(tt.args.objectPtr); gotName != tt.wantName {
				t.Errorf("getCollectionName() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}
