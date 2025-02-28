package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
	"tiehfood/thrifty/docs"
)

type Flow struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Icon        string  `json:"icon"`
	Tags      []string  `json:"tags"`
}

type HTTPResponse struct {
	Ok string `json:"ok"`
}

type HTTPError struct {
	Error string `json:"error"`
}

var defaultIconId = "00000000-0000-0000-0000-000000000000"
var errorPrefix = "Error: "

func main() {
	dbCon, err := initAndOpenDb()
	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	docs.SwaggerInfo.Title = "Thrifty API"
	docs.SwaggerInfo.Description = "This is the documentation for the Thrifty API"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Version = "1.0"

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PATCH", "OPTIONS", "DELETE"},
		AllowHeaders: []string{"Content-Type"},
	}))

	router.Use(func(context *gin.Context) {
		context.Set("dbCon", dbCon)
		context.Next()
	})

	v1 := router.Group("/api")
	{
		flows := v1.Group("/flows")
		{
			flows.GET("", getFlows)
			flows.POST("", addFlow)
			flows.PATCH(":id", updateFlow)
			flows.DELETE(":id", deleteFlow)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	port, err := getAndValidatePort()
	if err != nil {
		fmt.Println(errorPrefix, err)
	}
	fmt.Printf("Running on port: %d\n", port)
	err = router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(errorPrefix, err)
	}
}

func getAndValidatePort() (int, error) {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080, nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 8080, fmt.Errorf("invalid PORT value: %v", err)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("PORT value out of range (1-65535)")
	}

	return port, nil
}

func isInvalidFlow(flow Flow) bool {
	if flow.Amount != 0.0 {
		if len(flow.Name) > 0 {
			return false
		}
	}
	return true
}

func initAndOpenDb() (*sql.DB, error) {
	dataSourceName := os.Getenv("SQLITE_DB_PATH")
	defaultIconFlow := Flow{
		Icon: "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPHN2ZyBpZD0iTGF5ZXJfMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCAyNTQuMiAyNTQuMiI+CiAgPGRlZnM+CiAgICA8c3R5bGU+CiAgICAgIC5zdDAgewogICAgICAgIGZpbGw6ICM0NDliMzg7CiAgICAgIH0KCiAgICAgIC5zdDEgewogICAgICAgIGZpbGw6ICNmZmMzMzY7CiAgICAgIH0KCiAgICAgIC5zdDIgewogICAgICAgIGZpbGw6ICNjODkzMTg7CiAgICAgIH0KCiAgICAgIC5zdDMgewogICAgICAgIGZpbGw6ICM2MmM3NTE7CiAgICAgIH0KCiAgICAgIC5zdDQgewogICAgICAgIGZpbGw6ICNlYWIwMjg7CiAgICAgIH0KICAgIDwvc3R5bGU+CiAgPC9kZWZzPgogIDxyZWN0IGNsYXNzPSJzdDAiIHk9IjU0LjYiIHdpZHRoPSIyNTQuMiIgaGVpZ2h0PSIxMzAuNyIvPgogIDxwYXRoIGNsYXNzPSJzdDMiIGQ9Ik0yMTMuNSw3My41SDQwLjhjMCwxMi4xLTkuOCwyMS45LTIxLjksMjEuOXY0OS4yYzEyLjEsMCwyMS45LDkuOCwyMS45LDIxLjloMTcyLjdjMC0uMiwwLS40LDAtLjcsMC0xMi4xLDkuOC0yMS45LDIxLjktMjEuOXYtNDkuMmMtMTEuOCwwLTIxLjUtOS40LTIxLjgtMjEuMloiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDAiIGN4PSIxMjcuMSIgY3k9IjEyMCIgcj0iMzguOSIvPgogIDxjaXJjbGUgY2xhc3M9InN0MCIgY3g9IjY3IiBjeT0iMTIwIiByPSI3LjgiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDAiIGN4PSIxODcuMiIgY3k9IjEyMCIgcj0iNy44IiB0cmFuc2Zvcm09InRyYW5zbGF0ZSgzOC44IDI4NS41KSByb3RhdGUoLTgwLjgpIi8+CiAgPHBhdGggY2xhc3M9InN0MyIgZD0iTTEyNi44LDE0MC40Yy01LjksMC0xMS4yLTMuMS0xMy4zLThsNy43LTMuNWMuOCwxLjgsMy4xLDMsNS42LDNzNC44LTEuMiw1LjYtM2MuMy0uNi4zLTEuMi4yLTEuNS0uMi0uNS0xLjQtMS45LTYuNy0zLjUtNC40LTEuMy05LjctMy40LTExLjktNy45LS45LTEuOC0xLjUtNC43LDAtOC41LDIuMi01LDcuNS04LjEsMTMuNC04LjFzMTEuMiwzLjEsMTMuMyw4bC03LjcsMy41Yy0uOC0xLjgtMy4xLTMtNS42LTNzLTQuOCwxLjItNS42LDNjLS4zLjYtLjMsMS4yLS4yLDEuNS4yLjUsMS40LDEuOSw2LjcsMy41LDQuNCwxLjMsOS43LDMuNCwxMS45LDcuOS45LDEuOCwxLjUsNC43LDAsOC41LTIuMiw1LTcuNSw4LjEtMTMuNCw4LjFaIi8+CiAgPHJlY3QgY2xhc3M9InN0MyIgeD0iMTIyLjYiIHk9IjEzNi4yIiB3aWR0aD0iOC41IiBoZWlnaHQ9IjkuOCIvPgogIDxyZWN0IGNsYXNzPSJzdDMiIHg9IjEyMi42IiB5PSI5My45IiB3aWR0aD0iOC41IiBoZWlnaHQ9IjkuOCIvPgogIDxjaXJjbGUgY2xhc3M9InN0MyIgY3g9IjI0LjQiIGN5PSI3OC44IiByPSI1LjUiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDMiIGN4PSIyNC40IiBjeT0iMTYxIiByPSI1LjUiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDMiIGN4PSIyMjkuOCIgY3k9Ijc4LjkiIHI9IjUuNSIvPgogIDxjaXJjbGUgY2xhc3M9InN0MyIgY3g9IjIyOS44IiBjeT0iMTYxLjEiIHI9IjUuNSIvPgogIDxjaXJjbGUgY2xhc3M9InN0NCIgY3g9IjIxNy43IiBjeT0iMTY2LjYiIHI9IjI5LjUiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDEiIGN4PSIyMTcuNyIgY3k9IjE2Ni42IiByPSIyMy45Ii8+CiAgPHBhdGggY2xhc3M9InN0MiIgZD0iTTIxNy41LDE3OS4xYy0zLjYsMC02LjktMS45LTguMi00LjlsNC44LTIuMWMuNSwxLjEsMS45LDEuOCwzLjQsMS44czMtLjgsMy40LTEuOGMuMi0uNC4yLS43LjEtLjktLjEtLjMtLjktMS4xLTQuMS0yLjEtMi43LS44LTYtMi4xLTcuMy00LjktLjUtMS4xLS45LTIuOSwwLTUuMiwxLjQtMy4xLDQuNi01LDguMi01czYuOSwxLjksOC4yLDQuOWwtNC44LDIuMWMtLjUtMS4xLTEuOS0xLjgtMy40LTEuOHMtMywuOC0zLjQsMS44Yy0uMi40LS4yLjctLjEuOS4xLjMuOSwxLjEsNC4xLDIuMSwyLjcuOCw2LDIuMSw3LjMsNC45LjUsMS4xLjksMi45LDAsNS4yLTEuNCwzLjEtNC42LDUtOC4yLDVaIi8+CiAgPHJlY3QgY2xhc3M9InN0MiIgeD0iMjE0LjkiIHk9IjE3Ni41IiB3aWR0aD0iNS4yIiBoZWlnaHQ9IjYiLz4KICA8cmVjdCBjbGFzcz0ic3QyIiB4PSIyMTQuOSIgeT0iMTUwLjYiIHdpZHRoPSI1LjIiIGhlaWdodD0iNiIvPgogIDxjaXJjbGUgY2xhc3M9InN0NCIgY3g9IjE3MC42IiBjeT0iMTgzLjEiIHI9IjI5LjUiIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0zNy41IDMyMi4xKSByb3RhdGUoLTgwLjgpIi8+CiAgPGNpcmNsZSBjbGFzcz0ic3QxIiBjeD0iMTcwLjUiIGN5PSIxODMuMSIgcj0iMjMuOSIvPgogIDxwYXRoIGNsYXNzPSJzdDIiIGQ9Ik0xNzAuMywxOTUuN2MtMy42LDAtNi45LTEuOS04LjItNC45bDQuOC0yLjFjLjUsMS4xLDEuOSwxLjgsMy40LDEuOHMzLS44LDMuNC0xLjhjLjItLjQuMi0uNy4xLS45LS4xLS4zLS45LTEuMS00LjEtMi4xLTIuNy0uOC02LTIuMS03LjMtNC45LS41LTEuMS0uOS0yLjksMC01LjIsMS40LTMuMSw0LjYtNSw4LjItNXM2LjksMS45LDguMiw0LjlsLTQuOCwyLjFjLS41LTEuMS0xLjktMS44LTMuNC0xLjhzLTMsLjgtMy40LDEuOGMtLjIuNC0uMi43LS4xLjkuMS4zLjksMS4xLDQuMSwyLjEsMi43LjgsNiwyLjEsNy4zLDQuOS41LDEuMS45LDIuOSwwLDUuMi0xLjQsMy4xLTQuNiw1LTguMiw1WiIvPgogIDxyZWN0IGNsYXNzPSJzdDIiIHg9IjE2Ny43IiB5PSIxOTMuMSIgd2lkdGg9IjUuMiIgaGVpZ2h0PSI2Ii8+CiAgPHJlY3QgY2xhc3M9InN0MiIgeD0iMTY3LjciIHk9IjE2Ny4xIiB3aWR0aD0iNS4yIiBoZWlnaHQ9IjYiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDQiIGN4PSIyMTQuOCIgY3k9IjIwNS42IiByPSIyOS41Ii8+CiAgPGNpcmNsZSBjbGFzcz0ic3QxIiBjeD0iMjE0LjgiIGN5PSIyMDUuNiIgcj0iMjMuOSIvPgogIDxwYXRoIGNsYXNzPSJzdDIiIGQ9Ik0yMTQuNiwyMTguMmMtMy42LDAtNi45LTEuOS04LjItNC45bDQuOC0yLjFjLjUsMS4xLDEuOSwxLjgsMy40LDEuOHMzLS44LDMuNC0xLjhjLjItLjQuMi0uNy4xLS45LS4xLS4zLS45LTEuMS00LjEtMi4xLTIuNy0uOC02LTIuMS03LjMtNC45LS41LTEuMS0uOS0yLjksMC01LjIsMS40LTMuMSw0LjYtNSw4LjItNXM2LjksMS45LDguMiw0LjlsLTQuOCwyLjFjLS41LTEuMS0xLjktMS44LTMuNC0xLjhzLTMsLjgtMy40LDEuOGMtLjIuNC0uMi43LS4xLjkuMS4zLjksMS4xLDQuMSwyLjEsMi43LjgsNiwyLjEsNy4zLDQuOS41LDEuMS45LDIuOSwwLDUuMi0xLjQsMy4xLTQuNiw1LTguMiw1WiIvPgogIDxyZWN0IGNsYXNzPSJzdDIiIHg9IjIxMiIgeT0iMjE1LjYiIHdpZHRoPSI1LjIiIGhlaWdodD0iNiIvPgogIDxyZWN0IGNsYXNzPSJzdDIiIHg9IjIxMiIgeT0iMTg5LjYiIHdpZHRoPSI1LjIiIGhlaWdodD0iNiIvPgo8L3N2Zz4=",
	}

	if len(dataSourceName) == 0 {
		dataSourceName = "thrifty.sqlite"
	}
	dbCon, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	queryFlows := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS flows (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT,
    description TEXT,
    amount float,
    iconId TEXT DEFAULT '%s',
    FOREIGN KEY (iconId) REFERENCES icons(id));
`, defaultIconId)

	queryIcons := `
CREATE TABLE IF NOT EXISTS icons (
    id TEXT NOT NULL PRIMARY KEY,
    data TEXT,
    hash TEXT);
`

	addTagsTable := `
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER NOT NULL PRIMARY KEY,
    tag TEXT UNIQUE);
CREATE TABLE IF NOT EXISTS flows_tags (
    flowId TEXT,
    tagId INTEGER,
    FOREIGN KEY (flowId) REFERENCES flows(id),
    FOREIGN KEY (tagId) REFERENCES tags(id));
`

	printSqlVersion(dbCon)
	execSql(dbCon, queryFlows)
	execSql(dbCon, queryIcons)
	execSql(dbCon, addTagsTable)
	insertIcon(dbCon, defaultIconFlow, &defaultIconId)
	return dbCon, nil
}

func printSqlVersion(dbCon *sql.DB) {
	query := "SELECT sqlite_version();"
	row, err := getSqlRow(dbCon, query, nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("SQLite version: ", row[0])
}

func execSql(dbCon *sql.DB, query string, args ...interface{}) int64 {
	result, err := dbCon.Exec(query, args...)
	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	id, _ := result.LastInsertId()
	numRows, _ := result.RowsAffected()
	fmt.Printf("LastInsertId %d, RowsAffected: %d\n", id, numRows)

	return id
}

func getSqlRow(dbCon *sql.DB, query string, args ...interface{}) ([]interface{}, error) {
	rows, err := dbCon.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	row := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range row {
		scanArgs[i] = &row[i]
	}

	// get first row
	if rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
	}
	return row, rows.Err()
}

func getSqlRows(dbCon *sql.DB, query string, args ...interface{}) ([][]interface{}, error) {
	rows, err := dbCon.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var data [][]interface{}

	for rows.Next() {
		row := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range row {
			scanArgs[i] = &row[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return data, err
		}
		data = append(data, row)
	}
	if err = rows.Err(); err != nil {
		return data, err
	}
	return data, nil
}

func getMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func getDbConnection(c *gin.Context) *sql.DB {
	dbConVar, exists := c.Get("dbCon")
	if !exists {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database connection not found"})
		return nil
	}
	dbCon, ok := dbConVar.(*sql.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Invalid database connection"})
		return nil
	}
	return dbCon
}

func deleteTagsFromFlow(dbCon *sql.DB, flowId *string) {
	query := "DELETE FROM flows_tags WHERE flowId = ?; DELETE FROM tags WHERE tags.id NOT IN (SELECT tagId FROM flows_tags)"
	execSql(dbCon, query, []interface{}{flowId}...)
}

func getOrCreateTag(dbCon *sql.DB, tag *string) (int64, error) {
	if tag == nil || len(*tag) == 0 {
		return 0, fmt.Errorf("Empty tag given")
	}
	query := "SELECT id, tag FROM tags WHERE tag = ?"
	row, err := getSqlRow(dbCon, query, []interface{}{tag}...)

	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	var tagId int64
	if row[0] == nil {
		query = "INSERT OR REPLACE INTO tags (tag) VALUES (?)"
		queryArgs := []interface{}{tag}
		tagId = execSql(dbCon, query, queryArgs...)
	} else {
		tagId = row[0].(int64)
	}

	return tagId, nil
}

func insertIcon(dbCon *sql.DB, flow Flow, iconId *string) string {
	hash := getMD5(flow.Icon)

	if iconId == nil {
		var id = uuid.New().String()
		iconId = &id
	}

	query := "SELECT id, data, hash FROM icons WHERE hash = ?"
	row, err := getSqlRow(dbCon, query, []interface{}{hash}...)

	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	if row[0] == nil {
		query = "INSERT OR REPLACE INTO icons (id, data, hash) VALUES (?, ?, ?)"
		queryArgs := []interface{}{*iconId, flow.Icon, hash}
		execSql(dbCon, query, queryArgs...)
	} else {
		var value = row[0].(string)
		iconId = &value
	}
	return *iconId
}

// getFlows godoc
// @Summary      Get all flows
// @Description  Get all flows in a JSON object
// @Tags         Flows
// @Produce      json
// @Success      200  	{array}  main.Flow
// @Failure      500  	{object} main.HTTPError
// @Router       /flows [get]
func getFlows(c *gin.Context) {
	query := `
SELECT
	json_group_array(json_object('id', id, 'name', name, 'description', description, 'amount', amount, 'icon', data)) AS json_result
FROM
	(
		SELECT
			f.id, f.name as name, f.description, f.amount, icons.data
		FROM
			flows f
		LEFT JOIN icons ON f.iconId = icons.id
		ORDER BY
			amount DESC, name
	);
`
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	var jsonResult []byte
	var flows []Flow
	err := dbCon.QueryRow(query).Scan(&jsonResult)
	json.Unmarshal(jsonResult, &flows)

	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query database"})
		return
	}

	for i, item := range flows {
		tagQuery := `
SELECT
    tags.tag
FROM tags
LEFT JOIN flows_tags ON tags.id = flows_tags.tagId
WHERE flows_tags.flowId = ?
`
		rows, err := getSqlRows(dbCon, tagQuery, []interface{}{item.ID}...)
		if err != nil {
		     fmt.Println(errorPrefix, err)
		}

		var tags []string
		for _, row := range rows {
			if len(row) > 0 {
				if tag, ok := row[0].(string); ok {
					tags = append(tags, tag)
				}
			}
		}
		flows[i].Tags = tags
	}

	jsonWithTags, _ := json.Marshal(flows)

	if !json.Valid(jsonWithTags) {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Invalid data returned from database"})
	}
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jsonWithTags)
}

// addFlow godoc
// @Summary      	Add new flow
// @Description  	Get all flows in a JSON object
// @Tags         	Flows
// @Param request 	body 	main.Flow true "Flow object, id is set by the server and could be omitted"
// @Produce      	json
// @Success      	200  	{object}  main.Flow
// @Failure      	500  	{object}  main.HTTPError
// @Router       	/flows 	[post]
func addFlow(c *gin.Context) {
	var newFlow Flow

	if err := c.BindJSON(&newFlow); err != nil {
		fmt.Println(errorPrefix, err)
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}

	if isInvalidFlow(newFlow) {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Please provide name and amount"})
		return
	}

	newFlow.ID = uuid.New().String()

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	query := "INSERT OR REPLACE INTO flows (id, name, description, amount) VALUES (?, ?, ?, ?)"
	queryArgs := []interface{}{newFlow.ID, newFlow.Name, newFlow.Description, newFlow.Amount}
	if newFlow.Icon != "" {
		iconId := insertIcon(dbCon, newFlow, nil)
		query = "INSERT OR REPLACE INTO flows (id, name, description, amount, iconId) VALUES (?, ?, ?, ?, ?)"
		queryArgs = []interface{}{newFlow.ID, newFlow.Name, newFlow.Description, newFlow.Amount, iconId}
	}
	execSql(dbCon, query, queryArgs...)

	if len(newFlow.Tags) > 0 {
		for _, tag := range newFlow.Tags {
			tagId, err := getOrCreateTag(dbCon, &tag)
			if (err == nil) {
				execSql(dbCon, "INSERT OR IGNORE INTO flows_tags (flowId, tagId) VALUES (?, ?)", []interface{}{newFlow.ID, &tagId}...)
			}
		}
	}

	c.IndentedJSON(http.StatusCreated, newFlow)
}

// updateFlow godoc
// @Summary      	Update existing flow
// @Description  	Update an existing flow with new data
// @Tags         	Flows
// @Param id path 	int 		true 		"Flow id"
// @Param request 	body 		main.Flow 	true "Flow object, id is ignored and could be omitted"
// @Produce      	json
// @Success      	200  		{object}  	main.Flow
// @Failure      	500  		{object}  	main.HTTPError
// @Router       	/flows/{id} [patch]
func updateFlow(c *gin.Context) {
	var newFlow Flow

	if err := c.BindJSON(&newFlow); err != nil {
		fmt.Println(errorPrefix, err)
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}

	if isInvalidFlow(newFlow) {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Please provide name and amount"})
		return
	}

	newFlow.ID = c.Param("id")

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	query := "SELECT flows.id, iconId, hash FROM flows JOIN icons ON flows.iconId = icons.id WHERE flows.id = ?"
	row, err := getSqlRow(dbCon, query, []interface{}{newFlow.ID}...)

	if err != nil || row[0] == nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Flow with given ID does not exist"})
		return
	}

	iconId, ok := row[1].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
		return
	}

	hash, ok := row[2].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
		return
	}

	if iconId != defaultIconId && hash != getMD5(newFlow.Icon) {
		query = "SELECT count(id) FROM flows WHERE iconId = ?"
		row, err = getSqlRow(dbCon, query, []interface{}{iconId}...)
		count, ok := row[0].(int64)

		if err != nil || !ok {
			c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
			return
		}

		if count == 1 {
			query = "DELETE FROM icons WHERE id = ?"
			execSql(dbCon, query, []interface{}{iconId}...)
		}
	}

	query = "UPDATE flows SET name = ?, description = ?, amount = ? WHERE id = ?"
	queryArgs := []interface{}{newFlow.Name, newFlow.Description, newFlow.Amount, newFlow.ID}
	if newFlow.Icon != "" {
		iconId := insertIcon(dbCon, newFlow, nil)
		query = "UPDATE flows SET name = ?, description = ?, amount = ?, iconId = ? WHERE id = ?"
		queryArgs = []interface{}{newFlow.Name, newFlow.Description, newFlow.Amount, iconId, newFlow.ID}
	}

	execSql(dbCon, query, queryArgs...)

	deleteTagsFromFlow(dbCon, &newFlow.ID)
	if len(newFlow.Tags) > 0 {
		for _, tag := range newFlow.Tags {
			tagId, err := getOrCreateTag(dbCon, &tag)
			if (err == nil) {
				execSql(dbCon, "INSERT OR IGNORE INTO flows_tags (flowId, tagId) VALUES (?, ?)", []interface{}{newFlow.ID, &tagId}...)
			}
		}
	}

	c.JSON(http.StatusOK, newFlow)
}

// deleteFlow godoc
// @Summary      	Delete a flow
// @Description  	Delete an existing flow
// @Tags         	Flows
// @Param id path 	int 		true 		"Flow id"
// @Produce      	json
// @Success      	200  		{object}	main.HTTPResponse
// @Failure      	500  		{object}  	main.HTTPError
// @Router       	/flows/{id} [delete]
func deleteFlow(c *gin.Context) {
	flowId := c.Param("id")
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	query := "SELECT id, iconId FROM flows WHERE id = ?"
	row, err := getSqlRow(dbCon, query, []interface{}{flowId}...)

	if err != nil || row[0] == nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Flow with given ID does not exist"})
		return
	}

	iconId, ok := row[1].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
		return
	}

	if iconId != defaultIconId {
		query = "SELECT count(id) FROM flows WHERE iconId = ?"
		row, err = getSqlRow(dbCon, query, []interface{}{iconId}...)
		count, ok := row[0].(int64)

		if err != nil || !ok {
			c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
			return
		}

		if count == 1 {
			query = "DELETE FROM icons WHERE id = ?"
			execSql(dbCon, query, []interface{}{iconId}...)
		}
	}
	deleteTagsFromFlow(dbCon, &flowId)
	query = "DELETE FROM flows WHERE id = ?"
	execSql(dbCon, query, []interface{}{flowId}...)
	c.JSON(http.StatusOK, HTTPResponse{Ok: "Flow deleted"})
}
