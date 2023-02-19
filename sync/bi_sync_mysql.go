package sync

import (
	"bytes"
	"fmt"
	"github.com/2tvenom/myreplication"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/spf13/viper"
	"log"
	"xorm.io/core"
)

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Base     string `mapstructure:"base"`
}

type Config struct {
	Databases []DBConfig `mapstructure:"mysql"`
}

var (
	dbTables   = make(map[string]*core.Table)
	db1        DBConfig
	db2        DBConfig
	ConfigData *Config
)

func initConfig() {

	// set config file
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./config/mysql.yaml")

	// read config
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}
	db1 = config.Databases[0]
	db2 = config.Databases[1]
	log.Println("init config success……")
}

func StartSync() {
	initConfig()
	engine2, err := CreateEngine(&db2)

	if err != nil {
		panic("create engine fail." + err.Error())
	}
	StartSyncBinlog(&db1, engine2, 3)
}

func StartSyncBinlog(dbSource *DBConfig, targetEngine *xorm.Engine, serverId uint32) {
	newConnection := myreplication.NewConnection()
	err := newConnection.ConnectAndAuth(dbSource.Host, dbSource.Port, dbSource.Username, dbSource.Password)

	if err != nil {
		panic("Client not connected and not autentificate to master server with error:" + err.Error())
	}
	//Get position and file name
	pos, filename, err := newConnection.GetMasterStatus()

	if err != nil {
		panic("Master status fail: " + err.Error())
	}

	el, err := newConnection.StartBinlogDump(pos, filename, serverId)

	if err != nil {
		panic("Cant start bin log: " + err.Error())
	}

	events := el.GetEventChan()
	go func() {
		for {
			event := <-events
			//fmt.Println(event)
			switch e := event.(type) {
			case *myreplication.QueryEvent:
				//Output query event
				targetEngine.DB().Exec(e.GetQuery())
				//db.Exec(e.GetQuery())
				log.Println(e.GetQuery())
				updateTable(targetEngine)
			case *myreplication.IntVarEvent:
				//Output last insert_id  if statement based replication
				println(e.GetValue())
			case *myreplication.WriteEvent:
				//Output Write (insert) event
				println("Write", e.GetTable())
				//Rows loop
				columnNames := GetColumns(dbTables[e.GetTable()].Columns())
				for _, row := range e.GetRows() {
					//Columns loop
					columnVals := make([]string, len(row))
					for j, col := range row {
						//Output row number, column number, column type and column value
						//println(fmt.Sprintf("%d %d %d %v %d", i, j, col.GetType(), col.GetValue(), col.GetColumnId()))
						if IsChar(col.GetType()) {
							columnVals[j] = fmt.Sprintf("'%v'", col.GetValue())
						} else {
							columnVals[j] = fmt.Sprintf("%v", col.GetValue())
						}
					}
					strSql := GetInsertSql(e.GetSchema(), e.GetTable(), columnNames, columnVals)
					targetEngine.DB().Exec(strSql)
					log.Println(strSql)
				}
			case *myreplication.DeleteEvent:
				//Output delete event
				println("Delete", e.GetTable())
				//db.Query("")
				columnNames := GetColumns(dbTables[e.GetTable()].Columns())
				for _, row := range e.GetRows() {
					//Columns loop
					columnVals := make([]string, len(row))
					for j, col := range row {
						if IsChar(col.GetType()) {
							columnVals[j] = fmt.Sprintf("'%v'", col.GetValue())
						} else {
							columnVals[j] = fmt.Sprintf("%v", col.GetValue())
						}
					}
					strSql := GetdeleteSql(e.GetSchema(), e.GetTable(), columnNames, columnVals)
					targetEngine.DB().Exec(strSql)
					log.Println(strSql)
				}

			case *myreplication.UpdateEvent:
				//Output update event
				println("Update", e.GetTable())
				columnNames := GetColumns(dbTables[e.GetTable()].Columns())
				//Output new
				newColumnValList := make([][]string, len(e.GetNewRows()))
				for i, row := range e.GetNewRows() {
					//Columns loop
					newColumnValList[i] = make([]string, len(row))
					for j, col := range row {
						if IsChar(col.GetType()) {
							newColumnValList[i][j] = fmt.Sprintf("'%v'", col.GetValue())
							//fmt.Sprintf("'%v'", col.GetValue())
						} else {
							newColumnValList[i][j] = fmt.Sprintf("%v", col.GetValue())
							//fmt.Sprintf("%v", col.GetValue())
						}
					}
				}
				strSql := GetUpdateSql(e.GetSchema(), e.GetTable(), columnNames, newColumnValList)
				targetEngine.DB().Exec(strSql)
				log.Println(strSql)
			default:
			}
		}
	}()
	err = el.Start()
	println(err.Error())
}

func updateTable(engine *xorm.Engine) error {
	tbs, err := engine.DBMetas()
	if err != nil {
		return err
	}

	for i, tb := range tbs {
		fmt.Println("index:", i, "tbName", tb.Name)
		dbTables[tb.Name] = tb
	}
	return nil
}

func CreateEngine(db *DBConfig) (*xorm.Engine, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", db.Username, db.Password, db.Host, db.Port, db.Database, db.Base)
	engine, err := xorm.NewEngine("mysql", dataSourceName)
	//defer engine.Close()

	if err != nil {
		log.Fatal("connection mysql fail……")
	}
	err = engine.Ping()
	if err != nil {
		panic(err)
		return nil, err
	}
	log.Println("mysql connection success……")
	return engine, nil
}

func IsChar(mysqlType uint8) bool {
	if mysqlType == myreplication.MYSQL_TYPE_DECIMAL ||
		mysqlType == myreplication.MYSQL_TYPE_TINY ||
		mysqlType == myreplication.MYSQL_TYPE_SHORT ||
		mysqlType == myreplication.MYSQL_TYPE_LONG ||
		mysqlType == myreplication.MYSQL_TYPE_FLOAT ||
		mysqlType == myreplication.MYSQL_TYPE_DOUBLE ||
		mysqlType == myreplication.MYSQL_TYPE_LONGLONG ||
		mysqlType == myreplication.MYSQL_TYPE_INT24 {
		return false
	}
	return true
}

func GetColumns(cols []*core.Column) []string {
	columns := make([]string, len(cols))
	for i, col := range cols {
		columns[i] = col.Name
	}
	return columns
}

func GetUpdateSql(schemaName string, tableName string, columnNames []string, newColumnValList [][]string) string {
	var bt bytes.Buffer
	bt.WriteString("replace into " + schemaName + "." + tableName + " (")
	for i, columnName := range columnNames {
		if i+1 < len(columnNames) {
			bt.WriteString(columnName + ", ")
		} else {
			bt.WriteString(columnName + ") values ")
		}
	}

	for i, row := range newColumnValList {
		bt.WriteString("(")
		for j, item := range row {
			if j+1 < len(row) {
				bt.WriteString(item + ",")
			} else {
				bt.WriteString(item + ")")
			}
		}
		if i+1 < len(newColumnValList) {
			bt.WriteString("),")
		} else {
			bt.WriteString(";")
		}
	}
	return bt.String()
}

func GetInsertSql(schemaName string, tableName string, columnNames []string, columnValue []string) string {
	var bt bytes.Buffer
	bt.WriteString("insert into " + schemaName + "." + tableName + " (")
	for i, columnName := range columnNames {
		if i+1 < len(columnNames) {
			bt.WriteString(columnName + ", ")
		} else {
			bt.WriteString(columnName + ") values (")
		}
	}
	for i, val := range columnValue {
		if i+1 < len(columnNames) {
			bt.WriteString(val + ", ")
		} else {
			bt.WriteString(val + ");")
		}
	}
	return bt.String()
}

func GetdeleteSql(schemaName string, tableName string, columnNames []string, columnValue []string) string {
	var bt bytes.Buffer
	bt.WriteString("delete from " + schemaName + "." + tableName + " where ")
	for i, columnName := range columnNames {
		if i+1 < len(columnName) {
			bt.WriteString(columnName + " = " + columnValue[i] + " and ")
		} else {
			bt.WriteString(columnName + " = " + columnValue[i] + ";")
		}
	}
	return bt.String()
}
