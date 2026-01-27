// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Backup struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(500)" json:"description"`
	
	// Database connection info
	Host     string `xorm:"varchar(100)" json:"host"`
	Port     int    `json:"port"`
	Database string `xorm:"varchar(100)" json:"database"`
	Username string `xorm:"varchar(100)" json:"username"`
	Password string `xorm:"varchar(500)" json:"password"`
	
	// Backup info
	BackupFile string `xorm:"varchar(500)" json:"backupFile"`
	FileSize   int64  `json:"fileSize"`
	Status     string `xorm:"varchar(50)" json:"status"` // "Created", "InProgress", "Completed", "Failed"
}

func GetMaskedBackup(backup *Backup) *Backup {
	if backup == nil {
		return nil
	}

	// Mask password
	if backup.Password != "" {
		backup.Password = "***"
	}
	return backup
}

func GetMaskedBackups(backups []*Backup, err error) ([]*Backup, error) {
	if err != nil {
		return nil, err
	}

	for _, backup := range backups {
		backup = GetMaskedBackup(backup)
	}
	return backups, nil
}

func GetBackupCount(owner, field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Where("owner = ? or owner = ? ", "admin", owner).Count(&Backup{})
}

func GetBackups(owner string) ([]*Backup, error) {
	backups := []*Backup{}
	db := ormer.Engine.NewSession()
	if owner != "" {
		db = db.Where("owner = ? or owner = ? ", "admin", owner)
	}
	err := db.Desc("created_time").Find(&backups, &Backup{})
	if err != nil {
		return backups, err
	}

	return backups, nil
}

func GetPaginationBackups(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Backup, error) {
	backups := []*Backup{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Where("owner = ? or owner = ? ", "admin", owner).Find(&backups)
	if err != nil {
		return backups, err
	}

	return backups, nil
}

func GetGlobalBackupsCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&Backup{})
}

func GetGlobalBackups() ([]*Backup, error) {
	backups := []*Backup{}
	err := ormer.Engine.Desc("created_time").Find(&backups)
	if err != nil {
		return backups, err
	}

	return backups, nil
}

func GetPaginationGlobalBackups(offset, limit int, field, value, sortField, sortOrder string) ([]*Backup, error) {
	backups := []*Backup{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&backups)
	if err != nil {
		return backups, err
	}

	return backups, nil
}

func getBackup(owner string, name string) (*Backup, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	backup := Backup{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&backup)
	if err != nil {
		return &backup, err
	}

	if existed {
		return &backup, nil
	} else {
		return nil, nil
	}
}

func GetBackup(id string) (*Backup, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	backup, err := getBackup(owner, name)
	if backup == nil && owner != "admin" {
		return getBackup("admin", name)
	} else {
		return backup, err
	}
}

func UpdateBackup(id string, backup *Backup) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if b, err := getBackup(owner, name); err != nil {
		return false, err
	} else if b == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(backup)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddBackup(backup *Backup) (bool, error) {
	affected, err := ormer.Engine.Insert(backup)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteBackup(backup *Backup) (bool, error) {
	// Delete backup file if it exists
	if backup.BackupFile != "" {
		_ = os.Remove(backup.BackupFile)
	}

	affected, err := ormer.Engine.ID(core.PK{backup.Owner, backup.Name}).Delete(&Backup{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (b *Backup) GetId() string {
	return fmt.Sprintf("%s/%s", b.Owner, b.Name)
}

// ExecuteBackup performs a database backup using mysqldump
func (b *Backup) ExecuteBackup() error {
	// Update status to InProgress
	b.Status = "InProgress"
	_, err := UpdateBackup(b.GetId(), b)
	if err != nil {
		return err
	}

	// Determine backup file path
	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		b.Status = "Failed"
		_, _ = UpdateBackup(b.GetId(), b)
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	backupFilePath := fmt.Sprintf("%s/%s_%s_%s.sql", backupDir, b.Owner, b.Name, util.GetCurrentTime())
	b.BackupFile = backupFilePath

	// Build mysqldump command
	host := b.Host
	if host == "" {
		host = conf.GetConfigString("dataSourceName")
		// Extract host from dataSourceName if using default
		parts := strings.Split(host, "@tcp(")
		if len(parts) == 2 {
			hostPort := strings.Split(parts[1], ")/")
			if len(hostPort) == 2 {
				host = strings.Split(hostPort[0], ":")[0]
			}
		} else {
			host = "localhost"
		}
	}

	port := b.Port
	if port == 0 {
		port = 3306
	}

	database := b.Database
	if database == "" {
		// Extract database from dataSourceName
		dataSource := conf.GetConfigString("dataSourceName")
		parts := strings.Split(dataSource, ")/")
		if len(parts) == 2 {
			dbParams := strings.Split(parts[1], "?")
			database = dbParams[0]
		} else {
			database = "casdoor"
		}
	}

	username := b.Username
	password := b.Password

	// Execute mysqldump command
	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%d", port),
		fmt.Sprintf("--user=%s", username),
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
		database,
	}

	if password != "" {
		args = append([]string{fmt.Sprintf("--password=%s", password)}, args...)
	}

	cmd := exec.Command("mysqldump", args...)
	
	outFile, err := os.Create(backupFilePath)
	if err != nil {
		b.Status = "Failed"
		_, _ = UpdateBackup(b.GetId(), b)
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		b.Status = "Failed"
		_, _ = UpdateBackup(b.GetId(), b)
		return fmt.Errorf("mysqldump failed: %v, stderr: %s", err, stderr.String())
	}

	// Get file size
	fileInfo, err := os.Stat(backupFilePath)
	if err == nil {
		b.FileSize = fileInfo.Size()
	}

	// Update status to Completed
	b.Status = "Completed"
	_, err = UpdateBackup(b.GetId(), b)
	if err != nil {
		return err
	}

	return nil
}

// RestoreBackup restores a database from a backup file
func (b *Backup) RestoreBackup() error {
	if b.BackupFile == "" || b.Status != "Completed" {
		return fmt.Errorf("no valid backup file to restore")
	}

	// Check if backup file exists
	if _, err := os.Stat(b.BackupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", b.BackupFile)
	}

	// Build mysql restore command
	host := b.Host
	if host == "" {
		host = conf.GetConfigString("dataSourceName")
		parts := strings.Split(host, "@tcp(")
		if len(parts) == 2 {
			hostPort := strings.Split(parts[1], ")/")
			if len(hostPort) == 2 {
				host = strings.Split(hostPort[0], ":")[0]
			}
		} else {
			host = "localhost"
		}
	}

	port := b.Port
	if port == 0 {
		port = 3306
	}

	database := b.Database
	if database == "" {
		dataSource := conf.GetConfigString("dataSourceName")
		parts := strings.Split(dataSource, ")/")
		if len(parts) == 2 {
			dbParams := strings.Split(parts[1], "?")
			database = dbParams[0]
		} else {
			database = "casdoor"
		}
	}

	username := b.Username
	password := b.Password

	// Execute mysql restore command
	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%d", port),
		fmt.Sprintf("--user=%s", username),
		database,
	}

	if password != "" {
		args = append([]string{fmt.Sprintf("--password=%s", password)}, args...)
	}

	cmd := exec.Command("mysql", args...)
	
	inFile, err := os.Open(b.BackupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer inFile.Close()

	cmd.Stdin = inFile
	
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mysql restore failed: %v, stderr: %s", err, stderr.String())
	}

	return nil
}
