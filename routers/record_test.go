package routers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func newRecordTestContext(method, path, body, contentType string) *context.Context {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	ctx := context.NewContext()
	ctx.Reset(httptest.NewRecorder(), req)
	return ctx
}

func TestGetSetPasswordTargetUser(t *testing.T) {
	t.Parallel()

	ctx := newRecordTestContext(
		http.MethodPost,
		"/api/set-password",
		"userOwner=hyperion&userName=user_kk64gr",
		"application/x-www-form-urlencoded",
	)

	if got := getSetPasswordTargetUser(ctx); got != "hyperion/user_kk64gr" {
		t.Fatalf("getSetPasswordTargetUser() = %q, want hyperion/user_kk64gr", got)
	}
}

func TestGetSetPasswordTargetUser_MissingParams(t *testing.T) {
	t.Parallel()

	ctx := newRecordTestContext(
		http.MethodPost,
		"/api/set-password",
		"userOwner=hyperion",
		"application/x-www-form-urlencoded",
	)

	if got := getSetPasswordTargetUser(ctx); got != "" {
		t.Fatalf("getSetPasswordTargetUser() = %q, want empty", got)
	}
}

func TestGetVerificationCodeTargetUser_MissingParams(t *testing.T) {
	t.Parallel()

	ctx := newRecordTestContext(
		http.MethodPost,
		"/api/send-verification-code",
		"checkUser=user_kk64gr",
		"application/x-www-form-urlencoded",
	)

	if got := getVerificationCodeTargetUser(ctx); got != "" {
		t.Fatalf("getVerificationCodeTargetUser() = %q, want empty", got)
	}
}

func TestRecordMessage_OtherEndpointDoesNotSetTarget(t *testing.T) {
	t.Parallel()

	ctx := newRecordTestContext(http.MethodGet, "/api/get-user", "", "")

	RecordMessage(ctx)

	if target := ctx.Input.Params()["recordTargetUserId"]; target != "" {
		t.Fatalf("recordTargetUserId = %q, want empty", target)
	}
}

func TestRecordMessage_SetPasswordSetsTarget(t *testing.T) {
	t.Parallel()

	ctx := newRecordTestContext(
		http.MethodPost,
		"/api/set-password",
		"userOwner=hyperion&userName=user_kk64gr",
		"application/x-www-form-urlencoded",
	)

	RecordMessage(ctx)

	if target := ctx.Input.Params()["recordTargetUserId"]; target != "hyperion/user_kk64gr" {
		t.Fatalf("recordTargetUserId = %q, want hyperion/user_kk64gr", target)
	}
}

func TestRecordMessage_LoginSkipped(t *testing.T) {
	t.Parallel()

	ctx := newRecordTestContext(http.MethodPost, "/api/login", "", "")

	RecordMessage(ctx)

	if _, ok := ctx.Input.Params()["recordUserId"]; ok {
		t.Fatal("recordUserId should not be set for /api/login")
	}
}
