package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getUsers godoc
// @Summary      Get all users
// @Description  Get all users as a JSON array
// @Tags         Users
// @Produce      json
// @Success      200  {array}   main.User
// @Failure      500  {object}  main.HTTPError
// @Router       /users [get]
func getUsers(c *gin.Context) {
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	rows, err := dbCon.Query("SELECT id, name FROM users ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query users"})
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to read user data"})
			return
		}
		users = append(users, User{ID: id, Name: name})
	}
	c.JSON(http.StatusOK, users)
}

// createUser godoc
// @Summary      Create a new user
// @Description  Create a new user and switch to them
// @Tags         Users
// @Param request body   main.User true "User object (id is ignored)"
// @Produce      json
// @Success      201  {object}  main.User
// @Failure      500  {object}  main.HTTPError
// @Router       /users [post]
func createUser(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}
	if len(user.Name) == 0 {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Please provide a name"})
		return
	}

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	user.ID = uuid.New().String()
	execSql(dbCon, "INSERT INTO users (id, name) VALUES (?, ?)", user.ID, user.Name)
	execSql(dbCon, "UPDATE settings SET currentUserId = ? WHERE id = 1", user.ID)
	c.JSON(http.StatusCreated, user)
}

// updateUser godoc
// @Summary      Update a user
// @Description  Update the name of a user
// @Tags         Users
// @Param id      path  string    true "User id"
// @Param request body  main.User true "User object with updated name"
// @Produce      json
// @Success      200  {object}  main.HTTPResponse
// @Failure      500  {object}  main.HTTPError
// @Router       /users/{id} [patch]
func updateUser(c *gin.Context) {
	userId := c.Param("id")

	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}
	if len(user.Name) == 0 {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Please provide a name"})
		return
	}

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	execSql(dbCon, "UPDATE users SET name = ? WHERE id = ?", user.Name, userId)
	c.JSON(http.StatusOK, HTTPResponse{Ok: "User updated"})
}

// deleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user and all their flows. If the deleted user is current, switches to the next available user or creates a new default user.
// @Tags         Users
// @Param id path string true "User id"
// @Produce      json
// @Success      200  {object}  main.HTTPResponse
// @Failure      500  {object}  main.HTTPError
// @Router       /users/{id} [delete]
func deleteUser(c *gin.Context) {
	userId := c.Param("id")
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	currentId, err := getCurrentUserId(dbCon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to get current user"})
		return
	}

	execSql(dbCon, "DELETE FROM flows WHERE userId = ?", userId)
	execSql(dbCon,
		"DELETE FROM icons WHERE id != ? AND id NOT IN (SELECT DISTINCT iconId FROM flows)",
		defaultIconId)
	execSql(dbCon, "DELETE FROM users WHERE id = ?", userId)

	if currentId == userId {
		row, err := getSqlRow(dbCon, "SELECT id FROM users ORDER BY name LIMIT 1", nil)
		if err == nil && row[0] != nil {
			execSql(dbCon, "UPDATE settings SET currentUserId = ? WHERE id = 1", row[0].(string))
		} else {
			newId := uuid.New().String()
			execSql(dbCon, "INSERT INTO users (id, name) VALUES (?, 'thrifty')", newId)
			execSql(dbCon, "UPDATE settings SET currentUserId = ? WHERE id = 1", newId)
		}
	}

	c.JSON(http.StatusOK, HTTPResponse{Ok: "User deleted"})
}
