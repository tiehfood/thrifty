package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func isInvalidFlow(flow Flow) bool {
	if flow.Amount != 0.0 {
		if len(flow.Name) > 0 {
			return false
		}
	}
	return true
}

// getFlows godoc
// @Summary      Get all flows
// @Description  Get all flows for the current user in a JSON object
// @Tags         Flows
// @Produce      json
// @Success      200  {array}   main.Flow
// @Failure      500  {object}  main.HTTPError
// @Router       /flows [get]
func getFlows(c *gin.Context) {
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	query := `
SELECT
	json_group_array(json_object('id', id, 'name', name, 'description', description, 'amount', amount, 'icon', data, 'groupId', groupId)) AS json_result
FROM
	(
		SELECT
			f.id, f.name as name, f.description, f.amount, icons.data, f.groupId
		FROM
			flows f
		LEFT JOIN icons ON f.iconId = icons.id
		WHERE f.userId = (SELECT currentUserId FROM settings WHERE id = 1)
		AND f.groupId IS NULL
		ORDER BY
			amount DESC, name
	);
`
	var jsonResult []byte
	err := dbCon.QueryRow(query).Scan(&jsonResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query database"})
		return
	}

	if !json.Valid(jsonResult) {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Invalid data returned from database"})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jsonResult)
}

// addFlow godoc
// @Summary      Add new flow
// @Description  Add a new flow for the current user
// @Tags         Flows
// @Param request body   main.Flow true "Flow object, id is set by the server and could be omitted"
// @Produce      json
// @Success      201  {object}  main.Flow
// @Failure      500  {object}  main.HTTPError
// @Router       /flows [post]
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

	currentUserId, err := getCurrentUserId(dbCon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to get current user"})
		return
	}

	query := "INSERT OR REPLACE INTO flows (id, name, description, amount, userId, groupId) VALUES (?, ?, ?, ?, ?, ?)"
	queryArgs := []interface{}{newFlow.ID, newFlow.Name, newFlow.Description, newFlow.Amount, currentUserId, newFlow.GroupId}
	if newFlow.Icon != "" {
		iconId := insertIcon(dbCon, newFlow, nil)
		query = "INSERT OR REPLACE INTO flows (id, name, description, amount, iconId, userId, groupId) VALUES (?, ?, ?, ?, ?, ?, ?)"
		queryArgs = []interface{}{newFlow.ID, newFlow.Name, newFlow.Description, newFlow.Amount, iconId, currentUserId, newFlow.GroupId}
	}

	execSql(dbCon, query, queryArgs...)
	c.IndentedJSON(http.StatusCreated, newFlow)
}

// updateFlow godoc
// @Summary      Update existing flow
// @Description  Update an existing flow with new data
// @Tags         Flows
// @Param id      path  int       true "Flow id"
// @Param request body  main.Flow true "Flow object, id is ignored and could be omitted"
// @Produce      json
// @Success      200  {object}  main.Flow
// @Failure      500  {object}  main.HTTPError
// @Router       /flows/{id} [patch]
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

	row, err := getSqlRow(dbCon,
		"SELECT flows.id, iconId, hash FROM flows JOIN icons ON flows.iconId = icons.id WHERE flows.id = ?",
		newFlow.ID)
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
		execSql(dbCon, "UPDATE flows SET iconId = ? WHERE id = ?", defaultIconId, newFlow.ID)
		if isIconUnused(dbCon, iconId) {
			execSql(dbCon, "DELETE FROM icons WHERE id = ?", iconId)
		}
	}

	query := "UPDATE flows SET name = ?, description = ?, amount = ?, groupId = ? WHERE id = ?"
	queryArgs := []interface{}{newFlow.Name, newFlow.Description, newFlow.Amount, newFlow.GroupId, newFlow.ID}
	if newFlow.Icon != "" {
		iconId := insertIcon(dbCon, newFlow, nil)
		query = "UPDATE flows SET name = ?, description = ?, amount = ?, iconId = ?, groupId = ? WHERE id = ?"
		queryArgs = []interface{}{newFlow.Name, newFlow.Description, newFlow.Amount, iconId, newFlow.GroupId, newFlow.ID}
	}

	execSql(dbCon, query, queryArgs...)
	c.JSON(http.StatusOK, newFlow)
}

// deleteFlow godoc
// @Summary      Delete a flow
// @Description  Delete an existing flow
// @Tags         Flows
// @Param id path int true "Flow id"
// @Produce      json
// @Success      200  {object}  main.HTTPResponse
// @Failure      500  {object}  main.HTTPError
// @Router       /flows/{id} [delete]
func deleteFlow(c *gin.Context) {
	flowId := c.Param("id")
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	row, err := getSqlRow(dbCon, "SELECT id, iconId FROM flows WHERE id = ?", flowId)
	if err != nil || row[0] == nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Flow with given ID does not exist"})
		return
	}

	iconId, ok := row[1].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
		return
	}

	execSql(dbCon, "DELETE FROM flows WHERE id = ?", flowId)
	if iconId != defaultIconId && isIconUnused(dbCon, iconId) {
		execSql(dbCon, "DELETE FROM icons WHERE id = ?", iconId)
	}
	c.JSON(http.StatusOK, HTTPResponse{Ok: "Flow deleted"})
}
