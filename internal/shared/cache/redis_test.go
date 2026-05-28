package cache

import (
	"testing"
)

func TestUserSessionKey(t *testing.T) {
	if got := UserSessionKey(42); got != "user:session:42" {
		t.Errorf("UserSessionKey(42) = %v, want user:session:42", got)
	}
}

func TestUserQuotaKey(t *testing.T) {
	if got := UserQuotaKey(1); got != "user:quota:1" {
		t.Errorf("UserQuotaKey(1) = %v, want user:quota:1", got)
	}
}

func TestTokenProductKey(t *testing.T) {
	if got := TokenProductKey(5); got != "token:product:5" {
		t.Errorf("TokenProductKey(5) = %v, want token:product:5", got)
	}
}

func TestSupplierKey(t *testing.T) {
	if got := SupplierKey(3); got != "token:supplier:3" {
		t.Errorf("SupplierKey(3) = %v, want token:supplier:3", got)
	}
}

func TestRouteHealthKey(t *testing.T) {
	if got := RouteHealthKey(10); got != "route:health:10" {
		t.Errorf("RouteHealthKey(10) = %v, want route:health:10", got)
	}
}

func TestRateLimitKey(t *testing.T) {
	if got := RateLimitKey("192.168.1.1"); got != "rate:limit:192.168.1.1" {
		t.Errorf("RateLimitKey() = %v, want rate:limit:192.168.1.1", got)
	}
}

func TestVerifyCodeKey(t *testing.T) {
	if got := VerifyCodeKey("user@test.com"); got != "verify:code:user@test.com" {
		t.Errorf("VerifyCodeKey() = %v, want verify:code:user@test.com", got)
	}
}

func TestRefreshTokenKey(t *testing.T) {
	if got := RefreshTokenKey("abc123"); got != "refresh:token:abc123" {
		t.Errorf("RefreshTokenKey() = %v, want refresh:token:abc123", got)
	}
}

func TestI18nKey(t *testing.T) {
	if got := I18nKey("zh-CN", "common"); got != "i18n:translations:zh-CN:common" {
		t.Errorf("I18nKey() = %v, want i18n:translations:zh-CN:common", got)
	}
}

func TestSystemConfigKey(t *testing.T) {
	if got := SystemConfigKey("site_name"); got != "config:system:site_name" {
		t.Errorf("SystemConfigKey() = %v, want config:system:site_name", got)
	}
}
