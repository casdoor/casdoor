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

	err := UpdateLdapSyncTime(ldap.Id)
	if err != nil {
		logs.Warning(fmt.Sprintf("autoSync failed to update the sync time for %s, error %s", ldap.Id, err))
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

func UpdateLdapSyncTime(ldapId string) error {
	_, err := ormer.Engine.ID(ldapId).Update(&Ldap{LastSync: util.GetCurrentTime()})
	if err != nil {
		return err
	}

	return nil
}
