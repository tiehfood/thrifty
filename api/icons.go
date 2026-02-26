package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getIcons godoc
// @Summary      Get all icons
// @Description  Get all non-default icons as a JSON array, including whether each icon is in use
// @Tags         Icons
// @Produce      json
// @Success      200  {array}   main.Icon
// @Failure      500  {object}  main.HTTPError
// @Router       /icons [get]
func getIcons(c *gin.Context) {
	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	rows, err := dbCon.Query(`
		SELECT id, data, EXISTS(SELECT 1 FROM flows WHERE iconId = icons.id) AS isUsed
		FROM icons
		WHERE id != '00000000-0000-0000-0000-000000000000'
		ORDER BY rowid`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to query icons"})
		return
	}
	defer rows.Close()

	icons := []Icon{}
	for rows.Next() {
		var icon Icon
		var isUsed int
		if err := rows.Scan(&icon.ID, &icon.Data, &isUsed); err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to read icon data"})
			return
		}
		icon.IsUsed = isUsed != 0
		icons = append(icons, icon)
	}
	c.JSON(http.StatusOK, icons)
}

// addIcons godoc
// @Summary      Add one or more icons
// @Description  Add icons (deduplicates by content hash). Accepts an array; send a single-element array for one icon.
// @Tags         Icons
// @Param request body   []main.IconRequest true "Array of icon data URIs"
// @Produce      json
// @Success      201  {array}   main.Icon
// @Failure      400  {object}  main.HTTPError
// @Failure      500  {object}  main.HTTPError
// @Router       /icons [post]
func addIcons(c *gin.Context) {
	var reqs []IconRequest
	if err := c.BindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, HTTPError{Error: "Wrong data format"})
		return
	}
	if len(reqs) == 0 {
		c.JSON(http.StatusBadRequest, HTTPError{Error: "Please provide at least one icon"})
		return
	}

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	result := []Icon{}
	for _, req := range reqs {
		if len(req.Data) == 0 {
			continue
		}
		id := insertIcon(dbCon, Flow{Icon: req.Data}, nil)

		rows, err := dbCon.Query(`
			SELECT id, data, EXISTS(SELECT 1 FROM flows WHERE iconId = icons.id) AS isUsed
			FROM icons WHERE id = ?`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to retrieve icon"})
			return
		}
		if rows.Next() {
			var icon Icon
			var isUsed int
			if err := rows.Scan(&icon.ID, &icon.Data, &isUsed); err != nil {
				rows.Close()
				c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to read icon data"})
				return
			}
			icon.IsUsed = isUsed != 0
			result = append(result, icon)
		}
		rows.Close()
	}
	c.JSON(http.StatusCreated, result)
}

// deleteIcon godoc
// @Summary      Delete an icon
// @Description  Delete an icon by ID. Fails if icon is in use or is the default icon.
// @Tags         Icons
// @Param id path string true "Icon id"
// @Produce      json
// @Success      200  {object}  main.HTTPResponse
// @Failure      400  {object}  main.HTTPError
// @Failure      500  {object}  main.HTTPError
// @Router       /icons/{id} [delete]
func deleteIcon(c *gin.Context) {
	id := c.Param("id")

	if id == defaultIconId {
		c.JSON(http.StatusBadRequest, HTTPError{Error: "Cannot delete default icon"})
		return
	}

	dbCon := getDbConnection(c)
	if dbCon == nil {
		return
	}

	var count int
	err := dbCon.QueryRow("SELECT count(*) FROM flows WHERE iconId = ?", id).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HTTPError{Error: "Failed to check icon usage"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, HTTPError{Error: "Icon is in use"})
		return
	}

	execSql(dbCon, "DELETE FROM icons WHERE id = ?", id)
	c.JSON(http.StatusOK, HTTPResponse{Ok: "Icon deleted"})
}
