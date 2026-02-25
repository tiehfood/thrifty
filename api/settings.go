package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getSettings godoc
// @Summary      Get settings
// @Description  Get application settings including the currently active user
// @Tags         Settings
// @Produce      json
// @Success      200  {object}  main.Settings
// @Failure      500  {object}  main.HTTPError
// @Router       /settings [get]
func getSettings(c *gin.Context) {
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	var multiUserEnabled int64
	var userId, userName, numberFormat string
	err := dbCon.QueryRow(`
SELECT s.multiUserEnabled, u.id, u.name, s.numberFormat
FROM settings s
JOIN users u ON s.currentUserId = u.id
WHERE s.id = 1`).Scan(&multiUserEnabled, &userId, &userName, &numberFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query settings"})
		return
	}

	c.JSON(http.StatusOK, Settings{
		MultiUserEnabled: multiUserEnabled == 1,
		CurrentUser:      &User{ID: userId, Name: userName},
		NumberFormat:     numberFormat,
	})
}

// updateSettings godoc
// @Summary      Update settings
// @Description  Update application settings. Provide currentUserId to switch the active user. Provide multiUserEnabled to toggle multi-user mode (disabling deletes all other users and their data).
// @Tags         Settings
// @Param request body   main.SettingsPatch true "Fields to update (all optional)"
// @Produce      json
// @Success      200  {object}  main.HTTPResponse
// @Failure      500  {object}  main.HTTPError
// @Router       /settings [patch]
func updateSettings(c *gin.Context) {
	var patch SettingsPatch
	if err := c.BindJSON(&patch); err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Wrong data format"})
		return
	}

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	if patch.CurrentUserId != nil {
		execSql(dbCon, "UPDATE settings SET currentUserId = ? WHERE id = 1", *patch.CurrentUserId)
	}

	if patch.NumberFormat != nil {
		execSql(dbCon, "UPDATE settings SET numberFormat = ? WHERE id = 1", *patch.NumberFormat)
	}

	if patch.MultiUserEnabled != nil {
		var currentEnabled int64
		if err := dbCon.QueryRow("SELECT multiUserEnabled FROM settings WHERE id = 1").Scan(&currentEnabled); err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query settings"})
			return
		}

		if !*patch.MultiUserEnabled && currentEnabled == 1 {
			currentUserId, err := getCurrentUserId(dbCon)
			if err != nil {
				c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to get current user"})
				return
			}
			execSql(dbCon, "DELETE FROM flows WHERE userId != ?", currentUserId)
			execSql(dbCon,
				"DELETE FROM icons WHERE id != ? AND id NOT IN (SELECT DISTINCT iconId FROM flows)",
				defaultIconId)
			execSql(dbCon, "DELETE FROM users WHERE id != ?", currentUserId)
			execSql(dbCon, "UPDATE users SET name = 'thrifty' WHERE id = ?", currentUserId)
		}

		multiUserEnabled := 0
		if *patch.MultiUserEnabled {
			multiUserEnabled = 1
		}
		execSql(dbCon, "UPDATE settings SET multiUserEnabled = ? WHERE id = 1", multiUserEnabled)
	}

	c.JSON(http.StatusOK, HTTPResponse{Ok: "Settings updated"})
}
