package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getGroups godoc
// @Summary      Get all groups
// @Description  Get all groups for the current user with aggregated amount
// @Tags         Groups
// @Produce      json
// @Success      200  {array}   main.Group
// @Failure      500  {object}  main.HTTPError
// @Router       /groups [get]
func getGroups(c *gin.Context) {
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	query := `
SELECT
	json_group_array(json_object('id', id, 'name', name, 'description', description, 'icon', icon, 'amount', amount, 'entryCount', entryCount)) AS json_result
FROM (
	SELECT
		g.id,
		g.name,
		g.description,
		i.data AS icon,
		COALESCE((SELECT SUM(f.amount) FROM flows f WHERE f.groupId = g.id), 0) AS amount,
		(SELECT count(*) FROM flows f WHERE f.groupId = g.id) AS entryCount
	FROM groups g
	LEFT JOIN icons i ON g.iconId = i.id
	WHERE g.userId = (SELECT currentUserId FROM settings WHERE id = 1)
	ORDER BY amount DESC, g.name
);
`
	var jsonResult []byte
	err := dbCon.QueryRow(query).Scan(&jsonResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query groups"})
		return
	}

	if !json.Valid(jsonResult) {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Invalid data returned from database"})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jsonResult)
}

// createGroup godoc
// @Summary      Create a new group
// @Description  Create a new group for the current user
// @Tags         Groups
// @Param request body   main.Group true "Group object"
// @Produce      json
// @Success      201  {object}  main.Group
// @Failure      500  {object}  main.HTTPError
// @Router       /groups [post]
func createGroup(c *gin.Context) {
	var group Group
	if err := c.BindJSON(&group); err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}
	if len(group.Name) == 0 {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Please provide a name"})
		return
	}

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	currentUserId, err := getCurrentUserId(dbCon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to get current user"})
		return
	}

	group.ID = uuid.New().String()
	query := "INSERT INTO groups (id, name, description, userId) VALUES (?, ?, ?, ?)"
	queryArgs := []interface{}{group.ID, group.Name, group.Description, currentUserId}
	if group.Icon != "" {
		iconId := insertIcon(dbCon, Flow{Icon: group.Icon}, nil)
		query = "INSERT INTO groups (id, name, description, iconId, userId) VALUES (?, ?, ?, ?, ?)"
		queryArgs = []interface{}{group.ID, group.Name, group.Description, iconId, currentUserId}
	}

	execSql(dbCon, query, queryArgs...)
	c.IndentedJSON(http.StatusCreated, group)
}

// updateGroup godoc
// @Summary      Update an existing group
// @Description  Update name, description, or icon of a group
// @Tags         Groups
// @Param id      path  string     true "Group id"
// @Param request body  main.Group true "Group object"
// @Produce      json
// @Success      200  {object}  main.Group
// @Failure      500  {object}  main.HTTPError
// @Router       /groups/{id} [patch]
func updateGroup(c *gin.Context) {
	var group Group
	if err := c.BindJSON(&group); err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}
	if len(group.Name) == 0 {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Please provide a name"})
		return
	}

	group.ID = c.Param("id")

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	row, err := getSqlRow(dbCon,
		"SELECT groups.id, iconId, hash FROM groups JOIN icons ON groups.iconId = icons.id WHERE groups.id = ?",
		group.ID)
	if err != nil || row[0] == nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Group with given ID does not exist"})
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

	if iconId != defaultIconId && hash != getMD5(group.Icon) {
		execSql(dbCon, "UPDATE groups SET iconId = ? WHERE id = ?", defaultIconId, group.ID)
		if isIconUnused(dbCon, iconId) {
			execSql(dbCon, "DELETE FROM icons WHERE id = ?", iconId)
		}
	}

	query := "UPDATE groups SET name = ?, description = ? WHERE id = ?"
	queryArgs := []interface{}{group.Name, group.Description, group.ID}
	if group.Icon != "" {
		newIconId := insertIcon(dbCon, Flow{Icon: group.Icon}, nil)
		query = "UPDATE groups SET name = ?, description = ?, iconId = ? WHERE id = ?"
		queryArgs = []interface{}{group.Name, group.Description, newIconId, group.ID}
	}

	execSql(dbCon, query, queryArgs...)
	c.JSON(http.StatusOK, group)
}

// deleteGroup godoc
// @Summary      Delete a group
// @Description  Delete a group; its flows become ungrouped
// @Tags         Groups
// @Param id path string true "Group id"
// @Produce      json
// @Success      200  {object}  main.HTTPResponse
// @Failure      500  {object}  main.HTTPError
// @Router       /groups/{id} [delete]
func deleteGroup(c *gin.Context) {
	groupId := c.Param("id")
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	row, err := getSqlRow(dbCon, "SELECT id, iconId FROM groups WHERE id = ?", groupId)
	if err != nil || row[0] == nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Group with given ID does not exist"})
		return
	}

	iconId, ok := row[1].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Database error"})
		return
	}

	execSql(dbCon, "UPDATE flows SET groupId = NULL WHERE groupId = ?", groupId)
	execSql(dbCon, "DELETE FROM groups WHERE id = ?", groupId)
	if iconId != defaultIconId && isIconUnused(dbCon, iconId) {
		execSql(dbCon, "DELETE FROM icons WHERE id = ?", iconId)
	}

	c.JSON(http.StatusOK, HTTPResponse{Ok: "Group deleted"})
}

// getGroupFlows godoc
// @Summary      Get flows within a group
// @Description  Get all flows belonging to the specified group
// @Tags         Groups
// @Param id path string true "Group id"
// @Produce      json
// @Success      200  {array}   main.Flow
// @Failure      500  {object}  main.HTTPError
// @Router       /groups/{id}/flows [get]
func getGroupFlows(c *gin.Context) {
	groupId := c.Param("id")
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	query := `
SELECT
	json_group_array(json_object('id', id, 'name', name, 'description', description, 'amount', amount, 'icon', data, 'groupId', groupId)) AS json_result
FROM (
	SELECT
		f.id, f.name, f.description, f.amount, icons.data, f.groupId
	FROM flows f
	LEFT JOIN icons ON f.iconId = icons.id
	WHERE f.groupId = ?
	ORDER BY f.amount DESC, f.name
);
`
	var jsonResult []byte
	err := dbCon.QueryRow(query, groupId).Scan(&jsonResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query group flows"})
		return
	}

	if !json.Valid(jsonResult) {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Invalid data returned from database"})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", jsonResult)
}
