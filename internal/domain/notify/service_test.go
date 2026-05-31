package notify

import (
	"testing"

	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&model.Notification{}, &model.NotificationTemplate{})
	return db
}

func TestService_Send(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	notif, err := svc.Send(&SendRequest{
		UserID: 1, Type: "order", Channel: "in_app",
		Title: "Order Created", Content: "Your order has been created",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if notif.IsRead != 0 {
		t.Errorf("is_read = %v, want 0", notif.IsRead)
	}
}

func TestService_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	svc.Send(&SendRequest{UserID: 1, Type: "security", Channel: "in_app", Title: "T2", Content: "C2"})
	svc.Send(&SendRequest{UserID: 2, Type: "order", Channel: "in_app", Title: "T3", Content: "C3"})

	notifs, total, _ := svc.ListByUser(1, "", nil, 1, 20)
	if total != 2 {
		t.Errorf("total = %v, want 2", total)
	}
	if len(notifs) != 2 {
		t.Errorf("len = %v, want 2", len(notifs))
	}
}

func TestService_ListByUser_FilterType(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	svc.Send(&SendRequest{UserID: 1, Type: "security", Channel: "in_app", Title: "T2", Content: "C2"})

	notifs, _, _ := svc.ListByUser(1, "order", nil, 1, 20)
	if len(notifs) != 1 {
		t.Errorf("len = %v, want 1", len(notifs))
	}
}

func TestService_MarkRead(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	notif, _ := svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T", Content: "C"})
	svc.MarkRead(notif.ID, 1)

	isRead := true
	notifs, _, _ := svc.ListByUser(1, "", &isRead, 1, 20)
	if len(notifs) != 1 {
		t.Errorf("len = %v, want 1", len(notifs))
	}
}

func TestService_MarkAllRead(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T2", Content: "C2"})
	svc.MarkAllRead(1)

	isRead := false
	notifs, _, _ := svc.ListByUser(1, "", &isRead, 1, 20)
	if len(notifs) != 0 {
		t.Errorf("len = %v, want 0", len(notifs))
	}
}

func TestService_GetUnreadCount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	svc.Send(&SendRequest{UserID: 1, Type: "order", Channel: "in_app", Title: "T2", Content: "C2"})

	count, _ := svc.GetUnreadCount(1)
	if count != 2 {
		t.Errorf("count = %v, want 2", count)
	}
}

func TestService_SendFromTemplate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.NotificationTemplate{
		Code: "order_created", Name: "Order Created", Channel: "in_app",
		Content: "Your order {{order_no}} has been created", Language: "zh-CN", Status: 1,
	})

	notif, err := svc.SendFromTemplate(1, "order_created", map[string]string{"order_no": "ORD123"})
	if err != nil {
		t.Fatalf("SendFromTemplate() error = %v", err)
	}
	if notif.Content != "Your order ORD123 has been created" {
		t.Errorf("content = %v", notif.Content)
	}
}

func TestService_ListTemplates(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	status := 1
	svc.CreateTemplate(&TemplateRequest{Code: "t1", Name: "T1", Channel: "in_app", Content: "C1", Language: "zh-CN", Status: &status})
	svc.CreateTemplate(&TemplateRequest{Code: "t2", Name: "T2", Channel: "email", Content: "C2", Language: "en", Status: &status})

	templates, err := svc.ListTemplates("", "")
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("len = %v, want 2", len(templates))
	}
}

func TestService_ListTemplates_FilterChannel(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	status := 1
	svc.CreateTemplate(&TemplateRequest{Code: "t1", Name: "T1", Channel: "in_app", Content: "C1", Status: &status})
	svc.CreateTemplate(&TemplateRequest{Code: "t2", Name: "T2", Channel: "email", Content: "C2", Status: &status})

	templates, _ := svc.ListTemplates("email", "")
	if len(templates) != 1 {
		t.Errorf("len = %v, want 1", len(templates))
	}
}

func TestService_ListTemplates_FilterLanguage(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	status := 1
	svc.CreateTemplate(&TemplateRequest{Code: "t1", Name: "T1", Channel: "in_app", Content: "C1", Language: "zh-CN", Status: &status})
	svc.CreateTemplate(&TemplateRequest{Code: "t2", Name: "T2", Channel: "in_app", Content: "C2", Language: "en", Status: &status})

	templates, _ := svc.ListTemplates("", "en")
	if len(templates) != 1 {
		t.Errorf("len = %v, want 1", len(templates))
	}
}

func TestService_CreateTemplate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	status := 1
	tmpl, err := svc.CreateTemplate(&TemplateRequest{
		Code: "welcome", Name: "Welcome", Channel: "email",
		Content: "Hello {{name}}", Language: "en", Status: &status,
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}
	if tmpl.Code != "welcome" {
		t.Errorf("code = %v, want welcome", tmpl.Code)
	}
	if tmpl.Language != "en" {
		t.Errorf("language = %v, want en", tmpl.Language)
	}
	if tmpl.Status != 1 {
		t.Errorf("status = %v, want 1", tmpl.Status)
	}
}

func TestService_CreateTemplate_DefaultLanguage(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	tmpl, err := svc.CreateTemplate(&TemplateRequest{
		Code: "test", Channel: "in_app", Content: "test",
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}
	if tmpl.Language != "zh-CN" {
		t.Errorf("language = %v, want zh-CN", tmpl.Language)
	}
	if tmpl.Status != 1 {
		t.Errorf("status = %v, want 1", tmpl.Status)
	}
}

func TestService_UpdateTemplate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	created, _ := svc.CreateTemplate(&TemplateRequest{
		Code: "old_code", Name: "Old", Channel: "in_app", Content: "Old content",
	})

	status := 0
	updated, err := svc.UpdateTemplate(created.ID, &TemplateRequest{
		Code: "new_code", Name: "New", Channel: "email", Content: "New content", Status: &status,
	})
	if err != nil {
		t.Fatalf("UpdateTemplate() error = %v", err)
	}
	if updated.Code != "new_code" {
		t.Errorf("code = %v, want new_code", updated.Code)
	}
	if updated.Channel != "email" {
		t.Errorf("channel = %v, want email", updated.Channel)
	}
	if updated.Status != 0 {
		t.Errorf("status = %v, want 0", updated.Status)
	}
}

func TestService_UpdateTemplate_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.UpdateTemplate(999, &TemplateRequest{
		Code: "x", Channel: "in_app", Content: "x",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "template not found" {
		t.Errorf("error = %v, want 'template not found'", err)
	}
}
