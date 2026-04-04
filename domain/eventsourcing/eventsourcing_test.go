package eventsourcing

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// --- 测试用聚合：银行账户 ---

// BankAccount 银行账户聚合.
type BankAccount struct {
	BaseAggregate
	Balance int64  `json:"balance"`
	Owner   string `json:"owner"`
}

// NewBankAccount 创建银行账户.
func NewBankAccount(id, owner string) *BankAccount {
	return &BankAccount{
		BaseAggregate: NewBaseAggregate(id, "BankAccount"),
		Owner:         owner,
	}
}

// ApplyEvent 应用事件.
func (a *BankAccount) ApplyEvent(event Event) error {
	switch event.EventType {
	case "AccountCreated":
		var data struct {
			Owner string `json:"owner"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}
		a.Owner = data.Owner
	case "Deposited":
		var data struct {
			Amount int64 `json:"amount"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}
		a.Balance += data.Amount
	case "Withdrawn":
		var data struct {
			Amount int64 `json:"amount"`
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			return err
		}
		if a.Balance < data.Amount {
			return errors.New("余额不足")
		}
		a.Balance -= data.Amount
	}
	return nil
}

// Deposit 存款.
func (a *BankAccount) Deposit(amount int64) error {
	return a.RaiseEvent(a.ApplyEvent, "Deposited", map[string]int64{"amount": amount})
}

// Withdraw 取款.
func (a *BankAccount) Withdraw(amount int64) error {
	return a.RaiseEvent(a.ApplyEvent, "Withdrawn", map[string]int64{"amount": amount})
}

// Create 创建账户事件.
func (a *BankAccount) Create(owner string) error {
	return a.RaiseEvent(a.ApplyEvent, "AccountCreated", map[string]string{"owner": owner})
}

// --- helpers ---

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	return db
}

func setupEventStore(t *testing.T) (*GORMEventStore, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	store := NewGORMEventStore(db)
	require.NoError(t, store.AutoMigrate())
	return store, db
}

func setupSnapshotStore(t *testing.T, db *gorm.DB) *GORMSnapshotStore {
	t.Helper()
	store := NewGORMSnapshotStore(db)
	require.NoError(t, store.AutoMigrate())
	return store
}

// --- BaseAggregate 测试 ---

func TestBaseAggregate_RaiseEvent(t *testing.T) {
	account := NewBankAccount("acc-1", "张三")

	err := account.Deposit(100)
	require.NoError(t, err)

	assert.Equal(t, int64(100), account.Balance)
	assert.Equal(t, int64(1), account.Version())
	assert.Len(t, account.UncommittedEvents(), 1)

	event := account.UncommittedEvents()[0]
	assert.Equal(t, "acc-1", event.AggregateID)
	assert.Equal(t, "BankAccount", event.AggregateType)
	assert.Equal(t, int64(1), event.Version)
	assert.Equal(t, "Deposited", event.EventType)
	assert.NotEmpty(t, event.ID)

	// 多次操作
	err = account.Deposit(50)
	require.NoError(t, err)
	err = account.Withdraw(30)
	require.NoError(t, err)

	assert.Equal(t, int64(120), account.Balance)
	assert.Equal(t, int64(3), account.Version())
	assert.Len(t, account.UncommittedEvents(), 3)

	// 清除事件
	account.ClearUncommittedEvents()
	assert.Empty(t, account.UncommittedEvents())
}

func TestBaseAggregate_RaiseEvent_ApplyError(t *testing.T) {
	account := NewBankAccount("acc-1", "张三")

	// 余额不足时 ApplyEvent 返回错误
	err := account.Withdraw(100)
	assert.Error(t, err)

	// 版本不变，无未提交事件
	assert.Equal(t, int64(0), account.Version())
	assert.Empty(t, account.UncommittedEvents())
}

// --- GORMEventStore 测试 ---

func TestGORMEventStore_SaveAndLoad(t *testing.T) {
	store, _ := setupEventStore(t)
	ctx := t.Context()

	account := NewBankAccount("acc-1", "张三")
	require.NoError(t, account.Create("张三"))
	require.NoError(t, account.Deposit(100))
	require.NoError(t, account.Deposit(50))

	// 保存事件
	err := store.Save(ctx, account.UncommittedEvents())
	require.NoError(t, err)

	// LoadAll
	events, err := store.LoadAll(ctx, "acc-1")
	require.NoError(t, err)
	assert.Len(t, events, 3)
	assert.Equal(t, int64(1), events[0].Version)
	assert.Equal(t, int64(2), events[1].Version)
	assert.Equal(t, int64(3), events[2].Version)

	// Load（从版本 1 之后）
	events, err = store.Load(ctx, "acc-1", 1)
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, int64(2), events[0].Version)

	// Load 不存在的聚合
	events, err = store.Load(ctx, "not-exists", 0)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestGORMEventStore_Save_Empty(t *testing.T) {
	store, _ := setupEventStore(t)
	err := store.Save(t.Context(), nil)
	assert.ErrorIs(t, err, ErrNoEvents)
}

// --- GORMSnapshotStore 测试 ---

func TestGORMSnapshotStore_SaveAndLoad(t *testing.T) {
	db := setupTestDB(t)
	store := setupSnapshotStore(t, db)
	ctx := t.Context()

	snapshot := Snapshot{
		AggregateID:   "acc-1",
		AggregateType: "BankAccount",
		Version:       5,
		Data:          json.RawMessage(`{"balance":500,"owner":"张三"}`),
	}

	// 保存快照
	err := store.Save(ctx, snapshot)
	require.NoError(t, err)

	// 加载快照
	loaded, err := store.Load(ctx, "acc-1")
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, int64(5), loaded.Version)
	assert.Equal(t, "BankAccount", loaded.AggregateType)

	// Upsert：更新快照
	snapshot.Version = 10
	snapshot.Data = json.RawMessage(`{"balance":1000,"owner":"张三"}`)
	err = store.Save(ctx, snapshot)
	require.NoError(t, err)

	loaded, err = store.Load(ctx, "acc-1")
	require.NoError(t, err)
	assert.Equal(t, int64(10), loaded.Version)

	// 加载不存在的快照
	loaded, err = store.Load(ctx, "not-exists")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

// --- Repository 测试（无快照） ---

func TestRepository_SaveAndLoad(t *testing.T) {
	store, _ := setupEventStore(t)
	ctx := t.Context()

	factory := func() *BankAccount {
		return &BankAccount{BaseAggregate: NewBaseAggregate("", "BankAccount")}
	}

	repo, err := NewRepository[*BankAccount](store, factory)
	require.NoError(t, err)

	// 创建并保存聚合
	account := NewBankAccount("acc-1", "")
	require.NoError(t, account.Create("张三"))
	require.NoError(t, account.Deposit(100))
	require.NoError(t, account.Deposit(50))

	err = repo.Save(ctx, account)
	require.NoError(t, err)
	assert.Empty(t, account.UncommittedEvents())

	// 加载聚合
	loaded, err := repo.Load(ctx, "acc-1")
	require.NoError(t, err)
	assert.Equal(t, "张三", loaded.Owner)
	assert.Equal(t, int64(150), loaded.Balance)
	assert.Equal(t, int64(3), loaded.Version())

	// 加载不存在的聚合
	_, err = repo.Load(ctx, "not-exists")
	assert.ErrorIs(t, err, ErrAggregateNotFound)
}

// --- Repository 测试（带快照） ---

func TestRepository_WithSnapshot(t *testing.T) {
	db := setupTestDB(t)
	eventStore := NewGORMEventStore(db)
	require.NoError(t, eventStore.AutoMigrate())
	snapshotStore := setupSnapshotStore(t, db)
	ctx := t.Context()

	factory := func() *BankAccount {
		return &BankAccount{BaseAggregate: NewBaseAggregate("", "BankAccount")}
	}

	repo, err := NewRepository[*BankAccount](eventStore, factory,
		WithSnapshotStore[*BankAccount](snapshotStore),
		WithSnapshotEvery[*BankAccount](2),
	)
	require.NoError(t, err)

	// 创建聚合，产生 2 个事件 → 触发快照
	account := NewBankAccount("acc-1", "")
	require.NoError(t, account.Create("张三"))
	require.NoError(t, account.Deposit(100))

	err = repo.Save(ctx, account)
	require.NoError(t, err)

	// 验证快照已保存
	snapshot, err := snapshotStore.Load(ctx, "acc-1")
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Equal(t, int64(2), snapshot.Version)

	// 再追加一个事件（版本 3）
	require.NoError(t, account.Deposit(50))
	err = repo.Save(ctx, account)
	require.NoError(t, err)

	// 加载聚合：应从快照版本 2 开始，只需加载版本 3
	loaded, err := repo.Load(ctx, "acc-1")
	require.NoError(t, err)
	assert.Equal(t, int64(150), loaded.Balance)
	assert.Equal(t, int64(3), loaded.Version())
	assert.Equal(t, "张三", loaded.Owner)
}

// --- 并发冲突测试 ---

func TestConcurrencyConflict(t *testing.T) {
	store, _ := setupEventStore(t)
	ctx := t.Context()

	// 第一次保存
	account1 := NewBankAccount("acc-1", "")
	require.NoError(t, account1.Create("张三"))
	err := store.Save(ctx, account1.UncommittedEvents())
	require.NoError(t, err)

	// 模拟并发：用相同的 aggregate_id + version 再次保存
	account2 := NewBankAccount("acc-1", "")
	require.NoError(t, account2.Create("李四"))
	err = store.Save(ctx, account2.UncommittedEvents())
	assert.ErrorIs(t, err, ErrConcurrencyConflict)
}

// --- NewRepository 校验测试 ---

func TestNewRepository_NilEventStore(t *testing.T) {
	factory := func() *BankAccount { return &BankAccount{} }
	_, err := NewRepository[*BankAccount](nil, factory)
	assert.ErrorIs(t, err, ErrNilEventStore)
}

func TestNewRepository_NilFactory(t *testing.T) {
	store, _ := setupEventStore(t)
	_, err := NewRepository[*BankAccount](store, nil)
	assert.ErrorIs(t, err, ErrNilFactory)
}

// --- Save 无事件测试 ---

func TestRepository_Save_NoEvents(t *testing.T) {
	store, _ := setupEventStore(t)
	factory := func() *BankAccount {
		return &BankAccount{BaseAggregate: NewBaseAggregate("", "BankAccount")}
	}
	repo, err := NewRepository[*BankAccount](store, factory)
	require.NoError(t, err)

	account := NewBankAccount("acc-1", "张三")
	err = repo.Save(t.Context(), account)
	assert.ErrorIs(t, err, ErrNoEvents)
}
