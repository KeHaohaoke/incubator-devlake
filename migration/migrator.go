package migration

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"sort"
	"sync"
)

var m = migrator{scripts: make(map[string]scriptWithComment)}

type scriptWithComment struct {
	Script
	comment string
}
type migrator struct {
	sync.Mutex
	db      *gorm.DB
	scripts map[string]scriptWithComment
}

func Init(db *gorm.DB) {
	m.db = db
}

func (m *migrator) register(scripts []Script, comment string) {
	m.Lock()
	defer m.Unlock()
	for _, script := range scripts {
		m.scripts[fmt.Sprintf("%s:%d", script.Name(), script.Version())] = scriptWithComment{
			Script:  script,
			comment: comment,
		}
	}
}

func (m *migrator) bookKeep(script scriptWithComment) error {
	record := &MigrationHistory{
		ScriptVersion: script.Version(),
		ScriptName:    script.Name(),
		Comment:       script.comment,
	}
	return m.db.Create(record).Error
}

func (m *migrator) execute(ctx context.Context) error {
	versions, err := m.getExecuted()
	if err != nil {
		return err
	}
	for key := range versions {
		delete(m.scripts, key)
	}
	var scriptSlice []scriptWithComment
	for _, script := range m.scripts {
		scriptSlice = append(scriptSlice, script)
	}
	sort.Slice(scriptSlice, func(i, j int) bool {
		return scriptSlice[i].Version() < scriptSlice[j].Version()
	})
	for _, script := range scriptSlice {
		err = script.Up(ctx, m.db)
		if err != nil {
			return err
		}
		err = m.bookKeep(script)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *migrator) getExecuted() (map[string]struct{}, error) {
	var err error
	versions := make(map[string]struct{})
	err = m.db.Migrator().AutoMigrate(&MigrationHistory{})
	if err != nil {
		return nil, err
	}
	var records []MigrationHistory
	err = m.db.Find(&records).Error
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		versions[fmt.Sprintf("%s:%d", record.ScriptName, record.ScriptVersion)] = struct{}{}
	}
	return versions, nil
}

func Register(scripts []Script, comment string) {
	m.register(scripts, comment)
}

func Execute(ctx context.Context) error {
	return m.execute(ctx)
}
