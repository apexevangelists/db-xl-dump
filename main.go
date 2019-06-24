package main

/* db-xl-dump */

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
	_ "gopkg.in/goracle.v2"
)

// TConfig - parameters in config file
type TConfig struct {
	configFile       string
	debugMode        bool
	delimiter        string
	enclosedBy       string
	headers          bool
	exportObject     string
	connectionConfig string
	connectionsDir   string
	outputFilename   string
}

// TConnection - parameters passed by the user
type TConnection struct {
	dbConnectionString string
	username           string
	password           string
	hostname           string
	port               int
	service            string
}

var config = new(TConfig)
var connection TConnection

var logger = loggo.GetLogger("dbcsvdump")

var file *xlsx.File
var sheet *xlsx.Sheet
var row *xlsx.Row
var cell *xlsx.Cell
var err error

/********************************************************************************/
func setDebug(debugMode bool) {
	if debugMode == true {
		loggo.ConfigureLoggers("dbcsvdump=DEBUG")
		logger.Debugf("Debug log enabled")
	}
}

/********************************************************************************/
func parseFlags() {

	flag.StringVar(&config.configFile, "configFile", "config", "Configuration file for general parameters")
	flag.BoolVar(&config.headers, "headers", true, "Output Headers")
	flag.StringVar(&config.exportObject, "export", "", "Table(s), View(s) or querys to export")
	flag.StringVar(&config.exportObject, "e", "", "Table(s), View(s) or querys to export")
	flag.StringVar(&config.outputFilename, "output", "output.xlsx", "Output Filename")
	flag.StringVar(&config.outputFilename, "o", "output.xlsx", "Output Filename")

	flag.BoolVar(&config.debugMode, "debug", false, "Debug mode (default=false)")
	flag.StringVar(&config.connectionConfig, "connection", "", "Confguration file for connection")

	flag.StringVar(&connection.dbConnectionString, "db", "", "Database Connection, e.g. user/password@host:port/sid")

	flag.Parse()

	// At a minimum we either need a dbConnection or a configFile
	if (config.configFile == "") && (connection.dbConnectionString == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

}

/********************************************************************************/
func getPassword() []byte {
	fmt.Printf("Password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		// Handle gopass.ErrInterrupted or getch() read error
	}

	return pass
}

/********************************************************************************/
func getConnectionString(connection TConnection) string {

	if connection.dbConnectionString != "" {
		return connection.dbConnectionString
	}

	var str = fmt.Sprintf("%s/%s@%s:%d/%s", connection.username,
		connection.password,
		connection.hostname,
		connection.port,
		connection.service)

	return str
}

/********************************************************************************/
// To execute, at a minimum we need (connection && (object || sql))
func checkMinFlags() {
	// connection is required
	bHaveConnection := (getConnectionString(connection) != "")

	// check if we have either an object to export or a SQL statement
	bHaveObject := (config.exportObject != "")

	if !bHaveConnection || !bHaveObject {
		fmt.Printf("%s:\n", os.Args[0])
	}

	if !bHaveConnection {
		fmt.Printf("  requires a DB connection to be specified\n")
	}

	if !bHaveObject {
		fmt.Printf("  requires either an Object (Table or View) or SQL to export\n")
	}

	if !bHaveConnection || !bHaveObject {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

/********************************************************************************/
func loadConfig(configFile string) {
	if config.configFile == "" {
		return
	}

	logger.Debugf("reading configFile: %s", configFile)
	viper.SetConfigType("yaml")
	viper.SetConfigName(configFile)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// need to set debug mode if it's not already set
	setDebug(viper.GetBool("debugMode"))

	config.connectionsDir = viper.GetString("connectionsDir")
	config.connectionConfig = viper.GetString("connectionConfig")
	config.debugMode = viper.GetBool("debugMode")
	config.configFile = configFile
}

/********************************************************************************/
func loadConnection(connectionFile string) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(config.connectionConfig)
	v.AddConfigPath(config.connectionsDir)

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	connection.dbConnectionString = v.GetString("dbConnectionString")
	connection.dbConnectionString = v.GetString("dbConnectionString")
	connection.username = v.GetString("username")
	connection.password = v.GetString("password")
	connection.hostname = v.GetString("hostname")
	connection.port = v.GetInt("port")
	connection.service = v.GetString("service")

}

/********************************************************************************/
func debugConfig() {
	logger.Debugf("config.configFile: %s\n", config.configFile)
	logger.Debugf("config.debugMode: %s\n", strconv.FormatBool(config.debugMode))
	logger.Debugf("config.delimiter: %s\n", config.delimiter)
	logger.Debugf("config.enclosedBy: %s\n", config.enclosedBy)
	logger.Debugf("config.headers: %s\n", strconv.FormatBool(config.headers))
	logger.Debugf("config.exportObject: %s\n", config.exportObject)
	logger.Debugf("config.connectionConfig: %s\n", config.connectionConfig)
	logger.Debugf("connection.dbConnectionString: %s\n", connection.dbConnectionString)
}

/********************************************************************************/
func outputHeaders(cols []string) {
	// Output the headers
	row = sheet.AddRow()

	for _, colName := range cols {
		cell = row.AddCell()
		cell.Value = colName
	}
}

/********************************************************************************/
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

/********************************************************************************/
func floatValue(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		fmt.Printf("floatValue: %s error: %s\n", s, err)
		os.Exit(1)
	}

	return v
}

/********************************************************************************/
func outputData(rows *sql.Rows) {
	cols, _ := rows.Columns()
	data := make(map[string]string)

	for rows.Next() {
		columns := make([]string, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		rows.Scan(columnPointers...)

		row = sheet.AddRow()

		for i, colName := range cols {
			data[colName] = columns[i]
			cell = row.AddCell()

			if isNumeric(data[colName]) {
				cell.SetFloat(floatValue(data[colName]))
			} else {
				cell.Value = data[colName]
			}
		}

	}
}

/********************************************************************************/
func process(object string) {
	var bQuery bool
	var sheetName string

	db, err := sql.Open("goracle", getConnectionString(connection))

	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	var query string
	// check if object starts with the word 'select'
	// otherwise we'll assume it's an object name
	if strings.HasPrefix(strings.ToUpper(object), "SELECT") {
		bQuery = true
		query = object
	} else {
		query = fmt.Sprintf("select * from %s", object)
		logger.Debugf("Using query: %s\n", query)
	}

	rows, err := db.Query(query)

	if err != nil {
		fmt.Println("Error running query")
		fmt.Println(err)
		return
	}
	defer rows.Close()

	if bQuery {
		sheetName = fmt.Sprintf("Sheet %d", len(file.Sheets))
	} else {
		sheetName = object
	}
	sheet, err = file.AddSheet(sheetName)
	if err != nil {
		fmt.Printf(err.Error())
	}

	cols, _ := rows.Columns()

	if config.headers == true {
		outputHeaders(cols)
	}

	outputData(rows)
}

/********************************************************************************/
func main() {
	parseFlags()
	setDebug(config.debugMode)
	loadConfig(config.configFile)
	loadConnection(config.connectionConfig)

	debugConfig()
	checkMinFlags()

	if connection.password == "" {
		connection.password = string(getPassword())
	}

	file = xlsx.NewFile()

	// see if need to loop round the input
	s := strings.Split(config.exportObject, ",")
	logger.Debugf("Count: %d\n", len(s))
	for i, object := range s {
		process(object)
		logger.Debugf("[%d]: %s\n", i, object)
	}

	err = file.Save(config.outputFilename)
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("Created file: %s\n", config.outputFilename)
}
