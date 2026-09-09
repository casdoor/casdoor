package object

import (
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
)

type LdapAutoSynchronizer struct {
	sync.Mutex
	ldapIdToStopChan map[string]chan struct{}
}

var globalLdapAutoSynchronizer *LdapAutoSynchronizer

func InitLdapAutoSynchronizer() {
	globalLdapAutoSynchronizer = NewLdapAutoSynchronizer()
	err := globalLdapAutoSynchronizer.LdapAutoSynchronizerStartUpAll()
	if err != nil {
		panic(err)
	}
}

func NewLdapAutoSynchronizer() *LdapAutoSynchronizer {
	return &LdapAutoSynchronizer{
		ldapIdToStopChan: make(map[string]chan struct{}),
	}
}

func GetLdapAutoSynchronizer() *LdapAutoSynchronizer {
	return globalLdapAutoSynchronizer
}

// StartAutoSync
// start autosync for specified ldap, old existing autosync goroutine will be ceased
func (l *LdapAutoSynchronizer) StartAutoSync(ldapId string) error {
	l.Lock()
	defer l.Unlock()

	ldap, err := GetLdap(ldapId)
	if err != nil {
		return err
	}

	if ldap == nil {
		return fmt.Errorf("ldap %s doesn't exist", ldapId)
	}
	l.stopAutoSync(ldapId)

	stopChan := make(chan struct{})
	l.ldapIdToStopChan[ldapId] = stopChan
	logs.Info(fmt.Sprintf("autoSync started for %s", ldap.Id))
	util.SafeGoroutine(func() {
		l.syncRoutine(ldap, stopChan)
	})
	return nil
}

func (l *LdapAutoSynchronizer) StopAutoSync(ldapId string) {
	l.Lock()
	defer l.Unlock()
	l.stopAutoSync(ldapId)
}

// stopAutoSync signals the running goroutine to quit, the caller must hold the lock.
// The channel is closed instead of being sent to, so that a goroutine which already
// died doesn't block the caller forever.
func (l *LdapAutoSynchronizer) stopAutoSync(ldapId string) {
	if stopChan, ok := l.ldapIdToStopChan[ldapId]; ok {
		close(stopChan)
		delete(l.ldapIdToStopChan, ldapId)
	}
}

// autosync goroutine
func (l *LdapAutoSynchronizer) syncRoutine(ldap *Ldap, stopChan chan struct{}) {
	ticker := time.NewTicker(time.Duration(ldap.AutoSync) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			logs.Info(fmt.Sprintf("autoSync goroutine for %s stopped", ldap.Id))
			return
		case <-ticker.C:
		}

		l.syncOnce(ldap)
	}
}

// syncOnce runs one sync cycle. Every failure, panics included, is confined to the
// cycle: a long sync can outlive the LDAP server's connection timeout, and killing
// the goroutine for that would stop the periodic sync forever.
func (l *LdapAutoSynchronizer) syncOnce(ldap *Ldap) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error(fmt.Sprintf("autoSync panicked for %s, error %v, retrying at the next cycle", ldap.Id, r))
		}
	}()

	claimed, err := claimSyncCycle(ldap.Id, ldap.AutoSync)
	if err != nil {
		logs.Warning(fmt.Sprintf("autoSync failed to claim the sync cycle for %s, error %s", ldap.Id, err))
		return
	}
	if !claimed {
		logs.Info(fmt.Sprintf("autoSync skipped for %s, another instance is running this cycle", ldap.Id))
		return
	}

	// fetch all users and groups
	conn, err := ldap.GetLdapConn()
	if err != nil {
		logs.Warning(fmt.Sprintf("autoSync failed for %s, error %s", ldap.Id, err))
		return
	}
	defer conn.Close()

	// Sync groups first if enabled (so they exist before assigning users)
	if ldap.EnableGroups {
		groups, err := conn.GetLdapGroups(ldap)
		if err != nil {
			logs.Warning(fmt.Sprintf("autoSync failed to fetch groups for %s, error %s", ldap.Id, err))
		} else {
			newGroups, updatedGroups, err := SyncLdapGroups(ldap.Owner, groups, ldap.Id)
			if err != nil {
				logs.Warning(fmt.Sprintf("autoSync failed to sync groups for %s, error %s", ldap.Id, err))
			} else {
				logs.Info(fmt.Sprintf("ldap group sync success for %s, %d new groups, %d updated groups", ldap.Id, newGroups, updatedGroups))
			}
		}
	}

	users, err := conn.GetLdapUsers(ldap)
	if err != nil {
		logs.Warning(fmt.Sprintf("autoSync failed for %s, error %s", ldap.Id, err))
		return
	}

	existed, failed, err := SyncLdapUsers(ldap.Owner, AutoAdjustLdapUser(users), ldap.Id)
	if err != nil {
		logs.Warning(fmt.Sprintf("autoSync failed for %s, error %s", ldap.Id, err))
		return
	}

	if len(failed) != 0 {
		logs.Warning(fmt.Sprintf("ldap autosync,%d new users,but %d user failed during :", len(users)-len(existed)-len(failed), len(failed)), failed)
	} else {
		logs.Info(fmt.Sprintf("ldap autosync success, %d new users, %d existing users", len(users)-len(existed), len(existed)))
	}
}

// LdapAutoSynchronizerStartUpAll
// start all autosync goroutine for existing ldap servers in each organizations
func (l *LdapAutoSynchronizer) LdapAutoSynchronizerStartUpAll() error {
	organizations := []*Organization{}
	err := ormer.Engine.Desc("created_time").Find(&organizations)
	if err != nil {
		logs.Info("failed to Star up LdapAutoSynchronizer; ")
	}
	for _, org := range organizations {
		ldaps, err := GetLdaps(org.Name)
		if err != nil {
			return err
		}

		for _, ldap := range ldaps {
			if ldap.AutoSync != 0 {
				err = l.StartAutoSync(ldap.Id)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// claimSyncCycle reserves the cycle for this instance, so that a multi-node deployment
// doesn't sync the same LDAP from every node at once. It is a compare-and-swap on last_sync:
// only the node whose UPDATE still matches the value it just read owns the cycle.
func claimSyncCycle(ldapId string, autoSync int) (bool, error) {
	ldap, err := GetLdap(ldapId)
	if err != nil {
		return false, err
	}
	if ldap == nil {
		return false, fmt.Errorf("ldap %s doesn't exist", ldapId)
	}

	if !isSyncDue(ldap.LastSync, autoSync) {
		return false, nil
	}

	session := ormer.Engine.ID(ldapId)
	if ldap.LastSync == "" {
		session = session.Where("last_sync = '' or last_sync is null")
	} else {
		session = session.Where("last_sync = ?", ldap.LastSync)
	}

	affected, err := session.Cols("last_sync").Update(&Ldap{LastSync: util.GetCurrentTime()})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// isSyncDue reports whether a whole interval has passed since the last sync. The tolerance
// absorbs the drift between the tickers of different nodes.
func isSyncDue(lastSync string, autoSync int) bool {
	if lastSync == "" {
		return true
	}

	lastSyncTime, err := time.Parse(time.RFC3339, lastSync)
	if err != nil {
		return true
	}

	interval := time.Duration(autoSync) * time.Minute
	return time.Since(lastSyncTime) >= interval-interval/10
}

func UpdateLdapSyncTime(ldapId string) error {
	_, err := ormer.Engine.ID(ldapId).Update(&Ldap{LastSync: util.GetCurrentTime()})
	if err != nil {
		return err
	}

	return nil
}
