package object

import (
	"context"
	"os"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	InitConfig()
	os.Exit(m.Run())
}

func Test_getCollectionName(t *testing.T) {
	type args struct {
		objectPtr interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantName string
	}{
		{
			name: "single application",
			args: args{&Application{
				Name: "app",
			}},
			wantName: "application",
		},
		{
			name: "multiple applications",
			args: args{&[]*Application{
				{
					Name: "app1",
				},
				{
					Name: "app2",
				},
			}},
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

func TestSession_Insert(t *testing.T) {
	type args struct {
		ctx   context.Context
		beans interface{}
	}
	tests := []struct {
		name    string
		s       *Session
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "one object",
			s:       adapter.CreateSession(context.Background(), nil),
			args:    args{ctx: context.TODO(), beans: &Application{Name: "app"}},
			want:    1,
			wantErr: false,
		},
		{
			name:    "object slice",
			s:       adapter.CreateSession(context.Background(), nil),
			args:    args{ctx: context.TODO(), beans: &[]*Application{{Name: "app1"}, {Name: "app2"}}},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Insert(tt.args.beans)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.Insert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Session.Insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_Get(t *testing.T) {
	type args struct {
		objectPtr interface{}
	}
	tests := []struct {
		name    string
		s       *Session
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "success",
			s:    adapter.CreateSession(context.Background(), ConditionsBuilder().SetEqualFields(EqualFields{"name": "app"})),
			args: args{
				objectPtr: &Application{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "failed",
			s:    adapter.CreateSession(context.Background(), ConditionsBuilder().SetEqualFields(EqualFields{"name": "not existed"})),
			args: args{
				objectPtr: &Application{},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Get(tt.args.objectPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Session.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_Find(t *testing.T) {
	type args struct {
		objectSlicePtr interface{}
	}
	all := []*Application{}
	part := []*Application{}
	tests := []struct {
		name    string
		s       *Session
		args    args
		wantErr bool
	}{
		{
			name: "all data",
			s:    adapter.CreateSession(context.Background(), ConditionsBuilder()),
			args: args{
				objectSlicePtr: &all,
			},
			wantErr: false,
		},
		{
			name: "filter",
			s:    adapter.CreateSession(context.Background(), ConditionsBuilder().SetEqualFields(EqualFields{"name": "app"})),
			args: args{
				objectSlicePtr: &part,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Find(tt.args.objectSlicePtr); (err != nil) != tt.wantErr {
				t.Errorf("Session.Find() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				a := tt.args.objectSlicePtr.(*[]*Application)
				for _, v := range *a {
					t.Logf("%+v", *v)
				}
			}
		})
	}
}

func TestSession_Count(t *testing.T) {
	type args struct {
		objectPtr interface{}
	}
	tests := []struct {
		name    string
		s       *Session
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "all",
			s:       adapter.CreateSession(context.Background(), ConditionsBuilder()),
			args:    args{&Application{}},
			want:    3,
			wantErr: false,
		},
		{
			name:    "filter",
			s:       adapter.CreateSession(context.Background(), ConditionsBuilder().SetEqualFields(EqualFields{"name": "app"})),
			args:    args{&Application{}},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Count(tt.args.objectPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Session.Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_UpdateByID(t *testing.T) {
	type args struct {
		objectPtr interface{}
		owner     string
		name      string
	}
	tests := []struct {
		name    string
		s       *Session
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "success",
			s:    adapter.CreateSession(context.Background(), nil),
			args: args{
				objectPtr: &Application{Owner: "owner_update", Name: "app3"},
				owner:     "",
				name:      "app1",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "filter miss",
			s:    adapter.CreateSession(context.Background(), nil),
			args: args{
				objectPtr: &Application{Name: "app_update_miss"},
				owner:     "",
				name:      "not existed",
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.UpdateByID(tt.args.objectPtr, tt.args.owner, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.UpdateByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Session.UpdateByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_DeleteByID(t *testing.T) {
	type args struct {
		objectPtr interface{}
		owner     string
		name      string
	}
	tests := []struct {
		name    string
		s       *Session
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "filter",
			s:    adapter.CreateSession(context.Background(), nil),
			args: args{
				objectPtr: &Application{},
				owner:     "",
				name:      "app2",
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.DeleteByID(tt.args.objectPtr, tt.args.owner, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Session.DeleteByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
