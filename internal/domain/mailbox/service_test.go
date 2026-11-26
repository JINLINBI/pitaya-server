package mailbox

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testService(t *testing.T) *Service {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := gdb.AutoMigrate(&Message{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewService(NewRepository(gdb))
}

func TestSendListMarkRead(t *testing.T) {
	svc := testService(t)
	ctx := context.Background()

	m, err := svc.Send(ctx, SendInput{PlayerID: "1", Title: "欢迎", Body: "你好，冒险者"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if m.ID == 0 || m.Read {
		t.Fatalf("unexpected message: %+v", m)
	}

	list, err := svc.List(ctx, "1", 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v len=%d", err, len(list))
	}

	if err := svc.MarkRead(ctx, "1", m.ID); err != nil {
		t.Fatalf("mark: %v", err)
	}
	if err := svc.MarkRead(ctx, "2", m.ID); err != ErrNotFound {
		t.Fatalf("other player should not mark, got %v", err)
	}
}

func TestSendValidation(t *testing.T) {
	svc := testService(t)
	ctx := context.Background()
	if _, err := svc.Send(ctx, SendInput{PlayerID: "", Title: "t", Body: "b"}); err != ErrEmptyPlayer {
		t.Fatalf("want empty player, got %v", err)
	}
	if _, err := svc.Send(ctx, SendInput{PlayerID: "1", Title: "", Body: "b"}); err != ErrInvalidTitle {
		t.Fatalf("want title err, got %v", err)
	}
}
